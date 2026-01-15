package parquet

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type syncService struct {
	mu             sync.RWMutex
	config         config.Parquet
	logger         logger.Logger
	httpClient     *http.Client
	status         SyncStatus
	syncTicker     *time.Ticker
	stopChan       chan struct{}
	syncInProgress bool
}

// NewSyncService creates a new parquet sync service
func NewSyncService(cfg config.Parquet, logger logger.Logger) ISyncService {
	// Set defaults if not configured
	if cfg.LocalFilePath == "" {
		cfg.LocalFilePath = "/tmp/vault.parquet"
	}
	if cfg.SyncInterval == "" {
		cfg.SyncInterval = "1h"
	}
	if cfg.RemoteURL == "" {
		cfg.RemoteURL = "https://raw.githubusercontent.com/dwarvesf/memo.d.foundation/refs/heads/main/db/vault.parquet"
	}
	if cfg.QuickTimeout == "" {
		cfg.QuickTimeout = "2s"
	}
	if cfg.ExtendedTimeout == "" {
		cfg.ExtendedTimeout = "60s"
	}

	return &syncService{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Generous timeout for large file downloads
		},
		status: SyncStatus{
			IsLocalFileValid: false,
			SyncInProgress:   false,
		},
		stopChan: make(chan struct{}),
	}
}

// StartBackgroundSync starts the background synchronization process
func (s *syncService) StartBackgroundSync(ctx context.Context) error {
	if !s.config.EnableCaching {
		s.logger.Info("Parquet caching is disabled, skipping background sync")
		return nil
	}

	interval, err := time.ParseDuration(s.config.SyncInterval)
	if err != nil {
		return fmt.Errorf("invalid sync interval: %w", err)
	}

	s.logger.Debugf("Starting parquet background sync with interval: %s", interval)

	// Perform initial sync
	if err := s.SyncNow(ctx); err != nil {
		s.logger.Warnf("Initial parquet sync failed: %v", err)
	}

	// Start background ticker
	s.syncTicker = time.NewTicker(interval)
	go s.backgroundSyncLoop(ctx)

	return nil
}

// StopBackgroundSync stops the background synchronization process
func (s *syncService) StopBackgroundSync() error {
	if s.syncTicker != nil {
		s.syncTicker.Stop()
	}
	
	close(s.stopChan)
	s.logger.Info("Parquet background sync stopped")
	return nil
}

// SyncNow forces an immediate synchronization
func (s *syncService) SyncNow(ctx context.Context) error {
	s.mu.Lock()
	if s.syncInProgress {
		s.mu.Unlock()
		return fmt.Errorf("sync already in progress")
	}
	s.syncInProgress = true
	s.status.SyncInProgress = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.syncInProgress = false
		s.status.SyncInProgress = false
		s.mu.Unlock()
	}()

	s.logger.Info("Starting parquet file synchronization")
	start := time.Now()

	if err := s.performSync(ctx); err != nil {
		s.mu.Lock()
		s.status.LastError = err.Error()
		s.mu.Unlock()
		s.logger.Errorf(err, "Parquet sync failed")
		return err
	}

	s.mu.Lock()
	s.status.LastSyncTime = time.Now()
	s.status.LastError = ""
	s.status.IsLocalFileValid = s.isLocalFileValid()
	s.mu.Unlock()

	duration := time.Since(start)
	s.logger.Infof("Parquet sync completed successfully in %v", duration)
	return nil
}

// GetLocalFilePath returns the path to the local parquet file
func (s *syncService) GetLocalFilePath() string {
	return s.config.LocalFilePath
}

// IsLocalFileReady checks if the local file is ready for use
func (s *syncService) IsLocalFileReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.IsLocalFileValid && !s.status.SyncInProgress
}

// GetSyncStatus returns the current synchronization status
func (s *syncService) GetSyncStatus() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Update file size if file exists
	if info, err := os.Stat(s.config.LocalFilePath); err == nil {
		s.status.LocalFileSize = info.Size()
	}
	
	return s.status
}

// GetRemoteURL returns the configured remote URL
func (s *syncService) GetRemoteURL() string {
	return s.config.RemoteURL
}

// backgroundSyncLoop runs the background synchronization loop
func (s *syncService) backgroundSyncLoop(ctx context.Context) {
	for {
		select {
		case <-s.syncTicker.C:
			if err := s.SyncNow(ctx); err != nil {
				s.logger.Warnf("Background parquet sync failed: %v", err)
			}
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// performSync performs the actual file synchronization
func (s *syncService) performSync(ctx context.Context) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(s.config.LocalFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if we need to download based on ETag/Last-Modified
	shouldDownload, err := s.shouldDownloadFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to check remote file: %w", err)
	}

	if !shouldDownload {
		s.logger.Info("Remote parquet file unchanged, skipping download")
		return nil
	}

	// Download to temporary file first (atomic operation)
	tempFile := s.config.LocalFilePath + ".tmp"
	if err := s.downloadFile(ctx, tempFile); err != nil {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			s.logger.Warnf("Failed to clean up temp file %s: %v", tempFile, removeErr)
		}
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.config.LocalFilePath); err != nil {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			s.logger.Warnf("Failed to clean up temp file %s: %v", tempFile, removeErr)
		}
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	s.logger.Infof("Successfully downloaded parquet file to %s", s.config.LocalFilePath)
	return nil
}

// shouldDownloadFile checks if we should download the file based on ETag/Last-Modified
func (s *syncService) shouldDownloadFile(ctx context.Context) (bool, error) {
	// If local file doesn't exist, we definitely need to download
	if !s.isLocalFileValid() {
		return true, nil
	}

	// Make HEAD request to get remote file metadata
	req, err := http.NewRequestWithContext(ctx, "HEAD", s.config.RemoteURL, nil)
	if err != nil {
		return false, err
	}

	// Add If-None-Match header if we have ETag
	s.mu.RLock()
	if s.status.RemoteETag != "" {
		req.Header.Set("If-None-Match", s.status.RemoteETag)
	}
	if s.status.RemoteLastMod != "" {
		req.Header.Set("If-Modified-Since", s.status.RemoteLastMod)
	}
	s.mu.RUnlock()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// If we get 304 Not Modified, no need to download
	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("remote server returned status: %s", resp.Status)
	}

	// Update our stored metadata
	s.mu.Lock()
	s.status.RemoteETag = resp.Header.Get("ETag")
	s.status.RemoteLastMod = resp.Header.Get("Last-Modified")
	s.mu.Unlock()

	return true, nil
}

// downloadFile downloads the file from the remote URL
func (s *syncService) downloadFile(ctx context.Context, filepath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.config.RemoteURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status: %s", resp.Status)
	}

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy with progress logging
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	s.logger.Infof("Downloaded %d bytes to %s", written, filepath)
	return nil
}

// isLocalFileValid checks if the local file exists and is readable
func (s *syncService) isLocalFileValid() bool {
	info, err := os.Stat(s.config.LocalFilePath)
	if err != nil {
		return false
	}
	
	// File should be at least 1MB (reasonable minimum for a parquet file)
	return info.Size() > 1024*1024
}