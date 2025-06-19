package service

import (
	"github.com/nanaki-93/randatagen/internal/model"
	"io"
)

type RanDataWriter interface {
	io.WriteCloser
	Open(target model.GenerateData)
}

func NewRandataService(isToFile bool) RanDataWriter {

	if isToFile {
		return newFileService()
	}
	return newDbService()
}
