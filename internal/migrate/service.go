package migrate

import "github.com/nanaki-93/randatagen/internal/model"

type DbProvider interface {
	Open(migrateConfig model.MigrateData) error
	MigrateSchema(migrateConfig model.MigrateData) error
	Close() error
}
type ProviderFactory func() DbProvider

var ProviderFactories = map[string]ProviderFactory{
	"postgres": NewPostgresProvider,
	"oracle":   NewOracleProvider,
}

type MigratorService struct {
	migrateConfig model.MigrateData
}

func NewMigratorService(migrateConfig model.MigrateData) *MigratorService {
	return &MigratorService{
		migrateConfig: migrateConfig,
	}
}
