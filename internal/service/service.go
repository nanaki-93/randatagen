package service

import (
	"github.com/nanaki-93/randatagen/internal/model"
	"io"
)

type RanDataService interface {
	io.WriteCloser
	Open(target model.DataGen)
}

func NewRandataService(isToFile bool) RanDataService {

	if isToFile {
		return newFileService()
	}
	return newDbService()
}
