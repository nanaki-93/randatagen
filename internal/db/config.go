package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
	"github.com/nanaki-93/randatagen/internal/model"
	_ "github.com/sijms/go-ora/v2"
	//todo import other drivers as needed
)

// GetConn establishes a connection to the database using the provided DbStruct model.
// It returns a pointer to the sql.DB instance if successful, or logs a fatal error if the connection fails.
// The connection string is constructed using the database type, host, port, name, user, password, and schema.
// It also performs a ping to ensure the connection is valid.
// If the connection is successful, it prints a success message to the console.
// If any error occurs during the connection or ping, it logs the error and exits the program.
// It is important to ensure that the database driver for the specified DbType is imported in the main package.
// todo check if for all the dbtype the connection is always opened in the same way
func GetConn(dbType string, dbModel model.DbStruct) (*sql.DB, error) {
	dbConn, err := sql.Open(dbType, fmt.Sprintf("host=%s port=%d dbname=%s user=%s password='%s' sslmode=disable search_path=%s", dbModel.DbHost, dbModel.DbPort, dbModel.DbName, dbModel.DbUser, dbModel.DbPassword, dbModel.DbSchema))
	//connectionString := "oracle://" + dbParams["username"] + ":" + dbParams["password"] + "@" + dbParams["server"] + ":" + dbParams["port"] + "/" + dbParams["service"]
	//if val, ok := dbParams["walletLocation"]; ok && val != "" {
	//	connectionString += "?TRACE FILE=trace.log&SSL=enable&SSL Verify=false&WALLET=" + url.QueryEscape(dbParams["walletLocation"])
	//}
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}
	// Check if the connection is successful
	err = dbConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}
	fmt.Printf("Successfully connected to %s!\n", dbType)
	return dbConn, nil
}
