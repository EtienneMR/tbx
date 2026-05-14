package texec

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/EtienneMR/tbx/tlog"
)

// Exec tries multiple command names in order and returns the first one found.
type Exec struct {
	Names    []string
	resolved Resolved
}

// Resolved wraps an executable found on PATH.
type Resolved struct {
	name string
	path string
}

// Result holds captured output from a completed command.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

func New(names ...string) *Exec {
	return &Exec{Names: names}
}

func (c *Exec) Resolve() error {
	if c.resolved.path == "" {
		r, err := c.find()
		if err != nil {
			return err
		}
		c.resolved = r
	}
	return nil
}

// Run executes the command with the provided arguments and returns
// combined stdout/stderr output.
func (c *Exec) Run(args ...string) (*Result, error) {
	tlog.Run("%s %s", c.resolved.name, strings.Join(args, " "))

	if err := c.Resolve(); err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	proc := exec.Command(c.resolved.path, args...)
	proc.Stdout = &stdout
	proc.Stderr = &stderr

	err := proc.Run()
	res := &Result{
		Stdout: strings.TrimRight(stdout.String(), "\n"),
		Stderr: strings.TrimRight(stderr.String(), "\n"),
	}
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			res.Code = exit.ExitCode()
		}
		return res, fmt.Errorf("%s: %w", c.resolved.name, err)
	}
	return res, nil
}

// Check calls tlog.Fatal on any error, printing stderr first.
func Check(res *Result, err error) *Result {
	if err != nil {
		if res.Stderr != "" {
			tlog.Indent(res.Stderr)
		}
		tlog.Fatal("%v", err)
	}
	return res
}

// Must is like Run but calls Check on result
func (c *Exec) Must(args ...string) *Result {
	return Check(c.Run(args...))
}

// Output runs the command and returns trimmed stdout.
// Convenient for single-value reads like the current branch name.
func (c *Exec) Output(args ...string) string {
	res := c.Must(args...)
	return res.Stdout
}

// Live runs the command with stdin/stdout/stderr wired directly to the
// controlling terminal. Use this for long-running commands where real-time
// output matters (git clone, npm install, make …).
//
// Unlike Run, output is not captured and cannot be indented — it flows
// to the terminal exactly as the child process writes it.
func (c *Exec) Live(args ...string) error {
	tlog.Run("%s %s", c.resolved.name, strings.Join(args, " "))

	if err := c.Resolve(); err != nil {
		return err
	}

	proc := exec.Command(c.resolved.path, args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	if err := proc.Run(); err != nil {
		return fmt.Errorf("%s: %w", c.resolved.name, err)
	}
	return nil
}

func (c *Exec) find() (Resolved, error) {
	var errs []string
	for _, name := range c.Names {
		path, err := exec.LookPath(name)
		if err == nil {
			return Resolved{name: name, path: path}, nil
		}
		errs = append(errs, fmt.Sprintf("%q: %v", name, err))
	}

	if len(errs) == 0 {
		return Resolved{}, errors.New("texec: no valid command names provided")
	}
	return Resolved{}, fmt.Errorf("texec: none of the candidate commands were found: %s", strings.Join(errs, "; "))
}
