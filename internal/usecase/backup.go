package usecase

import (
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

func (r *Result) Backup(path string) error {

}

type BackupService struct {
	setting models.Settings
	remote  Backup
}

func (b *BackupService) BackupAll() {
	wg := &sync.WaitGroup{}

	result := &Result{
		success: 0,
	}

	for _, path := range b.setting.Databases.Paths {
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

	logger.Info("Backup success: %d/%d", result.getSuccess(), len(b.setting.Databases.Paths))
}

func (b *BackupService) Backup(path string) error {
	return b.remote.Backup(path)
}
