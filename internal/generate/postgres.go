package generate

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"time"
)

const PgCharSet = "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"

var postgresTypes = map[string]string{
	"BIGINT":            GetNumber,
	"BIGSERIAL":         GetNumber,
	"BIT":               GetNumber,
	"BIT_VARYING":       GetNumber,
	"BOOLEAN":           GetBool,
	"BYTEA":             GetString,
	"CHARACTER":         GetString,
	"CHARACTER_VARYING": GetString,
	"DATE":              GetDateOrTs,
	"DOUBLE_PRECISION":  GetFloat,
	"INTEGER":           GetNumber,
	"INTERVAL":          GetDateOrTs,
	"JSON":              GetJson,
	"JSONB":             GetJson,
	"MONEY":             GetFloat,
	"NUMERIC":           GetNumber,
	"PG_LSN":            GetString,
	"PG_SNAPSHOT":       GetString,
	"REAL":              GetNumber,
	"SMALLINT":          GetNumber,
	"SMALLSERIAL":       GetNumber,
	"SERIAL":            GetNumber,
	"TEXT":              GetString,
	"TIME":              GetDateOrTs,
	"TIMESTAMP":         GetDateOrTs,
	"TSQUERY":           GetString,
	"TSVECTOR":          GetString,
	"UUID":              GetUuid,
	"XML":               GetString,
}

type PostgresDataProvider struct {
	postgresTypes map[string]string
}

func NewPostgresDataProvider() DataProvider {
	return &PostgresDataProvider{
		postgresTypes: postgresTypes,
	}
}
func (tp *PostgresDataProvider) GetValueType(datatype string) (string, error) {
	if valueType, exists := tp.postgresTypes[datatype]; exists {
		return valueType, nil
	}

	return "", fmt.Errorf("[!] Postgres: datatype %s is not supported\n", datatype)
}
func (tp *PostgresDataProvider) GenString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = PgCharSet[seededRand.Intn(len(PgCharSet)-1)]
	}
	return WithSingleQuote(string(b))
}

func (tp *PostgresDataProvider) GenBool() string {
	if seededRand.Intn(2) == 0 {
		return "true"
	}
	return "false"

}
func (tp *PostgresDataProvider) GenNumber(length int) string {
	rang := seededRand.Intn(10 * length)
	return strconv.Itoa(rang)

}
func (tp *PostgresDataProvider) GenFloat() string {
	rang := seededRand.Float64()
	return strconv.FormatFloat(rang, 'f', -1, 64)
}

func (tp *PostgresDataProvider) GenUUid() string {
	return WithSingleQuote(uuid.New().String())
}

func (tp *PostgresDataProvider) GenTs(now bool) string {
	if now {
		return "now()"
	}
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	randomNow := time.Unix(randomTime, 0)

	return WithSingleQuote(randomNow.Format(time.RFC3339))
}
