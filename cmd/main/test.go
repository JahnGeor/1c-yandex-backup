package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"yd_backup/internal/models"
	"yd_backup/internal/repo/local"
	"yd_backup/internal/repo/remote"
	"yd_backup/internal/usecase"
)

func main() {
	logger, err := initLog()

	if err != nil {
		log.Fatal(err)
	}

	defer logger.Sync()

	setting, err := readSetting()

	if err != nil {
		logger.Info("unable to read config file", zap.Error(err))
	}

	logger.Debug("config", zap.Any("config", setting))

	localBackup := local.NewBackupLocal(setting)
	remoteBackup := remote.NewBackupRemote(setting)

	service := usecase.NewBackupService(setting, remoteBackup, localBackup, logger)

	service.BackupAll()

	service.EraseBackup()

}

func initLog() (*zap.Logger, error) {
	fileLog, err := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	fileLogger := zapcore.AddSync(fileLog)
	consoleLogger := zapcore.AddSync(os.Stdout)

	logger, err := zap.NewProduction()

	if err != nil {
		return logger, err
	}

	productionConfig := zap.NewDevelopmentEncoderConfig()
	productionConfig.TimeKey = "time"
	productionConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	productionConfig.CallerKey = "caller"

	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	developmentConfig := zap.NewDevelopmentEncoderConfig()
	developmentConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	developmentConfig.TimeKey = "time"
	developmentConfig.CallerKey = "caller"

	consoleEncoder := zapcore.NewConsoleEncoder(developmentConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileLogger, zapcore.DebugLevel),
		zapcore.NewCore(consoleEncoder, consoleLogger, zapcore.DebugLevel),
	)

	l := zap.New(core, zap.WithCaller(true))

	return l, nil
}

func readSetting() (models.Setting, error) {
	fileSetting, err := os.Open("./config/config.json")

	if err != nil {
		return models.Setting{}, err
	}

	decoder := json.NewDecoder(fileSetting)

	var setting models.Setting
	err = decoder.Decode(&setting)

	if err != nil {
		return models.Setting{}, err
	}

	return setting, nil
}
