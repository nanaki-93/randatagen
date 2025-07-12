package generate

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOracleDataProvider_GetValueType(t *testing.T) {
	provider := NewOracleDataProvider()

	valueType, err := provider.GetValueType("INTEGER")
	if err != nil || valueType != GetNumber {
		t.Errorf("expected value type %s, got %s", GetNumber, valueType)
	}

	_, err = provider.GetValueType("UNKNOWN_TYPE")
	if err == nil {
		t.Errorf("expected error for unknown type")
	}

}

func TestOracleDataProvider_GenString(t *testing.T) {
	provider := NewOracleDataProvider()

	result := provider.GenString(10)
	if len(result) != 12 {
		t.Errorf("expected string length 12 (including quotes), got %d", len(result))
	}
	if !strings.HasPrefix(result, "'") || !strings.HasSuffix(result, "'") {
		t.Errorf("expected string to be single-quoted, got %s", result)
	}
	resWithoutQuotes := strings.Trim(result, "'")
	for i := 0; i < len(resWithoutQuotes); i++ {
		if !strings.ContainsRune(OracleCharSet, rune(resWithoutQuotes[i])) {
			t.Errorf("expected string to contain only characters from OracleCharSet, got %s", resWithoutQuotes)
			return
		}
	}
}
func TestOracleDataProvider_GenBool(t *testing.T) {
	provider := NewOracleDataProvider()
	foundTrue := false
	foundFalse := false
	// Run enough times to likely hit both branches
	for i := 0; i < 100; i++ {
		val := provider.GenBool()
		if val == "TRUE" {
			foundTrue = true
		} else if val == "FALSE" {
			foundFalse = true
		} else {
			t.Errorf("unexpected value: %s", val)
		}
		if foundTrue && foundFalse {
			break
		}
	}
	if !foundTrue || !foundFalse {
		t.Errorf("GenBool did not return both TRUE and FALSE")
	}
}
func TestOracleDataProvider_GenNumber(t *testing.T) {
	provider := NewOracleDataProvider()
	val := provider.GenNumber(5)
	valNum, err := strconv.Atoi(val)
	if err != nil {
		t.Errorf("expected numeric string, got %s", val)
	}
	if valNum < 0 || valNum > 10000 {
		t.Errorf("expected number to be in range 0-10000, got %s", val)
	}
}

func TestOracleDataProvider_GenFloat(t *testing.T) {
	provider := NewOracleDataProvider()
	val := provider.GenFloat()
	_, err := strconv.ParseFloat(val, 64)
	if err != nil {
		t.Errorf("expected float string, got %s", val)
	}
}

func TestOracleDataProvider_GenUUid(t *testing.T) {
	provider := NewOracleDataProvider()
	val := provider.GenUUid()
	if !strings.HasPrefix(val, "'") || !strings.HasSuffix(val, "'") {
		t.Errorf("expected UUID to be single-quoted, got %s", val)
	}
	trimmed := strings.Trim(val, "'")
	if len(trimmed) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(trimmed))
	}
}

func TestOracleDataProvider_GenTs(t *testing.T) {
	provider := NewOracleDataProvider()

	valNow := provider.GenTs(true)
	if valNow != "CURRENT_TIMESTAMP" {
		t.Errorf("expected CURRENT_TIMESTAMP function, got %s", valNow)
	}

	valRandom := provider.GenTs(false)
	if !strings.HasPrefix(valRandom, "'") || !strings.HasSuffix(valRandom, "'") {
		t.Errorf("expected timestamp to be single-quoted, got %s", valRandom)
	}
	trimmed := strings.Trim(valRandom, "'")
	_, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		t.Errorf("expected valid timestamp string, got %s", trimmed)
	}
}
