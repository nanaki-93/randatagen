package migrate

import (
	"database/sql"
	"github.com/nanaki-93/randatagen/internal/model"
)

type DbProvider interface {
	Open(migrationConfig model.MigrationData) (*sql.DB, *sql.DB, error)
	MigrateSchema(migrateConfig model.MigrationData) error
	GetTablesToMigrate(migrateData model.MigrationData, sourceConn *sql.DB) ([]string, error)
	MigrateTable(sourceConn, targetConn *sql.DB, table string) error
	Close() error
}
type ProviderFactory func() DbProvider

var ProviderFactories = map[string]ProviderFactory{
	"postgres": NewPostgresProvider,
	"oracle":   NewOracleProvider,
}

type MigrationService struct {
	migrateConfig model.MigrationData
}

func NewMigrationService(migrateConfig model.MigrationData) *MigrationService {
	return &MigrationService{
		migrateConfig: migrateConfig,
	}
}
