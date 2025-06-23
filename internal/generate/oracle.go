package generate

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"time"
)

const OracleCharSet = "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"

var oracleTypes = map[string]string{
	"NUMBER":    GetNumber,
	"INTEGER":   GetNumber,
	"FLOAT":     GetFloat,
	"CHAR":      GetString,
	"NCHAR":     GetString,
	"VARCHAR2":  GetString,
	"NVARCHAR2": GetString,
	"CLOB":      GetString,
	"NCLOB":     GetString,
	"LONG":      GetNumber,
	"BLOB":      GetString,
	"DATE":      GetDateOrTs,
	"TIMESTAMP": GetDateOrTs,
	"RAW":       GetString,
}

type OracleTemplate struct {
	oracleTypes map[string]string
}

func NewOracleTemplate() *OracleTemplate {
	return &OracleTemplate{
		oracleTypes: oracleTypes,
	}
}
func (dbTemplate *OracleTemplate) GetValueType(datatype string) (string, error) {
	if valueType, exists := dbTemplate.oracleTypes[datatype]; exists {
		return valueType, nil
	}

	return "", fmt.Errorf("[!] Oracle: datatype %s is not supported\n", datatype)
}
func (dbTemplate *OracleTemplate) GenString(length int) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = OracleCharSet[seededRand.Intn(len(OracleCharSet)-1)]
	}
	return "'" + string(b) + "'"
}

func (dbTemplate *OracleTemplate) GenBool() string {
	if seededRand.Intn(2) == 0 {
		return "TRUE"
	}
	return "FALSE"

}
func (dbTemplate *OracleTemplate) GenNumber(length int) string {
	rang := seededRand.Intn(10 * length)
	return strconv.Itoa(rang)

}
func (dbTemplate *OracleTemplate) GenFloat() string {
	rang := seededRand.Float64()
	return strconv.FormatFloat(rang, 'f', -1, 64)
}

func (dbTemplate *OracleTemplate) GenUUid() string {
	return uuid.New().String()
}

func (dbTemplate *OracleTemplate) GenTs(now bool) string {
	if now {
		return time.Now().String()
	}
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	randomNow := time.Unix(randomTime, 0)
	return randomNow.String()
}
