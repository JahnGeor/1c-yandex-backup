package yandex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"yd_backup/internal/repo/yandex/models"
)

const yandexDiskURL = "https://cloud-api.yandex.net"

const (
	uploadURL = "v1/disk/resources/upload"
)

type BackupYandex struct {
	client    fasthttp.Client
	Token     string
	expired   time.Duration
	retention int
	Timeout   time.Duration
}

func (y *BackupYandex) GetToken() string {
	return y.Token
}

func (y *BackupYandex) SetToken(token string) {
	y.Token = token
}

func NewBackupYandex(token string, expired time.Duration, retention int) *BackupYandex {
	return &BackupYandex{
		Token:     token,
		client:    fasthttp.Client{},
		expired:   expired,
		retention: retention,
	}
}

func (y *BackupYandex) CreateResource(resourcePath string) error {
	return nil
}

// CreateLink - create yandex link
// ? path=<путь, по которому следует загрузить файл>
// & [overwrite=<признак перезаписи>]
// & [fields=<свойства, которые нужно включить в ответ>]
// Valid status codes: 200 OK
func (y *BackupYandex) CreateLink(params models.Params) (models.Link, error) {
	var link models.Link

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(fmt.Sprintf("%s/%s", yandexDiskURL, uploadURL))

	if params.Path == "" {
		return link, fmt.Errorf("path is empty")
	}

	if params.Fields != nil && len(params.Fields) > 0 {
		request.URI().QueryArgs().Add("fields", strings.Join(params.Fields, ","))
	}

	request.Header.SetMethod(fasthttp.MethodGet)
	request.Header.SetContentType("application/json")
	request.Header.Set("Authorization", "OAuth "+y.Token)

	request.URI().QueryArgs().Add("path", params.Path)
	request.URI().QueryArgs().Add("overwrite", strconv.FormatBool(params.Overwrite))

	err := y.client.DoTimeout(request, response, y.Timeout)

	if err != nil {
		return link, err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		var err *models.ResponseError
		errMarshaller := json.Unmarshal(response.Body(), &err)

		if errMarshaller != nil {
			return link, errMarshaller
		}

		return link, err
	}

	body := response.Body()

	err = json.Unmarshal(body, &link)

	if err != nil {
		return link, err
	}

	return link, nil
}

func (y *BackupYandex) UploadFile(link string, path string) error {
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(link)
	request.Header.SetMethod(fasthttp.MethodPut)

	bufferedReader := bufio.NewReader(file)

	bodyReader, bodyWriter := io.Pipe()

	multipartWriter := multipart.NewWriter(bodyWriter)

	defer multipartWriter.Close()

	go func() {
		file, err := os.Open(path)

		if err != nil {
			return
		}

		defer file.Close()

		part, err := multipartWriter.CreateFormFile("file", filepath.Base(path))

		if err != nil {
			return
		}

		io.Copy(part, bufferedReader)
	}()

	request.Header.SetContentType(multipartWriter.FormDataContentType())
	request.Header.Set("Authorization", "OAuth "+y.Token)

	request.SetBodyStream(bodyReader, -1)

	defer request.CloseBodyStream()

	err = y.client.DoTimeout(request, response, y.Timeout)

	if err != nil {
		return err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		var err *models.ResponseError
		errMarshaller := json.Unmarshal(response.Body(), &err)

		if errMarshaller != nil {
			return errMarshaller
		}

		return err
	}

	return nil
}

func (y *BackupYandex) RemoveFile() error {
	return nil
}
