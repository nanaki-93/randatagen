package generate

import (
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"strings"
)

type TemplateService struct {
	template DataGenerator
}

func NewService(template DataGenerator) *TemplateService {
	return &TemplateService{
		template: template,
	}
}

type DataGenerator interface {
	GenString(length int) string
	GenBool() string
	GenNumber(length int) string
	GenFloat() string
	GenUUid() string
	GenTs(bool) string
	GetValueType(datatype string) (string, error)
}

func (ts *TemplateService) GetSqlTemplate(dataGen model.RanData) []string {

	insertSqlSlice := make([]string, dataGen.Rows/BatchSize+1)

	columns := dataGen.Columns
	columnsName := make([]string, len(columns))

	prefix := getPrefixInsert(dataGen)
	sql := prefix
	for index := range dataGen.Rows {

		valuesJoin := ts.getValues(columns, columnsName)
		sql += "\n(" + valuesJoin + ")"

		if index != 0 && index%BatchSize == 0 || index == dataGen.Rows-1 {
			sql = sql + ";"
			insertSqlSlice = append(insertSqlSlice, sql)
			sql = prefix
		} else {
			sql = sql + ","
		}
	}

	fmt.Println("generated insert Slice")
	return insertSqlSlice
}

func (ts *TemplateService) getValues(columns []model.Column, columnsName []string) string {
	valuesSLice := make([]string, len(columns))
	for i, column := range columns {
		columnsName[i] = column.Name
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

	prefix := "Insert into " + dataGen.Target.DbSchema + "." + WithDoubleQuote(dataGen.Target.DbTable) + "(" + columnsJoin + ") values "
	return prefix
}

func (ts *TemplateService) GetValue(datatype string, length int, now bool) string {

	valueType, err := ts.template.GetValueType(datatype)
	if err != nil {
		fmt.Println("[!] Error getting value type:", err)
	}
	switch valueType {
	case GetNumber:
		return ts.template.GenNumber(length)
	case GetFloat:
		return ts.template.GenFloat()
	case GetBool:
		return ts.template.GenBool()
	case GetString:
		return ts.template.GenString(length)
	case GetUuid:
		return ts.template.GenUUid()
	case GetDateOrTs:
		return ts.template.GenTs(now)
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
