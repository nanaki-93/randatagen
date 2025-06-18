package service

import (
	"database/sql"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"log"
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

func (dbService *DbService) Open(gen model.DataGen) {
	dbService.DbConn = getDbConn(gen)
}

func getDbConn(dataGen model.DataGen) *sql.DB {
	dbConn, err := sql.Open(dataGen.DbType, "postgres://"+dataGen.DbUser+":"+dataGen.DbPassword+"@"+dataGen.DbHost+":"+dataGen.DbPort+"/"+dataGen.DbName+"?sslmode=disable")
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	// Check if the connection is successful
	err = dbConn.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")
	return dbConn
}
