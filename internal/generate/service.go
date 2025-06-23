package generate

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/db"
	"github.com/nanaki-93/randatagen/internal/model"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type RanDataWriter interface {
	io.WriteCloser
	Open(target model.RanData)
}

func NewRandataService(isToFile bool) RanDataWriter {

	if isToFile {
		return newFileService()
	}
	return newDbService()
}

const GetNumber = "getNumber"
const GetFloat = "getFloat"
const GetString = "getString"
const GetBool = "getBool"
const GetDateOrTs = "getDateOrTimestamp"
const GetUuid = "getUUID"
const GetJson = "getJson"
const BatchSize = 1000

type FileService struct {
	FileToWrite *os.File
}

func newFileService() *FileService {
	return &FileService{}
}
func (service *FileService) Open(gen model.RanData) {
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

type DbService struct {
	DbConn *sql.DB
}

func newDbService() *DbService {
	return &DbService{}
}

func (dbService *DbService) Write(insertSql []byte) (n int, err error) {
	_, err = dbService.DbConn.Exec(string(insertSql))
	if err != nil {
		return 0, fmt.Errorf("Error executing query: %w ", err)
	}
	return len(insertSql), err
}

func (dbService *DbService) Close() error {
	if err := dbService.DbConn.Close(); err != nil {
		return fmt.Errorf("[!] %s\n", err)
	}
	return nil
}

func (dbService *DbService) Open(gen model.RanData) {
	dbService.DbConn = db.GetConn(gen.Target)
}
