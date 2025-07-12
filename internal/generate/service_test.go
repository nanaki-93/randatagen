package generate

import (
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"testing"
)

// MockDataProvider implements DataProvider for testing
type MockDataProvider struct{}

func (m *MockDataProvider) GenString(length int) string { return "'mockstr'" }
func (m *MockDataProvider) GenBool() string             { return "true" }
func (m *MockDataProvider) GenNumber(length int) string { return "123" }
func (m *MockDataProvider) GenFloat() string            { return "1.23" }
func (m *MockDataProvider) GenUUid() string             { return "'mock-uuid'" }
func (m *MockDataProvider) GenTs(now bool) string       { return "'2024-01-01T00:00:00Z'" }
func (m *MockDataProvider) GetValueType(datatype string) (string, error) {
	switch datatype {
	case "int":
		return GetNumber, nil
	case "float":
		return GetFloat, nil
	case "bool":
		return GetBool, nil
	case "string":
		return GetString, nil
	case "uuid":
		return GetUuid, nil
	case "timestamp":
		return GetDateOrTs, nil
	default:
		return "", fmt.Errorf("unknown type")
	}
}

func TestWithSingleQuote(t *testing.T) {
	input := "abc"
	expected := "'abc'"
	if got := WithSingleQuote(input); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestWithDoubleQuote(t *testing.T) {
	input := "abc"
	expected := "\"abc\""
	if got := WithDoubleQuote(input); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGeneratorService_GetValue(t *testing.T) {
	provider := &MockDataProvider{}
	service := NewGeneratorService(provider)

	tests := []struct {
		datatype string
		length   int
		now      bool
		expected string
	}{
		{"int", 5, false, "123"},
		{"float", 0, false, "1.23"},
		{"bool", 0, false, "true"},
		{"string", 0, false, "'mockstr'"},
		{"uuid", 0, false, "'mock-uuid'"},
		{"timestamp", 0, false, "'2024-01-01T00:00:00Z'"},
		{"unknown", 0, false, "NotSupported"},
	}

	for _, tt := range tests {
		got := service.GetValue(tt.datatype, tt.length, tt.now)
		if got != tt.expected {
			t.Errorf("GetValue(%s) = %s, want %s", tt.datatype, got, tt.expected)
		}
	}
}

func TestGeneratorService_getValues(t *testing.T) {
	provider := &MockDataProvider{}
	service := NewGeneratorService(provider)
	columns := []model.Column{
		{Name: "col1", Datatype: "int", Length: 0},
		{Name: "col2", Datatype: "string", Length: 0},
	}
	got := service.getValues(columns)
	expected := "123, 'mockstr'"
	if got != expected {
		t.Errorf("getValues() = %s, want %s", got, expected)
	}
}

func TestGeneratorService_GenerateSql(t *testing.T) {
	provider := &MockDataProvider{}
	service := NewGeneratorService(provider)
	dataGen := model.RanData{
		Rows: 2,
		Columns: []model.Column{
			{Name: "col1", Datatype: "int", Length: 0},
			{Name: "col2", Datatype: "string", Length: 0},
		},
		Target: model.DbStruct{
			DbSchema: "myschema",
			DbTable:  "mytable",
		},
	}
	sqls := service.GenerateSql(dataGen)
	if len(sqls) != 1 {
		t.Errorf("expected 1 SQL statement, got %d", len(sqls))
	}
	if len(sqls) > 0 && sqls[0] == "" {
		t.Errorf("generated SQL is empty")
	}
}

func Test_getPrefixInsert(t *testing.T) {
	dataGen := model.RanData{
		Columns: []model.Column{
			{Name: "col1"},
			{Name: "col2"},
		},
		Target: model.DbStruct{
			DbSchema: "myschema",
			DbTable:  "mytable",
		},
	}
	got := getPrefixInsert(dataGen)
	expected := "Insert into myschema.\"mytable\"(col1, col2) values"
	if got != expected {
		t.Errorf("getPrefixInsert() = %s, want %s", got, expected)
	}
}
