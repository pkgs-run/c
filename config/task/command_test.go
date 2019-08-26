package task

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	yaml "gopkg.in/yaml.v2"
)

func TestCommand_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want Command
	}{
		{
			"short-command",
			`example`,
			Command{
				Do:    "example",
				Print: "example",
			},
		},
		{
			"do-no-echo",
			`do: example`,
			Command{
				Do:    "example",
				Print: "example",
			},
		},
		{
			"command-with-print",
			`{do: something, print: echo example}`,
			Command{
				Do:    "something",
				Print: "echo example",
			},
		},
		{
			"many-fields",
			`{do: dovalue, print: printvalue, dir: dirvalue}`,
			Command{
				Do:    "dovalue",
				Print: "printvalue",
				Dir:   "dirvalue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Command

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}

// TestCommand_exec_helper is a helper test that is called when mocking exec.
//
// The following environment variables can configure this function:
//
// - TUSK_WANT_TEST_COMMAND: Set to "1" to run this function.
// - TUSK_TEST_COMMAND_ARGS: Set to a comma-separated list of expected command
//   arguments.
// - TUSK_TEST_COMMAND_DIR: Set to the expected directory
func TestCommand_exec_helper(t *testing.T) {
	if os.Getenv("TUSK_WANT_TEST_COMMAND") != "1" {
		return
	}

	wantArgs := strings.Split(os.Getenv("TUSK_TEST_COMMAND_ARGS"), ",")
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if diff := cmp.Diff(wantArgs, args); diff != "" {
		t.Errorf("arguments differ:\n%s", diff)
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get working dir: ", err)
	}

	wantDir := os.Getenv("TUSK_TEST_COMMAND_DIR")
	if wantDir != "" && dir != wantDir {
		t.Errorf("want working dir %s, got %s", wantDir, dir)
	}
}

func TestCommand_exec(t *testing.T) {
	wantCommand := "echo hello world"
	wantArgs := strings.Join([]string{getShell(), "-c", wantCommand}, ",")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Dir(wd)

	command := Command{
		Do:  wantCommand,
		Dir: "..",
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestCommand_exec_helper", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...) // nolint: gosec
		cmd.Env = []string{
			"TUSK_WANT_TEST_COMMAND=1",
			"TUSK_TEST_COMMAND_ARGS=" + wantArgs,
			"TUSK_TEST_COMMAND_DIR=" + wantDir,
		}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	if err := command.exec(); err != nil {
		t.Fatal(err)
	}
}

func TestCommandList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want CommandList
	}{
		{
			"single-short-command",
			`example`,
			CommandList{
				{Do: "example", Print: "example"},
			},
		},
		{
			"list-short-commands",
			`[one,two]`,
			CommandList{
				{Do: "one", Print: "one"},
				{Do: "two", Print: "two"},
			},
		},
		{
			"single-do-command",
			`do: example`,
			CommandList{
				{Do: "example", Print: "example"},
			},
		},
		{
			"list-do-commands",
			`[{do: one},{do: two}]`,
			CommandList{
				{Do: "one", Print: "one"},
				{Do: "two", Print: "two"},
			},
		},
		{
			"empty-list",
			`[]`,
			CommandList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CommandList

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}

func TestGetShell(t *testing.T) {
	originalShell := os.Getenv(shellEnvVar)
	defer func() {
		if err := os.Setenv(shellEnvVar, originalShell); err != nil {
			t.Errorf("Failed to reset SHELL environment variable: %v", err)
		}
	}()

	customShell := "/my/custom/sh"
	if err := os.Setenv(shellEnvVar, customShell); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	if actual := getShell(); actual != customShell {
		t.Errorf("getShell(): expected %v, actual %v", customShell, actual)
	}

	if err := os.Unsetenv(shellEnvVar); err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}

	if actual := getShell(); actual != defaultShell {
		t.Errorf("getShell(): expected %v, actual %v", defaultShell, actual)
	}
}
