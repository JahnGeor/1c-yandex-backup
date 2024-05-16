package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type BackupLocal struct {
	BackupDir string
	Retention int
	Expired   time.Duration
}

func (b *BackupLocal) CreateBackup(path string) (string, error) {

	file, err := os.OpenFile(path, os.O_RDONLY, 0666)

	if err != nil {
		return "", fmt.Errorf("unable to open source file %s", path)
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		return "", fmt.Errorf("unable to stat source file %s", path)
	}

	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("source file %s is empty", path)
	}

	backupFileName := fmt.Sprintf("%s.%s", time.Now().Format("20060102150405"), filepath.Base(fileInfo.Name()))

	backupFilePath := filepath.Join(b.BackupDir, backupFileName)

	backupFile, err := os.OpenFile(backupFilePath, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", fmt.Errorf("unable to create backup file %s", backupFilePath)
	}

	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)

	if err != nil {
		return "", fmt.Errorf("unable to copy source file %s to backup file %s", path, backupFilePath)
	}

	return backupFileName, nil
}

func (b *BackupLocal) EraseBackups() error {
	if b.Retention == 0 {
		return fmt.Errorf("retention is not set")
	}

	if b.Expired == 0 {
		return fmt.Errorf("expired is not set")
	}

	files, err := os.ReadDir(b.BackupDir)

	if err != nil {
		return fmt.Errorf("unable to read backup directory %s", b.BackupDir)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			return fmt.Errorf("unable to get file info %s", file.Name())
		}

		if fileInfo.ModTime().Add(b.Expired).Before(time.Now()) {
			err = os.Remove(filepath.Join(b.BackupDir, file.Name()))
			if err != nil {
				return fmt.Errorf("unable to remove file %s", file.Name())
			}
		}
	}

	return nil
}
