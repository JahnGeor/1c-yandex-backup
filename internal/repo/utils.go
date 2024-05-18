package repo

import (
	"fmt"
	"io"
	"os"
)

type Piper struct {
	length   int64
	progress float64
	rtotal   int64
	wtotal   int64
	rw       io.ReadWriter
	name     string
}

func NewWithFile(file *os.File) (*Piper, error) {

	stat, err := file.Stat()

	if err != nil {
		return nil, err
	}

	length := stat.Size()

	return &Piper{
		length:   length,
		progress: 0,
		rtotal:   0,
		wtotal:   0,
		rw:       file,
		name:     stat.Name(),
	}, nil

}

func (rp *Piper) Read(p []byte) (int, error) {
	var n int
	var err error

	n, err = rp.rw.Read(p)

	if n > 0 {
		rp.rtotal += int64(n)

		percentage := float64(rp.rtotal) / float64(rp.length) * 100

		if percentage-rp.progress > 2 {
			fmt.Printf("Чтение файла %s: процент чтения: %.2f%%, объем прочитанных данных: %d МБайт\n", rp.name, percentage, rp.rtotal/1024/1024)
			rp.progress = percentage
		}
	} else if n == 0 {

		fmt.Printf("Чтение файла %s завершено\n", rp.name)
	}

	return n, err

}

func (rp *Piper) Write(p []byte) (int, error) {
	var n int
	var err error

	n, err = rp.rw.Write(p)

	if n > 0 {
		rp.wtotal += int64(n)
		fmt.Printf("Чтение файла %s: объем прочитанных данных: %d МБайт\n", rp.name, rp.rtotal/1024/1024)
	}

	return n, err
}
