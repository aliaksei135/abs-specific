package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmdtest"
)

func TestCLI(t *testing.T) {
	ts, err := cmdtest.Read(".")
	if err != nil {
		t.Fatal(err)
	}

	if err := exec.Command("go", "build", ".").Run(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("abs-specific")

	ts.Commands["abs-specific"] = cmdtest.Program("abs-specific")
	ts.Run(t, false)
}
