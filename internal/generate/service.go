package generate

import (
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"math/rand"
	"strings"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

const GetString = "getString"
const GetNumber = "getNumber"
const GetFloat = "getFloat"
const GetBool = "getBool"
const GetDateOrTs = "getDateOrTimestamp"
const GetUuid = "getUUID"
const GetJson = "getJson"
const BatchSize = 1000

type DataProvider interface {
	GenString(length int) string
	GenBool() string
	GenNumber(length int) string
	GenFloat() string
	GenUUid() string
	GenTs(bool) string
	GetValueType(datatype string) (string, error)
}

type ProviderFactory func() DataProvider

var ProviderFactories = map[string]ProviderFactory{
	"postgres": NewPostgresDataProvider,
	"oracle":   NewOracleDataProvider,
}

type GeneratorService struct {
	provider DataProvider
}

func NewGeneratorService(provider DataProvider) *GeneratorService {
	return &GeneratorService{
		provider: provider,
	}
}

func (ts *GeneratorService) GenerateSql(dataGen model.RanData) []string {

	var insertSqlSlice []string

	columns := dataGen.Columns
	prefix := getPrefixInsert(dataGen)
	var builder strings.Builder
	rowsInBatch := 0

	for i := 0; i < dataGen.Rows; i++ {
		if rowsInBatch == 0 {
			builder.Reset()
			builder.WriteString(prefix)
		}
		if rowsInBatch > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("\n(")
		builder.WriteString(ts.getValues(columns))
		builder.WriteString(")")
		rowsInBatch++
		if rowsInBatch == BatchSize || i == dataGen.Rows-1 {
			builder.WriteString(";")
			insertSqlSlice = append(insertSqlSlice, builder.String())
			rowsInBatch = 0
		}
	}

	fmt.Println("generated insert Slice")
	return insertSqlSlice
}

func (ts *GeneratorService) getValues(columns []model.Column) string {
	valuesSLice := make([]string, len(columns))
	for i, column := range columns {
		valuesSLice[i] = ts.GetValue(column.Datatype, column.Length, column.Now)
	}
	valuesJoin := strings.Join(valuesSLice, ", ")
	return valuesJoin
}

func getPrefixInsert(dataGen model.RanData) string {
	columns := dataGen.Columns
	columnsName := make([]string, len(columns))
	for i, column := range columns {
		columnsName[i] = column.Name
	}
	columnsJoin := strings.Join(columnsName, ", ")
	return fmt.Sprintf("Insert into %s.%s(%s) values", dataGen.Target.DbSchema, WithDoubleQuote(dataGen.Target.DbTable), columnsJoin)
}

func (ts *GeneratorService) GetValue(datatype string, length int, now bool) string {

	valueType, err := ts.provider.GetValueType(datatype)
	if err != nil {
		fmt.Println("[!] Error getting value type:", err)
	}
	switch valueType {
	case GetNumber:
		return ts.provider.GenNumber(length)
	case GetFloat:
		return ts.provider.GenFloat()
	case GetBool:
		return ts.provider.GenBool()
	case GetString:
		return ts.provider.GenString(length)
	case GetUuid:
		return ts.provider.GenUUid()
	case GetDateOrTs:
		return ts.provider.GenTs(now)
	default:
		return "NotSupported"

	}
}

func WithSingleQuote(input string) string {
	return "'" + input + "'"
}

func WithDoubleQuote(input string) string {
	return "\"" + input + "\""
}
