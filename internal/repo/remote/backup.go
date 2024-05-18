package remote

import (
	"time"
	"yd_backup/pkg/yandex/disk"
)

type BackupRemote struct {
	disk *disk.YandexDisk
}

func (b *BackupRemote) RemoveBackup(path string) error {
	return nil
}

func NewBackupRemote(token string, timeout time.Duration) *BackupRemote {
	return &BackupRemote{
		disk: disk.NewBackupYandex(token, timeout),
	}
}

func (b *BackupRemote) UploadBackup(path string) error {
	return nil
}
