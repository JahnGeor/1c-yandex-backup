package repo

import (
	"fmt"
	"io"
	"net/http"
)

type ReaderParsed struct {
	length   int64
	progress float64
	total    int64
	io.Reader
}

func (rp *ReaderParsed) Read(p []byte) (int, error) {
	n, err := rp.Reader.Read(p)

	if n > 0 {
		rp.total += int64(n)

		percentage := float64(rp.total) / float64(rp.length) * 100

		fmt.Printf("Процент загрузки: %.2f%%\n", rp.progress)
		fmt.Printf("Объем загруженных данных: %d МБайт\n", rp.total/1024/1024)

		rp.progress = percentage
	}

	return n, err

}

func CreateTransport() *http.Transport {
	result := http.DefaultTransport.(*http.Transport).Clone()

	result.DisableKeepAlives = true

	result.MaxIdleConnsPerHost = result.MaxIdleConns

	return result
}
