package service

import (
	"errors"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"log"
	"os"
)

type FileService struct {
	FileToWrite *os.File
}

func newFileService() *FileService {
	return &FileService{}
}
func (service *FileService) Open(gen model.GenerateData) {
	service.FileToWrite = openOutputFile(gen.OutputFilePath)
}
func (service *FileService) Write(insertSql []byte) (n int, err error) {
	if _, err = service.FileToWrite.WriteString(string(insertSql)); err != nil {
		return 0, fmt.Errorf("[!] %s\n", err)
	}
	return len(insertSql), nil
}
func (service *FileService) Close() error {
	if err := service.FileToWrite.Close(); err != nil {
		return fmt.Errorf("[!] %s\n", err)
	}
	return nil
}

func openOutputFile(outputFilePath string) *os.File {
	err := os.Remove(outputFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("[!] %s\n", err)
	}

	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("[!] %s\n", err)
	}
	return f
}
