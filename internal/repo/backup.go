package repo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/logger"
	"github.com/valyala/fasthttp"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"yd_backup/internal/models"
)

const yandexUrl = "https://cloud-api.yandex.net/v1/disk"

const (
	uploadPath = "resources/upload?path=%s&overwrite=true"
)

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

func BackupAll(settings models.Settings) {
	wg := sync.WaitGroup{}
	builder, err := NewBackupBuilder(settings)

	if err != nil {
		logger.Errorf("Ошибка создания бэкапа: %v", err)
		return
	}

	for key, v := range settings.Databases.Paths {
		dbPath := v
		goroutineKey := key

		wg.Add(1)

		go func() {
			if err := builder.Backup(dbPath); err != nil {
				logger.Errorf("Ошибка выполнения горутины %d: %v", goroutineKey, err)
			} else {
				logger.Infof("Успешное выполнение горутины %d", goroutineKey)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

type BackupBuilder struct {
	yandex models.Yandex
	backup models.Backup
	client *http.Client
}

func NewBackupBuilder(settings models.Settings) (*BackupBuilder, error) {
	b := &BackupBuilder{
		yandex: settings.Yandex,
		backup: settings.Backup,
		//client: &http.Client{},
		client: &http.Client{Timeout: settings.Yandex.Timeout.Duration, Transport: CreateTransport()},
	}

	err := b.CreateFolderIfNotExist()

	if err != nil {
		return nil, err
	}

	return b, nil

}

func (b *BackupBuilder) CreateFolderIfNotExist() error {
	if _, err := b.getResource(b.yandex.Dir); err != nil {
		err = b.CreateFolder()

		if err != nil {
			return fmt.Errorf("ошибка в процессе создания папки %s: %w", b.yandex.Dir, err)
		}

		logger.Infof("Папка %s создана", b.yandex.Dir)
	} else {
		logger.Infof("Папка %s уже существует", b.yandex.Dir)
	}

	return nil
}

func (b *BackupBuilder) Backup(dbPath string) error {
	// TODO: Выполнить бэкапирование на диск в файл
	backupPath, err := b.CreateBackupLocal(dbPath)

	if err != nil {
		return fmt.Errorf("ошибка в процессе создания локальной копии файла %s: %w", dbPath, err)
	}

	// TODO: Выполнить создание файла на диске

	href, err := b.CreateBackupRemote(filepath.Base(backupPath))

	if err != nil {
		return fmt.Errorf("ошибка в процессе создания файла на Yandex Disk %s: %w", dbPath, err)

	}

	// TODO: Перенести данные на yandex disk

	//err = b.PushBackupRemote(href, backupPath)

	err = b.PushBackupRemoteFastHttp(href, backupPath)

	if err != nil {
		return fmt.Errorf("ошибка в процессе загрузки файла %s на Yandex Disk: %w", backupPath, err)
	}

	// TODO: Произвести удаление старых бэкапов на диске
	if err := b.EraseLocal(); err != nil {
		return fmt.Errorf("ошибка в процессе удаления локальной копии файла %s: %w", dbPath, err)
	}

	return nil
}

func (b *BackupBuilder) CreateFolder() error {

	path := fmt.Sprintf("resources?path=%s", b.yandex.Dir)

	response, err := b.createRequest(path, http.MethodPut)

	defer response.Body.Close()

	if err != nil {
		return err
	}

	var responseModel models.ResponseError

	err = json.NewDecoder(response.Body).Decode(&responseModel)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return responseModel
	}

	return nil
}

func (b *BackupBuilder) CreateBackupLocal(dbPath string) (string, error) {
	dbFile, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		return "", err
	}

	defer dbFile.Close()

	backupName := fmt.Sprintf("%s_%s", time.Now().Format("2006_01_02_15_04_05"), filepath.Base(dbFile.Name()))

	backupDir, err := filepath.Abs(b.backup.Dir)

	if err != nil {
		return "", err
	}

	backupPath := fmt.Sprintf("%s\\%s", backupDir, backupName)

	backupFile, err := os.OpenFile(backupPath, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		return "", err
	}

	defer backupFile.Close()

	written, err := io.Copy(backupFile, dbFile)

	if err != nil {
		return "", err
	}

	logger.Infof("Создан бэкап %s, объем %d байт", backupPath, written)

	return backupFile.Name(), nil
}

func (b *BackupBuilder) CreateBackupRemote(remotePath string) (string, error) {
	result, err := b.createRequest(fmt.Sprintf(uploadPath, fmt.Sprintf("%s/%s", b.yandex.Dir, remotePath)), http.MethodGet)
	defer result.Body.Close()
	if err != nil {
		return "", err
	}

	var resultJson struct {
		Href string `json:"href"`
	}
	err = json.NewDecoder(result.Body).Decode(&resultJson)

	if err != nil {
		return "", err
	}

	return resultJson.Href, err
}

func (b *BackupBuilder) PushBackupRemote(href string, backupPath string) error {

	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	rp := &ReaderParsed{Reader: r}

	go func() {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("file", filepath.Base(backupPath))

		if err != nil {
			return
		}

		file, err := os.OpenFile(backupPath, os.O_RDONLY, 0660)

		if err != nil {
			return
		}

		defer file.Close()

		if _, err := io.Copy(part, file); err != nil {
			return
		}
	}()

	client := &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
		MaxConnsPerHost:           2000,
		MaxIdleConnDuration:       100 * time.Second,
		MaxConnDuration:           150 * time.Second,
		ReadTimeout:               b.yandex.Timeout.Duration,
		WriteTimeout:              b.yandex.Timeout.Duration,
		MaxConnWaitTimeout:        50 * time.Second,
		MaxIdemponentCallAttempts: 5,
	}

	req := fasthttp.AcquireRequest()
	req.SetBodyStream(rp, 0)

	req.SetRequestURI(href)
	req.Header.SetMethod(fasthttp.MethodPut)
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", b.yandex.Token))
	//req.Header.Add("Content-Type", m.FormDataContentType())

	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		return err
	}

	if resp.StatusCode() != fasthttp.StatusCreated {
		return errors.New(fmt.Sprintf("Не удалось загрузить файл по ссылке %s: %d", href, resp.StatusCode()))
	}

	return nil

	// в header запроса добавляем токен

	// result, err := b.client.Do(req)

	//if err != nil {
	//	return err
	//}

	//defer result.Body.Close()

	//if resp.StatusCode() != fasthttp.StatusCreated {
	//	return errors.New(fmt.Sprintf("Не удалось загрузить файл по ссылке %s: %d", href, resp.StatusCode()))
	//}

	return nil
}

func (b *BackupBuilder) PushBackupRemoteFastHttp(href string, backupPath string) error {
	file, err := os.Open(backupPath)

	if err != nil {
		return err
	}

	defer file.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(backupPath))

	if err != nil {
		return err
	}

	n, err := io.Copy(part, file)

	if err != nil {
		return err
	}

	err = writer.Close()

	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", href, body)

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", b.yandex.Token))

	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Set("Content-Length", strconv.Itoa(int(n)))

	client := &http.Client{Timeout: b.yandex.Timeout.Duration}

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("Не удалось загрузить файл по ссылке %s: %d", href, res.StatusCode))
	}

	_, err = io.Copy(io.Discard, res.Body)

	if err != nil {
		return err
	}

	return nil
}

func (b *BackupBuilder) EraseLocal() error {
	e := filepath.Walk(b.backup.Dir, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() && time.Now().Sub(info.ModTime()) >= b.backup.Expired.Duration {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return e
}

func (b *BackupBuilder) EraseRemote() error {
	return nil
}

func (b *BackupBuilder) createRequest(path string, method string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", yandexUrl, path)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", b.yandex.Token))

	return b.client.Do(req)
}

func (b *BackupBuilder) getResource(resourcePath string) (models.Resource, error) {
	var resource models.Resource
	url := fmt.Sprintf("resources?path=%s", resourcePath)
	result, err := b.createRequest(url, http.MethodGet)

	if err != nil {
		return resource, err
	}

	if result.StatusCode != http.StatusOK {
		var responseModel models.ResponseError
		err = json.NewDecoder(result.Body).Decode(&responseModel)

		if err != nil {
			return resource, err
		}

		return resource, responseModel
	}

	err = json.NewDecoder(result.Body).Decode(&resource)

	if err != nil {
		return resource, err
	}

	return resource, nil
}
