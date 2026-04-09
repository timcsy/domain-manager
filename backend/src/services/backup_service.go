package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupInfo holds metadata about a backup file
type BackupInfo struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// BackupService handles database backup operations
type BackupService struct {
	dbPath    string
	backupDir string
}

// NewBackupService creates a new backup service
func NewBackupService(dbPath string) *BackupService {
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	return &BackupService{
		dbPath:    dbPath,
		backupDir: backupDir,
	}
}

// CreateBackup creates a backup of the SQLite database file
func (s *BackupService) CreateBackup() (*BackupInfo, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("backup-%s.db", timestamp)
	destPath := filepath.Join(s.backupDir, filename)

	src, err := os.Open(s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to copy database: %w", err)
	}

	log.Printf("Backup created: %s (%d bytes)", filename, written)

	return &BackupInfo{
		Filename:  filename,
		Size:      written,
		CreatedAt: time.Now(),
	}, nil
}

// ListBackups returns all available backups sorted by date (newest first)
func (s *BackupService) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "backup-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// GetBackupPath returns the full path of a backup file
func (s *BackupService) GetBackupPath(filename string) (string, error) {
	// Prevent path traversal
	clean := filepath.Base(filename)
	if clean != filename || !strings.HasPrefix(filename, "backup-") {
		return "", fmt.Errorf("invalid backup filename")
	}
	path := filepath.Join(s.backupDir, clean)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("backup not found: %s", filename)
	}
	return path, nil
}

// CleanOldBackups removes backups older than the given retention period
func (s *BackupService) CleanOldBackups(retentionDays int) (int, error) {
	backups, err := s.ListBackups()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	removed := 0

	for _, b := range backups {
		if b.CreatedAt.Before(cutoff) {
			path := filepath.Join(s.backupDir, b.Filename)
			if err := os.Remove(path); err != nil {
				log.Printf("Warning: failed to remove old backup %s: %v", b.Filename, err)
				continue
			}
			removed++
		}
	}

	if removed > 0 {
		log.Printf("Cleaned %d old backup(s) older than %d days", removed, retentionDays)
	}

	return removed, nil
}

// DeleteBackup deletes a specific backup file
func (s *BackupService) DeleteBackup(filename string) error {
	path, err := s.GetBackupPath(filename)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
