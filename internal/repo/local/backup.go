package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	entity "yd_backup/internal/models"
)

type BackupLocal struct {
	setting entity.Setting
}

func NewBackupLocal(setting entity.Setting) *BackupLocal {
	return &BackupLocal{setting: setting}
}

func (b *BackupLocal) CreateBackup(path entity.Files) (string, error) {

	file, err := os.OpenFile(path.Path, os.O_RDONLY, 0666)

	if err != nil {
		return "", fmt.Errorf("unable to open source file %s", path.Path)
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		return "", fmt.Errorf("unable to stat source file %s", path.Path)
	}

	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("source file %s is empty", path.Path)
	}

	backupFileName := fmt.Sprintf("%s_%s_%s", path.Name, time.Now().Format("20060102150405"), filepath.Base(fileInfo.Name()))

	backupFilePath := filepath.Join(b.setting.Backup.Dir, backupFileName)

	backupFile, err := os.OpenFile(backupFilePath, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", fmt.Errorf("unable to create backup file %s", backupFilePath)
	}

	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)

	if err != nil {
		return "", fmt.Errorf("unable to copy source file %s to backup file %s", path, backupFilePath)
	}

	return backupFile.Name(), nil
}

func (b *BackupLocal) EraseBackup() ([]string, error) {
	var deletedFiles []string
	if b.setting.Backup.Retention == 0 {
		return nil, fmt.Errorf("retention is not set")
	}

	if b.setting.Backup.Expired.Duration == 0 {
		return nil, fmt.Errorf("expired is not set")
	}

	files, err := os.ReadDir(b.setting.Backup.Dir)

	if err != nil {
		return nil, fmt.Errorf("unable to read backup directory %s", b.setting.Backup.Dir)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("unable to get file info %s", file.Name())
		}

		if fileInfo.ModTime().Add(b.setting.Backup.Expired.Duration).Before(time.Now()) {
			err = os.Remove(filepath.Join(b.setting.Backup.Dir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("unable to remove file %s", file.Name())
			}
			deletedFiles = append(deletedFiles, file.Name())
		}
	}

	return deletedFiles, nil
}
