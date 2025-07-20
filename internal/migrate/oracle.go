package migrate

import (
	"database/sql"
	"github.com/nanaki-93/randatagen/internal/model"
)

type OracleDbProvider struct {
	migrationData model.MigrationData
}

func NewOracleDbProvider(migrationData model.MigrationData) DbProvider {
	return &OracleDbProvider{
		migrationData: migrationData,
	}
}

func (p *OracleDbProvider) Open() (source, taget *sql.DB, err error) {
	// Implement the logic to open a connection to the OracleQL database
	// using p.migrationData.SourceConnection and p.migrationData.TargetConnection
	return nil, nil, nil
}
func (p *OracleDbProvider) GetTablesToMigrate(sourceConn *sql.DB) ([]string, error) {
	// Implement the logic to retrieve the list of tables to migrate
	return nil, nil
}
func (p *OracleDbProvider) MigrateTable(source, taget *sql.DB, table string) error {
	// Implement the logic to migrate a specific table from source to target
	return nil
}
func (p *OracleDbProvider) Close(source, target *sql.DB) error {
	// Implement the logic to close the database connection
	return nil
}
