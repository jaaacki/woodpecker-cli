package output

import (
	"bytes"
	"testing"
)

func TestNewContextDefaults(t *testing.T) {
	ctx := NewContext()
	if ctx.JSON || ctx.Raw {
		t.Fatal("default context should not enable json or raw")
	}
}

func TestContextDataJSON(t *testing.T) {
	var buf bytes.Buffer
	ctx := Context{JSON: true, Out: &buf}
	ctx.Data(map[string]string{"hello": "world"})
	out := buf.String()
	if !contains(out, `"ok": true`) {
		t.Fatalf("expected ok=true envelope, got: %s", out)
	}
	if !contains(out, `"hello": "world"`) {
		t.Fatalf("expected data in envelope, got: %s", out)
	}
}

func TestContextPrintlnSkippedInJSONMode(t *testing.T) {
	var buf bytes.Buffer
	ctx := Context{JSON: true, Out: &buf}
	ctx.Println("should not appear")
	if buf.Len() != 0 {
		t.Fatalf("text output should be suppressed in json mode, got: %s", buf.String())
	}
}

func TestContextTableSkippedInJSONMode(t *testing.T) {
	var buf bytes.Buffer
	ctx := Context{JSON: true, Out: &buf}
	ctx.PrintTable([]string{"A"}, [][]string{{"b"}})
	if buf.Len() != 0 {
		t.Fatalf("table output should be suppressed in json mode, got: %s", buf.String())
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
