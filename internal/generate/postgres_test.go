package generate

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPostgresDataProvider_GenString(t *testing.T) {
	provider := NewPostgresDataProvider()
	result := provider.GenString(10)
	if len(result) != 12 { // 10 chars + 2 single quotes
		t.Errorf("expected length 12, got %d", len(result))
	}
	if !strings.HasPrefix(result, "'") || !strings.HasSuffix(result, "'") {
		t.Errorf("expected string to be single-quoted, got %s", result)
	}
}

func TestPostgresDataProvider_GenBool(t *testing.T) {
	provider := NewPostgresDataProvider()
	foundTrue := false
	foundFalse := false
	// Run enough times to likely hit both branches
	for i := 0; i < 100; i++ {
		val := provider.GenBool()
		if val == "true" {
			foundTrue = true
		} else if val == "false" {
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

func TestPostgresDataProvider_GenNumber(t *testing.T) {
	provider := NewPostgresDataProvider()
	val := provider.GenNumber(5)
	if _, err := strconv.Atoi(val); err != nil {
		t.Errorf("expected numeric string, got %s", val)
	}
}

func TestPostgresDataProvider_GenFloat(t *testing.T) {
	provider := NewPostgresDataProvider()
	val := provider.GenFloat()
	if _, err := strconv.ParseFloat(val, 64); err != nil {
		t.Errorf("expected float string, got %s", val)
	}
}

func TestPostgresDataProvider_GenUUid(t *testing.T) {
	provider := NewPostgresDataProvider()
	val := provider.GenUUid()
	if !strings.HasPrefix(val, "'") || !strings.HasSuffix(val, "'") {
		t.Errorf("expected UUID to be single-quoted, got %s", val)
	}
	trimmed := strings.Trim(val, "'")
	if len(trimmed) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(trimmed))
	}
}

func TestPostgresDataProvider_GenTs(t *testing.T) {
	provider := NewPostgresDataProvider()
	valNow := provider.GenTs(true)
	if valNow != "now()" {
		t.Errorf("expected 'now()', got %s", valNow)
	}
	valRandom := provider.GenTs(false)
	if !strings.HasPrefix(valRandom, "'") || !strings.HasSuffix(valRandom, "'") {
		t.Errorf("expected timestamp to be single-quoted, got %s", valRandom)
	}
	trimmed := strings.Trim(valRandom, "'")
	if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
		t.Errorf("expected RFC3339 timestamp, got %s", trimmed)
	}
}

func TestPostgresDataProvider_GetValueType(t *testing.T) {
	provider := NewPostgresDataProvider()
	typ, err := provider.GetValueType("BIGINT")
	if err != nil || typ != GetNumber {
		t.Errorf("expected %s, got %s, err: %v", GetNumber, typ, err)
	}
	_, err = provider.GetValueType("UNKNOWN_TYPE")
	if err == nil {
		t.Errorf("expected error for unknown type")
	}
}
