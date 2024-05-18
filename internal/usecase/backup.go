package usecase

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"yd_backup/internal/models"
)

type LocalBackup interface {
	CreateBackup(file models.Files) (string, error)
	EraseBackup() ([]string, error)
}

type RemoteBackup interface {
	CreateFolder(dir string) error
	UploadBackup(backupPath string) error
	RemoveBackup() ([]string, error)
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
	setting models.Setting
	remote  RemoteBackup
	local   LocalBackup
	logger  *zap.Logger
}

func NewBackupService(setting models.Setting, remote RemoteBackup, local LocalBackup, logger *zap.Logger) *BackupService {
	return &BackupService{
		setting: setting,
		remote:  remote,
		local:   local,
		logger:  logger,
	}
}

func (b *BackupService) BackupAll() {

	err := b.remote.CreateFolder(b.setting.Yandex.Dir)

	if err != nil {
		b.logger.With(zap.Error(err)).Error("Unable to create remote folder")
		return
	}

	wg := &sync.WaitGroup{}

	result := &Result{
		success: 0,
	}

	for _, path := range b.setting.Files {
		wg.Add(1)

		currentPath := path

		go func() {
			if err := b.Backup(currentPath); err != nil {
				b.logger.With(zap.String("Path", currentPath.Path)).With(zap.Error(err)).Error("Backup failed")
			} else {
				b.logger.With(zap.String("Path", currentPath.Path)).Info("Backup success")
				result.setSuccess()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	b.logger.With(zap.String("progress", fmt.Sprintf("%d/%d", result.getSuccess(), len(b.setting.Files)))).
		Info("Backup complete")

}

func (b *BackupService) Backup(files models.Files) error {
	//TODO: Создать локальную копию
	backupPath, err := b.local.CreateBackup(files)
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

func (b *BackupService) EraseBackup() {
	paths, err := b.local.EraseBackup()
	if err != nil {
		b.logger.With(zap.Error(err)).Error("unable to erase local backup: %v")
		return
	}

	b.logger.With(zap.Strings("paths", paths)).With(zap.Int("count", len(paths))).Info("Local backup erased")

	paths, err = b.remote.RemoveBackup()

	if err != nil {
		b.logger.With(zap.Error(err)).Error("unable to erase remote backup: %v")
		return
	}

	b.logger.With(zap.Strings("paths", paths)).With(zap.Int("count", len(paths))).Info("Remote backup erased")

	return
}
