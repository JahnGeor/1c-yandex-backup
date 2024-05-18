package usecase

import (
	"fmt"
	"github.com/google/logger"
	"sync"
	"yd_backup/internal/models"
)

type LocalBackup interface {
	CreateBackup(path string) (string, error)
	EraseBackups() error
}

type RemoteBackup interface {
	UploadBackup(path string) error
	RemoveBackup(path string) error
}

type Result struct {
	sync.Mutex
	success int
}

func (r *Result) getSuccess() int {
	r.Lock()
	defer r.Unlock()
	return r.success
}

func (r *Result) setSuccess() {
	r.Lock()
	defer r.Unlock()
	r.success++
}

type BackupService struct {
	paths  []string
	remote RemoteBackup
	local  LocalBackup
}

func NewBackupService(setting models.Settings, remote RemoteBackup, local LocalBackup) *BackupService {
	return &BackupService{
		paths:  setting.Databases.Paths,
		remote: remote,
		local:  local,
	}
}

func (b *BackupService) BackupAll() {
	wg := &sync.WaitGroup{}

	result := &Result{
		success: 0,
	}

	for _, path := range b.paths {
		wg.Add(1)

		currentPath := path

		go func() {
			if err := b.Backup(currentPath); err != nil {
				logger.Errorf("Backup %s failed: %v", currentPath, err)
			} else {
				logger.Infof("Backup %s success", currentPath)
				result.setSuccess()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	logger.Info("Backup success: %d/%d", result.getSuccess(), len(b.paths))
}

func (b *BackupService) Backup(path string) error {
	//TODO: Создать локальную копию
	backupPath, err := b.local.CreateBackup(path)
	if err != nil {
		return fmt.Errorf("unable to create local backup: %v", err)
	}
	//TODO: Создать удаленную копию

	err = b.remote.UploadBackup(backupPath)

	if err != nil {
		return fmt.Errorf("unable to upload backup to remote disk: %v", err)
	}

	return nil
}

func (b *BackupService) EraseBackups() error {
	//TODO: Удалить бэкапы локально
	//TODO: Удалить бэкапы на удаленном сервере

	return nil
}
