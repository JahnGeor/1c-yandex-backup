package main

import (
	"fmt"
	"os"
	"time"
	"yd_backup/internal/repo/yandex"
	"yd_backup/internal/repo/yandex/models"
)

type CheckReader struct {
	file *os.File
	len  int
}

func (c *CheckReader) Read(p []byte) (n int, err error) {
	n, err = c.file.Read(p)

	return
}

func main() {
	b := yandex.BackupYandex{
		Token:   "y0_AgAAAAAi4OVKAAo8SwAAAADoq7JNqpCSv53NSM6F9-pToFzBLznjR00",
		Timeout: time.Second * 10,
	}

	params := models.Params{
		Path:      "backup.txt",
		Overwrite: true,
		Fields:    nil,
	}

	response, err := b.CreateLink(params)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
