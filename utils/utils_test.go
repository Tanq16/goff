package utils

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestEscapeCells(t *testing.T) {
	in := []string{"Hello|World", "Normal", "Pipe|Another|Pipe"}
	got := escapeCells(in)
	want := []string{"Hello\\|World", "Normal", "Pipe\\|Another\\|Pipe"}
	if !slices.Equal(got, want) {
		t.Fatalf("escapeCells() = %v, want %v", got, want)
	}
}

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

func TestPromptSelectForAI(t *testing.T) {
	GlobalForAIFlag = true
	defer func() { GlobalForAIFlag = false }()

	tests := []struct {
		name    string
		options []string
		input   string
		want    int
	}{
		{"valid choice", []string{"opt1", "opt2", "opt3"}, "2\n", 1},
		{"out of bounds", []string{"opt1", "opt2"}, "5\n", -1},
		{"empty input", []string{"opt1"}, "\n", -1},
		{"no options", nil, "1\n", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			w.WriteString(tt.input)
			w.Close()
			os.Stdin = r
			stdinScanner = nil
			defer func() {
				os.Stdin = oldStdin
				stdinScanner = nil
			}()

			got, err := PromptSelect(tt.name, tt.options)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Errorf("PromptSelect() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPromptMultiSelectForAI(t *testing.T) {
	GlobalForAIFlag = true
	defer func() { GlobalForAIFlag = false }()

	tests := []struct {
		name    string
		options []string
		input   string
		want    []int
	}{
		{"multi choice", []string{"a", "b", "c"}, "1, 3\n", []int{0, 2}},
		{"none choice", []string{"a", "b"}, "none\n", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			w.WriteString(tt.input)
			w.Close()
			os.Stdin = r
			stdinScanner = nil
			defer func() {
				os.Stdin = oldStdin
				stdinScanner = nil
			}()

			got, err := PromptMultiSelect(tt.name, tt.options)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("got %v, want nil/empty", got)
				}
				return
			}
			for _, exp := range tt.want {
				if !got[exp] {
					t.Errorf("expected index %d selected in %v", exp, got)
				}
			}
		})
	}
}
