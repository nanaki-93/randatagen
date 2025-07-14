package migrate

import (
	"database/sql"
	"github.com/nanaki-93/randatagen/internal/model"
)

type DbProvider interface {
	Open() (source, target *sql.DB, err error)
	GetTablesToMigrate(sourceConn *sql.DB) ([]string, error)
	MigrateTable(source, target *sql.DB, table string) error
	Close(source, target *sql.DB) error
}
type ProviderFactory func(data model.MigrationData) DbProvider

var ProviderFactories = map[string]ProviderFactory{
	"postgres": NewPostgresDbProvider,
	"oracle":   NewOracleDbProvider,
}

type MigrationService struct {
	MigrationProvider DbProvider
}

func NewMigrationService(migrationProvider DbProvider) *MigrationService {
	return &MigrationService{
		MigrationProvider: migrationProvider,
	}
}
