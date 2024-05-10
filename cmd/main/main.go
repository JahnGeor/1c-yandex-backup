package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/google/logger"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var logsPath = fmt.Sprintf("./logs/%s.log", time.Now().Format("2006_01_02"))

func main() {
	// Args
	verbose := flag.Bool("verbose", false, "print info level logs to stdout")
	dbPath := flag.String("db_path", "none", "path to database")
	token := flag.String("token", "none", "token to yandex disk")
	backupDir := flag.String("backup_dir", "none", "backup dir")
	count := flag.Int("count", 2, "count of the yandex disk backups")
	timeErasing := flag.Duration("erase_interval", time.Hour*24*3, "time duration of erasing old backups")
	timeout := flag.Duration("timeout", time.Hour*2, "timeout for http-request")
	flag.Parse()
	// Args

	// Logger
	lf, err := os.OpenFile(logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Errorf("Критическая ошибка запуска приложения: %v", err)
		os.Exit(1)
	}

	l := logger.Init("Logger", *verbose, true, lf)

	defer func() {
		logger.Infof("Завершение работы приложения\n\n")
		l.Close()
		lf.Close()
	}()

	// Logger
	logger.Infof("Начало работы приложения")
	if *dbPath == "none" || *token == "none" || *backupDir == "none" {
		logger.Errorf("Не указаны ключевые параметры запуска")
		runtime.Goexit()
	}

	size, path, err := makeBackup(*dbPath, *backupDir)
	if err != nil {
		logger.Errorf("[ЭТАП 1] Ошибка при выполнении операции бэкапирования: %s", err.Error())
		runtime.Goexit()
	}
	logger.Infof("[ЭТАП 1] Создан бэкап, размер: %f MB, путь хранения: %s", float32(size)/(1<<20), path)
	err = eraseBackup(*backupDir, *timeErasing)
	if err != nil {
		logger.Errorf("[ЭТАП 2] Ошибка при выполнении операции очистки старых бэкапов: %s", err.Error())
		runtime.Goexit()
	}
	logger.Infof("[ЭТАП 2] Произведена очистка бэкапов")

	yandexDir := "backup/"
	backupFile := time.Now().Format("2006-01-02_15-04-05") + ".1CD"

	err = createFolder(*token, yandexDir, *timeout)
	if err != nil {
		logger.Errorf("[ЭТАП 3] Ошибка в процессе создания папки для бэкапа на Yandex Disk %s", err.Error())
		runtime.Goexit()
	}
	logger.Infof("[ЭТАП 3] Создана директория для бэкапов")

	err = uploadFile(path, yandexDir+backupFile, *token, *timeout)

	if err != nil {
		logger.Errorf("[ЭТАП 4] Ошибка в процессе загрузки бэкапа на Yandex Disk %s", err.Error())
		runtime.Goexit()
	}

	logger.Infof("[ЭТАП 4] Произведена загрузка бэкапов")

	if err = removeOldBackupsOnServer(*token, yandexDir, *count, *timeout); err != nil {
		logger.Errorf("[ЭТАП 5] Ошибка в процессе очистки бэкапов на Yandex Disk %s", err.Error())
		runtime.Goexit()
	}
	logger.Infof("[ЭТАП 5] Произведена очистка бэкапов на Яндекс Диске")

}

// Модуль бэкапирования
func makeBackup(dbPath, dbBackupDir string) (int64, string, error) {
	fs, err := os.OpenFile(dbPath, os.O_RDONLY, 0660)
	defer fs.Close()
	if err != nil {
		return 0, "", err
	}
	extension := filepath.Ext(dbPath)
	fileName := fmt.Sprintf("%s/%s%s", dbBackupDir, time.Now().Format("backup_2006_01_02_15_04_05"), extension)

	fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0660)
	defer fd.Close()
	if err != nil {
		return 0, "", err
	}

	written, err := io.Copy(fd, fs)

	if err != nil {
		return 0, "", err
	}

	return written, fileName, nil

}

func eraseBackup(backupDir string, duration time.Duration) error {
	e := filepath.Walk(backupDir, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() && time.Now().Sub(info.ModTime()) >= duration {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return e
}

// Модуль взаимодействия с Yandex Disk

var ynxUrl = "https://cloud-api.yandex.net/v1/disk"

type Resource struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Created    string `json:"created"`
	ResourceId string `json:"resource_id"`
	Type       string `json:"type"`
	MimeType   string `json:"mime_type"`
	Embedded   struct {
		Items []Resource `json:"items"`
		Path  string     `json:"path"`
	} `json:"_embedded"`
}

func apiRequest(path, token, method string, timeout time.Duration) (*http.Response, error) {
	client := http.Client{Timeout: timeout}
	url := fmt.Sprintf("%s/%s", ynxUrl, path)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token))
	return client.Do(req)
}

// создание директории
func createFolder(token, path string, timeout time.Duration) error {
	_, err := apiRequest(fmt.Sprintf("resources?path=%s", path), token, "PUT", timeout)
	return err
}

// загрузка файла
func uploadFile(localPath, remotePath, token string, timeout time.Duration) error {
	// функция получения url для загрузки файла
	getUploadUrl := func(path string) (string, error) {
		res, err := apiRequest(fmt.Sprintf("resources/upload?path=%s&overwrite=true", path), token, "GET", timeout)
		if err != nil {
			return "", err
		}
		var resultJson struct {
			Href string `json:"href"`
		}
		err = json.NewDecoder(res.Body).Decode(&resultJson)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		return resultJson.Href, err

	}

	// читаем локальный файл с диска
	data, err := os.Open(localPath)
	if err != nil {
		return err
	}
	// получем ссылку для загрузки файла
	href, err := getUploadUrl(remotePath)
	if err != nil {
		return err
	}
	defer data.Close()
	// загружаем файл по полученной ссылке методом PUT
	req, err := http.NewRequest("PUT", href, data)
	if err != nil {
		return err
	}
	// в header запроса добавляем токен
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token))

	client := &http.Client{Timeout: timeout, Transport: &http.Transport{DisableKeepAlives: true}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// удаление файла
func deleteFile(path string, token string, timeout time.Duration) error {
	_, err := apiRequest(fmt.Sprintf("resources?path=%s&permanently=true", path), token, "DELETE", timeout)
	return err
}

// получение содержимого директории
func getResource(path, token string, timeout time.Duration) (*Resource, error) {
	res, err := apiRequest(fmt.Sprintf("resources?path=%s&limit=50&sort=-created", path), token, "GET", timeout)
	if err != nil {
		return nil, err
	}

	var result *Resource
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func removeOldBackupsOnServer(token, dirname string, count int, timeout time.Duration) error {
	var reserr []string
	res, err := getResource(dirname, token, timeout)
	if err != nil {
		return err
	}

	for i, v := range res.Embedded.Items {
		if i > count {
			err = deleteFile(v.Path, token, timeout)
			logger.Infof("Удален файл %v %s %s\n", i, v.Name, v.Path)
			if err != nil {
				reserr = append(reserr, err.Error())
			}
		}
	}

	if len(reserr) > 0 {
		return errors.New(strings.Join(reserr, " AND "))
	}

	return nil
}
