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

type OracleDataProvider struct {
	oracleTypes map[string]string
}

func NewOracleDataProvider() DataProvider {
	return &OracleDataProvider{
		oracleTypes: oracleTypes,
	}
}
func (dbTemplate *OracleDataProvider) GetValueType(datatype string) (string, error) {
	if valueType, exists := dbTemplate.oracleTypes[datatype]; exists {
		return valueType, nil
	}

	return "", fmt.Errorf("[!] Oracle: datatype %s is not supported\n", datatype)
}
func (dbTemplate *OracleDataProvider) GenString(length int) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = OracleCharSet[seededRand.Intn(len(OracleCharSet)-1)]
	}
	return "'" + string(b) + "'"
}

func (dbTemplate *OracleDataProvider) GenBool() string {
	if seededRand.Intn(2) == 0 {
		return "TRUE"
	}
	return "FALSE"

}
func (dbTemplate *OracleDataProvider) GenNumber(length int) string {
	rang := seededRand.Intn(10 * length)
	return strconv.Itoa(rang)

}
func (dbTemplate *OracleDataProvider) GenFloat() string {
	rang := seededRand.Float64()
	return strconv.FormatFloat(rang, 'f', -1, 64)
}

func (dbTemplate *OracleDataProvider) GenUUid() string {
	return WithSingleQuote(uuid.New().String())
}

func (dbTemplate *OracleDataProvider) GenTs(now bool) string {
	if now {
		return "CURRENT_TIMESTAMP"
	}
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	randomNow := time.Unix(randomTime, 0)
	return WithSingleQuote(randomNow.Format(time.RFC3339))
}
