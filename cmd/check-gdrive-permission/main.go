package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
)

func main() {
	tokenFlag := flag.String("token", "", "Refresh token to check (optional, uses ACCOUNTING_GOOGLE_REFRESH_TOKEN if not provided)")
	dirFlag := flag.String("dir", "", "Directory ID to check (optional, uses INVOICE_DIR_ID if not provided)")
	listRootFlag := flag.Bool("list-root", false, "List root directories accessible by the token")
	flag.Parse()

	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger(cfg.LogLevel)

	log.Info("Checking Google Drive permission")

	v, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
	}

	if v != nil {
		cfg = config.Generate(v)
	}

	// Determine which token to use
	refreshToken := *tokenFlag
	if refreshToken == "" {
		refreshToken = cfg.Google.AccountingGoogleRefreshToken
	}
	if refreshToken == "" {
		log.Error(nil, "No refresh token provided. Use -token flag or set ACCOUNTING_GOOGLE_REFRESH_TOKEN")
		os.Exit(1)
	}
	log.Debugf("Using refresh token: %s...%s", refreshToken[:10], refreshToken[len(refreshToken)-5:])

	// Determine which directory to check
	dirID := *dirFlag
	if dirID == "" {
		dirID = cfg.Invoice.DirID
	}
	if dirID != "" {
		log.Debugf("Target directory ID: %s", dirID)
	}

	// Create OAuth2 config
	driveConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{googledrive.FullDriveAccessScope},
	}

	// Create token from refresh token
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Get new token using refresh token
	log.Debug("Refreshing access token...")
	tokenSource := driveConfig.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Errorf(err, "failed to refresh token - check if ACCOUNTING_GOOGLE_REFRESH_TOKEN is valid")
		os.Exit(1)
	}
	log.Debugf("Access token obtained, expires at: %v", newToken.Expiry)

	// Create Drive service
	log.Debug("Creating Google Drive service...")
	client := driveConfig.Client(context.Background(), newToken)
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Errorf(err, "failed to create Drive service")
		os.Exit(1)
	}

	// List root directories if requested
	if *listRootFlag {
		log.Info("Listing root directories accessible by this token...")
		result, err := service.Files.List().
			Q("'root' in parents and mimeType='application/vnd.google-apps.folder'").
			Fields("files(id, name, mimeType, owners)").
			PageSize(50).
			Do()
		if err != nil {
			log.Errorf(err, "failed to list root directories")
			os.Exit(1)
		}

		log.Infof("Found %d root directories:", len(result.Files))
		for _, f := range result.Files {
			ownerInfo := ""
			if len(f.Owners) > 0 {
				ownerInfo = fmt.Sprintf(" (owner: %s)", f.Owners[0].EmailAddress)
			}
			log.Infof("  - %s [%s]%s", f.Name, f.Id, ownerInfo)
		}

		// Also list shared drives
		log.Debug("Listing shared drives...")
		drives, err := service.Drives.List().PageSize(50).Do()
		if err != nil {
			log.Debugf("Failed to list shared drives: %v", err)
		} else if len(drives.Drives) > 0 {
			log.Infof("Found %d shared drives:", len(drives.Drives))
			for _, d := range drives.Drives {
				log.Infof("  - %s [%s]", d.Name, d.Id)
			}
		}

		os.Exit(0)
	}

	// Check specific directory if provided
	if dirID == "" {
		log.Error(nil, "No directory ID provided. Use -dir flag, set INVOICE_DIR_ID, or use -list-root")
		os.Exit(1)
	}

	// Get token info to show which account we're using
	about, err := service.About.Get().Fields("user").Do()
	if err != nil {
		log.Debugf("Failed to get account info: %v", err)
	} else {
		log.Infof("Authenticated as: %s (%s)", about.User.DisplayName, about.User.EmailAddress)
	}

	// Check token scopes
	log.Debugf("Token scopes: %v", newToken.Extra("scope"))

	// Try to get the directory metadata
	log.Debugf("Checking access to directory: %s", dirID)
	file, err := service.Files.Get(dirID).Fields("id, name, mimeType, owners").SupportsAllDrives(true).Do()
	if err != nil {
		log.Errorf(err, "failed to access directory via Files.Get")

		// Try searching in sharedWithMe
		log.Debug("Trying to find in sharedWithMe...")
		sharedResult, searchErr := service.Files.List().
			Q("sharedWithMe=true and mimeType='application/vnd.google-apps.folder'").
			Fields("files(id, name, owners)").
			PageSize(50).
			Do()
		if searchErr != nil {
			log.Errorf(searchErr, "failed to search sharedWithMe")
		} else {
			log.Infof("Found %d shared folders:", len(sharedResult.Files))
			for _, f := range sharedResult.Files {
				marker := ""
				if f.Id == dirID {
					marker = " <-- TARGET"
				}
				ownerInfo := ""
				if len(f.Owners) > 0 {
					ownerInfo = fmt.Sprintf(" (owner: %s)", f.Owners[0].EmailAddress)
				}
				log.Infof("  - %s [%s]%s%s", f.Name, f.Id, ownerInfo, marker)
			}
		}
		os.Exit(1)
	}

	log.Infof("Successfully accessed directory:")
	log.Infof("  ID: %s", file.Id)
	log.Infof("  Name: %s", file.Name)
	log.Infof("  Type: %s", file.MimeType)
	if len(file.Owners) > 0 {
		log.Infof("  Owner: %s (%s)", file.Owners[0].DisplayName, file.Owners[0].EmailAddress)
	}

	// Try to list files in the directory
	log.Debug("Listing files in directory...")
	result, err := service.Files.List().
		Q(fmt.Sprintf("'%s' in parents", dirID)).
		Fields("files(id, name, mimeType)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		PageSize(10).
		Do()
	if err != nil {
		log.Errorf(err, "failed to list files in directory - read permission may be limited")
		os.Exit(1)
	}

	log.Infof("Found %d items in directory (showing up to 10):", len(result.Files))
	for _, f := range result.Files {
		log.Infof("  - %s (%s)", f.Name, f.MimeType)
	}

	log.Info("")
	log.Info("Permission check PASSED")
}
