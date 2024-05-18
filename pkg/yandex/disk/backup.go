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
	uploadURL   = "v1/disk/resources/upload"
	resourceURL = "v1/disk/resources"
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

func (y *YandexDisk) GetResource(params models.Params) (models.Resource, error) {
	var result models.Resource

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	if params.Path == "" {
		return result, fmt.Errorf("path is empty")
	}

	request.SetRequestURI(fmt.Sprintf("%s/%s", yandexDiskURL, resourceURL))
	request.Header.SetMethod(fasthttp.MethodGet)
	request.Header.SetContentType("application/json")
	request.Header.Set("Authorization", fmt.Sprintf("OAuth %s", y.Token))

	if params.Fields != nil && len(params.Fields) > 0 {
		request.URI().QueryArgs().Add("fields", strings.Join(params.Fields, ","))
	}

	if params.Limit > 0 {
		request.URI().QueryArgs().Add("limit", strconv.Itoa(params.Limit))
	}

	if params.Offset > 0 {
		request.URI().QueryArgs().Add("offset", strconv.Itoa(params.Offset))
	}

	if params.Sort != "" {
		request.URI().QueryArgs().Add("sort", params.Sort)
	}

	if params.PreviewCrop {
		request.URI().QueryArgs().Add("preview_crop", strconv.FormatBool(params.PreviewCrop))
	}

	if params.PreviewSize != "" {
		request.URI().QueryArgs().Add("preview_size", params.PreviewSize)
	}

	request.URI().QueryArgs().Add("path", params.Path)

	if err := y.client.Do(request, response); err != nil {
		return result, err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		responseError := &models.ResponseError{}

		if err := json.Unmarshal(response.Body(), responseError); err != nil {
			return result, err
		}

		responseError.StatusCode = response.StatusCode()

		return result, responseError
	}

	if err := json.Unmarshal(response.Body(), &result); err != nil {
		return result, err
	}

	return result, nil

}

func (y *YandexDisk) CreateResource(params models.Params) (models.Link, error) {
	var link models.Link

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(response)
	defer fasthttp.ReleaseRequest(request)

	request.Header.SetMethod(fasthttp.MethodPut)
	request.Header.SetContentType("application/json")

	if params.Path == "" {
		return link, fmt.Errorf("path is empty")
	}

	if params.Fields != nil {
		request.URI().QueryArgs().Add("fields", strings.Join(params.Fields, ","))
	}
	request.URI().SetPath(resourceURL)
	request.URI().SetHost(yandexDiskURL)

	request.URI().QueryArgs().Add("path", params.Path)

	request.Header.Set("Authorization", fmt.Sprintf("OAuth %s", y.Token))
	request.SetRequestURI(fmt.Sprintf("%s/%s", yandexDiskURL, resourceURL))

	if err := y.client.Do(request, response); err != nil {
		return link, err
	}

	if response.StatusCode() != fasthttp.StatusCreated {
		responseError := &models.ResponseError{}

		if err := json.Unmarshal(response.Body(), responseError); err != nil {
			return link, err
		}
		responseError.StatusCode = response.StatusCode()

		return link, responseError
	}

	if err := json.Unmarshal(response.Body(), &link); err != nil {
		return link, err
	}

	return link, nil

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

	if response.StatusCode() != fasthttp.StatusCreated {
		var err *models.ResponseError
		errMarshaller := json.Unmarshal(response.Body(), &err)

		if errMarshaller != nil {
			return errMarshaller
		}

		return err
	}

	return nil
}

func (y *YandexDisk) RemoveResource(params models.Params) (models.Link, error) {
	var link models.Link

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(fmt.Sprintf("%s/%s", yandexDiskURL, resourceURL))

	request.Header.SetMethod(fasthttp.MethodDelete)
	request.Header.SetContentType("application/json")

	if params.Path == "" {
		return link, fmt.Errorf("path is empty")
	}

	request.URI().QueryArgs().Add("path", params.Path)
	request.URI().QueryArgs().Add("permanently", strconv.FormatBool(params.Permanently))

	request.Header.Set("Authorization", fmt.Sprintf("OAuth %s", y.Token))

	err := y.client.Do(request, response)

	if err != nil {
		return link, err
	}

	if response.StatusCode() == fasthttp.StatusNoContent {
		return link, nil
	}

	if response.StatusCode() == fasthttp.StatusAccepted {
		err := json.Unmarshal(response.Body(), &link)

		if err != nil {
			return link, err
		}

		return link, nil
	}

	var errResp *models.ResponseError
	errMarshaller := json.Unmarshal(response.Body(), &errResp)

	if errMarshaller != nil {
		return link, errMarshaller
	}

	return link, errResp
}
