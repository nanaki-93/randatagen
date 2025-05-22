package internal

import (
	"github.com/nanaki-93/randatagen/internal/model"
	"math/rand"
	"strings"
	"time"
)

func GetSqlTemplate(columns []model.Column) string {

	columnsName := make([]string, len(columns))
	valuesSLice := make([]string, len(columns))
	for i, column := range columns {
		columnsName[i] = column.Name
		valuesSLice[i] = GetValue(column.Datatype, column.Length)
	}
	columnsJoin := strings.Join(columnsName, ", ")
	valuesJoin := strings.Join(valuesSLice, ", ")
	return "Insert into table_name (" + columnsJoin + ") values (" + valuesJoin + ");"
}

func GetValue(datatype string, length int) string {
	switch datatype {
	case "int":
		return "1"
	case "float":
		return "1.0"
	case "bool":
		return "true"
	default:
		return StringWithCharset(length)
	}
}

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int) string {
	b := make([]byte, length)
	charSet := "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"
	for i := range b {
		b[i] = charSet[seededRand.Intn(len(charSet)-1)]
	}
	return string(b)
}
