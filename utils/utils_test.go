package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintMarkdownTable(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	headers := []string{"Name", "Role"}
	rows := [][]string{
		{"Alice", "Developer"},
		{"Bob|Builder", "Lead"},
	}

	printMarkdownTable(headers, rows)
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "| Name | Role |") {
		t.Errorf("missing header in markdown table: %s", out)
	}
	if !strings.Contains(out, "| Bob\\|Builder | Lead |") {
		t.Errorf("missing escaped content in markdown table: %s", out)
	}
}
