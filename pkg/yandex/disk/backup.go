package disk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
	"strconv"
	"strings"
	"time"
	"yd_backup/internal/repo"
	"yd_backup/pkg/yandex/disk/models"
)

const yandexDiskURL = "https://cloud-api.yandex.net"

const (
	uploadURL = "v1/disk/resources/upload"
)

type YandexDisk struct {
	client  *fasthttp.Client
	Token   string
	Timeout time.Duration
}

func (y *YandexDisk) GetToken() string {
	return y.Token
}

func (y *YandexDisk) SetToken(token string) {
	y.Token = token
}

func NewBackupYandex(token string, timeout time.Duration) *YandexDisk {
	client := &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
		MaxConnsPerHost:           2000,
		MaxIdleConnDuration:       100 * time.Second,
		MaxConnDuration:           150 * time.Second,
		ReadTimeout:               timeout,
		WriteTimeout:              timeout,
		MaxConnWaitTimeout:        50 * time.Second,
		MaxIdemponentCallAttempts: 5,
	}

	return &YandexDisk{
		Token:  token,
		client: client,
	}
}

func (y *YandexDisk) CreateResource(resourcePath string) error {
	return nil
}

// CreateLink - create yandex link
// ? path=<путь, по которому следует загрузить файл>
// & [overwrite=<признак перезаписи>]
// & [fields=<свойства, которые нужно включить в ответ>]
// Valid status codes: 200 OK
func (y *YandexDisk) CreateLink(params models.Params) (models.Link, error) {
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

	err := y.client.Do(request, response)

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

func (y *YandexDisk) UploadFile(link models.Link, path string) error {

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	request.Header.Set("Authorization", fmt.Sprintf("OAuth %s", y.Token))

	request.SetRequestURI(link.Href)
	request.Header.SetMethod(link.Method)
	request.Header.SetMethod(link.Method)
	request.SetRequestURI(link.Href)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	piper, err := repo.NewWithFile(file)

	if err != nil {
		return err
	}

	request.SetBodyStream(piper, -1)

	client := fasthttp.Client{
		ReadTimeout:  y.Timeout,
		WriteTimeout: y.Timeout,
	}

	err = client.Do(request, response)

	if err != nil {
		return err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		fmt.Println(response.StatusCode())
	}

	return nil
}

func (y *YandexDisk) RemoveFile() error {
	return nil
}
