package db

import (
	"database/sql"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"log"
)

func GetPostgresConn(dbModel model.DbStruct) *sql.DB {
	dbConn, err := sql.Open(dbModel.DbType, fmt.Sprintf("host=%s port=%d dbname=%s user=%s password='%s' sslmode=disable search_path=%s", dbModel.DbHost, dbModel.DbPort, dbModel.DbName, dbModel.DbUser, dbModel.DbPassword, dbModel.DbSchema))
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
