package remote

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	entity "yd_backup/internal/models"
	"yd_backup/pkg/yandex/disk"
	"yd_backup/pkg/yandex/disk/models"
)

type BackupRemote struct {
	disk    *disk.YandexDisk
	setting entity.Setting
}

func (b *BackupRemote) RemoveBackup() ([]string, error) {
	var result []string

	resource, err := b.disk.GetResource(models.Params{Path: b.setting.Yandex.Dir})

	if err != nil {
		return nil, err
	}

	for _, resource := range resource.Embedded.Items {

		if resource.Created.Local().Add(b.setting.Backup.Expired.Duration).Before(time.Now()) {

			var params models.Params

			params.Path = resource.Path
			params.Permanently = true

			if _, err := b.disk.RemoveResource(params); err != nil {
				return nil, err
			}

			result = append(result, resource.Path)
		}

	}

	return result, nil

}

func NewBackupRemote(setting entity.Setting) *BackupRemote {
	return &BackupRemote{
		setting: setting,
		disk:    disk.NewBackupYandex(setting.Yandex.Token, setting.Yandex.Timeout.Duration),
	}
}

func (b *BackupRemote) CreateFolder(dir string) error {
	return nil
}

func (b *BackupRemote) UploadBackup(backupPath string) error {
	var params models.Params

	var remoteFileName = filepath.Base(backupPath)

	if !b.setting.Yandex.Extension {
		remoteFileName = strings.TrimSuffix(remoteFileName, filepath.Ext(backupPath))
	}

	remotePath := fmt.Sprintf("%s/%s", b.setting.Yandex.Dir, remoteFileName)

	params.Path = remotePath
	params.Overwrite = true

	link, err := b.disk.CreateLink(params)
	if err != nil {
		return err
	}

	return b.disk.UploadFile(link, backupPath)
}

func (b *BackupRemote) EraseBackup() error {
	return nil
}
