package cmd

import (
	"bytes"
	"testing"
)

func TestRootHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected help error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Errorf("expected help output, got empty")
	}
}

func TestInspectSubcommandRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "inspect" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("inspect command not registered on rootCmd")
	}
}

func TestBatchSubcommandRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "batch" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("batch command not registered on rootCmd")
	}
}
