package luacats

import (
	"go/parser"
	"testing"
)

func TestGoTypeToLua(t *testing.T) {
	tests := []struct {
		goType string
		want   string
	}{
		{"int", "integer"},
		{"int8", "integer"},
		{"int16", "integer"},
		{"int32", "integer"},
		{"int64", "integer"},
		{"uint", "integer"},
		{"uint8", "integer"},
		{"uint16", "integer"},
		{"uint32", "integer"},
		{"uint64", "integer"},
		{"rune", "integer"},
		{"byte", "integer"},
		{"float32", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"string", "string"},
		{"error", "any"},
		{"Point", "Point"},
		{"[]Point", "Point[]"},
		{"[3]int", "integer[]"},
		{"map[string]int", "table<string,integer>"},
		{"*Point", "Point"},
		{"pkg.Point", "Point"},
		{"interface{}", "any"},
		{"chan int", "any"},
		{"map[string][]*pkg.Point", "table<string,Point[]>"},
	}

	for _, test := range tests {
		t.Run(test.goType, func(t *testing.T) {
			expr, err := parser.ParseExpr(test.goType)
			if err != nil {
				t.Fatalf("ParseExpr(%q): %v", test.goType, err)
			}

			if got := goTypeToLua(expr); got != test.want {
				t.Errorf("goTypeToLua(%q) = %q, want %q", test.goType, got, test.want)
			}
		})
	}
}
