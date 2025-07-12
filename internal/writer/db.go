package writer

import (
	"database/sql"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/db"
	"github.com/nanaki-93/randatagen/internal/model"
)

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

func (dbService *DbService) Open(gen model.RanData) error {
	var err error
	dbService.DbConn, err = db.GetConn(gen.Target)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}
	return nil
}
