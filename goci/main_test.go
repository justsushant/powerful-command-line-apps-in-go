package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var testCases = []struct{
		name string
		proj string
		out string
		expErr error
		setupGit bool
		mockCmd func(ctx context.Context, name string, arg... string) *exec.Cmd
	}{
		{name: "success", proj: "./testdata/tool", out: "Go Build: SUCCESS\nGo Test: SUCCESS\nGofmt: SUCCESS\nGolangCI-lint: SUCCESS\nGit Push: SUCCESS\n", expErr: nil, setupGit: true, mockCmd: nil},
		{name: "successMock", proj: "./testdata/tool", out: "Go Build: SUCCESS\nGo Test: SUCCESS\nGofmt: SUCCESS\nGolangCI-lint: SUCCESS\nGit Push: SUCCESS\n", expErr: nil, setupGit: false, mockCmd: mockCmdContext},
		{name: "fail", proj: "./testdata/toolErr", out: "", expErr: &stepErr{step: "go build"}, setupGit: false, mockCmd: nil},
		{name: "failFormat", proj: "./testdata/toolFmtErr", out: "", expErr: &stepErr{step: "go fmt"}, setupGit: false, mockCmd: nil},
		{name: "failTimeout", proj: "./testdata/tool", out: "", expErr: context.DeadlineExceeded, setupGit: false, mockCmd: mockCmdTimeout},
		// {name: "failGolangci-lint", proj: "./testdata/toolLintErr", out: "", expErr: &stepErr{step: "golangci-lint run"}, setupGit: false, mockCmd: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupGit {
				_, err := exec.LookPath("git")
				if err != nil {
					t.Skip("Git not installed. Skipping test.")
				}

				cleanup := setupGit(t, tc.proj)
				defer cleanup()
			}

			if tc.mockCmd != nil {
				command = tc.mockCmd
			}

			var out bytes.Buffer

			err := run(tc.proj, &out)

			// if an error is expected
			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q. Got 'nil' instead.", tc.expErr)
					return
				}

				// got some other error instead of the one expected
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q.", tc.expErr, err)
				}

				// test success, got the intended error
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			if out.String() != tc.out {
				t.Errorf("Expected output: %q. Got %q", tc.out, out.String())
			}
		})
	}
}

func TestRunKill(t *testing.T) {
	//RunKill test cases
	var testCases = []struct {
		name string
		proj string
		sig syscall.Signal
		expErr error
	}{
		{"SIGINT", "./testdata/tool", syscall.SIGINT, ErrSignal},
		{"SIGTERM", "./testdata/tool", syscall.SIGTERM, ErrSignal},
		{"SIGQUIT", "./testdata/tool", syscall.SIGQUIT, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command = mockCmdTimeout

			errCh := make(chan error)
			ignSigCh := make(chan os.Signal, 1)
			expSigCh := make(chan os.Signal, 1)

			signal.Notify(ignSigCh, syscall.SIGQUIT)
			defer signal.Stop(ignSigCh)

			signal.Notify(expSigCh, tc.sig)
			defer signal.Stop(expSigCh)

			go func() {
				errCh <- run(tc.proj, io.Discard)
			}()

			go func() {
				time.Sleep(2 * time.Second)
				syscall.Kill(syscall.Getpid(), tc.sig)
			}()

			select {
			case err := <-errCh:
				if err == nil {
					t.Errorf("Expected error. Got 'nil' instead")
					return
				}
				
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expErr, err)
				}
			select {
			case rec := <-expSigCh:
				if rec != tc.sig {
					t.Errorf("Expected signal %q, got %q", tc.sig, rec)
				}
			default:
				t.Errorf("Signal not received")
			}
			case <-ignSigCh:

			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}

	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "Everything up-to-date")
		os.Exit(0)
	}

	os.Exit(1)
}

func setupGit(t *testing.T, proj string) func() {
	t.Helper()

	// checking whether git command is available or not
	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}

	// temp directory for the simulated remote git repo
	tempDir, err := os.MkdirTemp("", "gocitest")
	if err != nil {
		t.Fatal(err)
	}

	// full path of the target project directory
	projPath, err := filepath.Abs(proj)
	if err != nil {
		t.Fatal(err)
	}

	// remoteURI address in the local file system
	remoteURI := fmt.Sprintf("file://%s", tempDir)

	// commands to set up git for test
	var gitCmdList = []struct{
		args []string
		dir string
		env []string
	}{
		{[]string{"init", "-bare"}, tempDir, nil},
		{[]string{"init"}, projPath, nil},
		{[]string{"remote", "add", "origin", remoteURI}, projPath, nil},
		{[]string{"add", "."}, projPath, nil},
		{[]string{"commit", "-m", "test"}, projPath, []string{
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
		}},
	}

	// running each git command
	for _, g := range gitCmdList {
		gitCmd := exec.Command(gitExec, g.args...)
		gitCmd.Dir = g.dir

		if g.env != nil {
			gitCmd.Env = append(os.Environ(), g.env...)
		}

		if err := gitCmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	return func() {
		os.RemoveAll(tempDir)
		os.RemoveAll(filepath.Join(projPath, ".git"))
	}
}

func mockCmdContext(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")
	return cmd
}