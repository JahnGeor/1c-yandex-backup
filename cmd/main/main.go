package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"os"
	"runtime"
	"time"
	"yd_backup/internal/models"
	"yd_backup/internal/repo"
)

var logsPath = fmt.Sprintf("./logs/%s.log", time.Now().Format("2006_01_02"))

var configPath = "./config/config.json"

func main() {

	settingFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDONLY, 0660)

	if err != nil {
		logger.Errorf("Критическая ошибка открытия файла конфигурации приложения: %v", err)
		runtime.Goexit()
	}

	var setting models.Setting

	if err := json.
		NewDecoder(settingFile).Decode(&setting); err != nil {
		logger.Errorf("Критическая ошибка чтения конфигурации: %v", err)
		runtime.Goexit()
	}

	if err := setting.Validate(); err != nil {
		logger.Errorf("Критическая ошибка конфигурации приложения: %v", err)
		runtime.Goexit()

	}

	// Logger
	lf, err := os.OpenFile(logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Errorf("Критическая ошибка запуска приложения: %v", err)
		runtime.Goexit()
	}

	l := logger.Init("Logger", setting.Verbose, true, lf)

	defer func() {
		logger.Infof("Завершение работы приложения\n\n")
		l.Close()
		lf.Close()
	}()

	// Logger
	logger.Infof("Начало работы приложения")

	repo.BackupAll(setting)

	//
	//size, path, err := makeBackup(setting.Databases.Files, setting.Backup.Dir)
	//if err != nil {
	//	logger.Errorf("[ЭТАП 1] Ошибка при выполнении операции бэкапирования: %s", err.Error())
	//	runtime.Goexit()
	//}
	//logger.Infof("[ЭТАП 1] Создан бэкап, размер: %f MB, путь хранения: %s", float32(size)/(1<<20), path)
	//err = eraseBackup(*backupDir, *timeErasing)
	//if err != nil {
	//	logger.Errorf("[ЭТАП 2] Ошибка при выполнении операции очистки старых бэкапов: %s", err.Error())
	//	runtime.Goexit()
	//}
	//logger.Infof("[ЭТАП 2] Произведена очистка бэкапов")
	//
	//yandexDir := "backup/"
	//backupFile := time.Now().Format("2006-01-02_15-04-05") + ".1CD"
	//
	//err = createFolder(*token, yandexDir, *timeout)
	//if err != nil {
	//	logger.Errorf("[ЭТАП 3] Ошибка в процессе создания папки для бэкапа на Yandex Disk %s", err.Error())
	//	runtime.Goexit()
	//}
	//logger.Infof("[ЭТАП 3] Создана директория для бэкапов")
	//
	//err = uploadFile(path, yandexDir+backupFile, *token, *timeout)
	//
	//if err != nil {
	//	logger.Errorf("[ЭТАП 4] Ошибка в процессе загрузки бэкапа на Yandex Disk %s", err.Error())
	//	runtime.Goexit()
	//}
	//
	//logger.Infof("[ЭТАП 4] Произведена загрузка бэкапов")
	//
	//if err = removeOldBackupsOnServer(*token, yandexDir, *count, *timeout); err != nil {
	//	logger.Errorf("[ЭТАП 5] Ошибка в процессе очистки бэкапов на Yandex Disk %s", err.Error())
	//	runtime.Goexit()
	//}
	//logger.Infof("[ЭТАП 5] Произведена очистка бэкапов на Яндекс Диске")

}

// Модуль бэкапирования
//func makeBackup(dbPath []string, dbBackupDir string) (int64, string, error) {
//
//	for _, v := range dbPath {
//		if _, err := os.Stat(v); os.IsNotExist(err) {
//			logger.Warning("Файл %s не существует: %v", v, err)
//			continue
//		}
//
//		fs, err := os.OpenFile(v, os.O_RDONLY, 0660)
//
//		if err != nil {
//			logger.Warning("Ошибка в процессе создания бэкапа, невозможно открыть файл базы для чтения %s: %v", v, err)
//			continue
//		}
//
//		extension := filepath.Ext(v)
//
//		fileName := fmt.Sprintf("%s/%s_%s%s", fs.Name(), dbBackupDir, time.Now().Format("backup_2006_01_02_15_04_05"), extension)
//
//		fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0660)
//
//		if err != nil {
//			logger.Warning("Ошибка в процессе создания бэкапа, невозможно открыть файл бэкапа для записи %s: %v", v, err)
//			continue
//		}
//
//		written, err := io.Copy(fd, fs)
//		if err != nil {
//			logger.Warning("Ошибка в процессе создания бэкапа, невозможно скопировать файл базы данных %s: %v", v, err)
//			continue
//		}
//
//		fd.Close()
//		fs.Close()
//	}
//
//	defer fd.Close()
//
//	return written, fileName, nil
//
//}
//
//func eraseBackup(backupDir string, duration time.Duration) error {
//	e := filepath.Walk(backupDir, func(path string, info fs.FileInfo, err error) error {
//		if err == nil && !info.IsDir() && time.Now().Sub(info.ModTime()) >= duration {
//			err = os.Remove(path)
//			if err != nil {
//				return err
//			}
//		}
//		return nil
//	})
//	return e
//}
//
//// Модуль взаимодействия с Yandex Disk
//
//var ynxUrl = "https://cloud-api.yandex.net/v1/disk"
//
//type Resource struct {
//	Name       string `json:"name"`
//	Path       string `json:"path"`
//	Created    string `json:"created"`
//	ResourceId string `json:"resource_id"`
//	Type       string `json:"type"`
//	MimeType   string `json:"mime_type"`
//	Embedded   struct {
//		Items []Resource `json:"items"`
//		Path  string     `json:"path"`
//	} `json:"_embedded"`
//}
//
//func apiRequest(path, token, method string, timeout time.Duration) (*http.Response, error) {
//	client := http.Client{Timeout: timeout}
//	url := fmt.Sprintf("%s/%s", ynxUrl, path)
//	req, _ := http.NewRequest(method, url, nil)
//	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token))
//	return client.Do(req)
//}
//
//// создание директории
//func createFolder(token, path string, timeout time.Duration) error {
//	_, err := apiRequest(fmt.Sprintf("resources?path=%s", path), token, "PUT", timeout)
//	return err
//}
//
//// загрузка файла
//func uploadFile(localPath, remotePath, token string, timeout time.Duration) error {
//	// функция получения url для загрузки файла
//	getUploadUrl := func(path string) (string, error) {
//		res, err := apiRequest(fmt.Sprintf("resources/upload?path=%s&overwrite=true", path), token, "GET", timeout)
//		if err != nil {
//			return "", err
//		}
//		var resultJson struct {
//			Href string `json:"href"`
//		}
//		err = json.NewDecoder(res.Body).Decode(&resultJson)
//		if err != nil {
//			return "", err
//		}
//		defer res.Body.Close()
//		return resultJson.Href, err
//
//	}
//
//	// читаем локальный файл с диска
//	data, err := os.Open(localPath)
//	if err != nil {
//		return err
//	}
//	// получем ссылку для загрузки файла
//	href, err := getUploadUrl(remotePath)
//
//	if err != nil {
//		return err
//	}
//	defer data.Close()
//	// загружаем файл по полученной ссылке методом PUT
//	req, err := http.NewRequest("PUT", href, data)
//	if err != nil {
//		return err
//	}
//	// в header запроса добавляем токен
//	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token))
//
//	client := &http.Client{Timeout: timeout, Transport: &http.Transport{DisableKeepAlives: true}}
//	res, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer res.Body.Close()
//	return nil
//}
//
//// удаление файла
//func deleteFile(path string, token string, timeout time.Duration) error {
//	_, err := apiRequest(fmt.Sprintf("resources?path=%s&permanently=true", path), token, "DELETE", timeout)
//	return err
//}
//
//// получение содержимого директории
//func getResource(path, token string, timeout time.Duration) (*Resource, error) {
//	res, err := apiRequest(fmt.Sprintf("resources?path=%s&limit=50&sort=-created", path), token, "GET", timeout)
//	if err != nil {
//		return nil, err
//	}
//
//	var result *Resource
//	err = json.NewDecoder(res.Body).Decode(&result)
//	if err != nil {
//		return nil, err
//	}
//	return result, nil
//}
//
//func removeOldBackupsOnServer(token, dirname string, count int, timeout time.Duration) error {
//	var reserr []string
//	res, err := getResource(dirname, token, timeout)
//	if err != nil {
//		return err
//	}
//
//	for i, v := range res.Embedded.Items {
//		if i > count {
//			err = deleteFile(v.Path, token, timeout)
//			logger.Infof("Удален файл %v %s %s\n", i, v.Name, v.Path)
//			if err != nil {
//				reserr = append(reserr, err.Error())
//			}
//		}
//	}
//
//	if len(reserr) > 0 {
//		return errors.New(strings.Join(reserr, " AND "))
//	}
//
//	return nil
//}
