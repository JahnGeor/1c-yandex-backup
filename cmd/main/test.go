package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
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

	file, _ := os.Open("D:\\Workplace\\main.fxml")

	reader := &CheckReader{
		file: file,
	}

	wr := &bufio.Writer{}

	n, err := io.Copy(wr, reader)

	if err != nil {
		log.Println(err)
	}

	fmt.Println(n)
}
