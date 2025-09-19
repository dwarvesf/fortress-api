package parquet

import (
	"context"
	"time"
)

// SyncStatus represents the status of parquet file synchronization
type SyncStatus struct {
	LastSyncTime    time.Time `json:"last_sync_time"`
	IsLocalFileValid bool     `json:"is_local_file_valid"`
	LocalFileSize    int64    `json:"local_file_size"`
	RemoteETag       string   `json:"remote_etag,omitempty"`
	RemoteLastMod    string   `json:"remote_last_modified,omitempty"`
	SyncInProgress   bool     `json:"sync_in_progress"`
	LastError        string   `json:"last_error,omitempty"`
}

// ISyncService defines the interface for parquet file synchronization
type ISyncService interface {
	// StartBackgroundSync starts the background synchronization process
	StartBackgroundSync(ctx context.Context) error
	
	// StopBackgroundSync stops the background synchronization process
	StopBackgroundSync() error
	
	// SyncNow forces an immediate synchronization
	SyncNow(ctx context.Context) error
	
	// GetLocalFilePath returns the path to the local parquet file
	GetLocalFilePath() string
	
	// IsLocalFileReady checks if the local file is ready for use
	IsLocalFileReady() bool
	
	// GetSyncStatus returns the current synchronization status
	GetSyncStatus() SyncStatus
	
	// GetRemoteURL returns the configured remote URL
	GetRemoteURL() string
}