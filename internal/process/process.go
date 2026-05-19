package process

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/EtienneMR/tbx/internal/tui"
)

var sudo = New("sudo")

// Process tries multiple command names in order and returns the first one found.
type Process struct {
	names    []string
	args     []string
	Resolved Resolved
}

// Resolved wraps an executable found on PATH.
type Resolved struct {
	Name string
	Path string
}

// Result holds captured output from a completed command.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

func New(names ...string) *Process {
	return &Process{names: names}
}

func NewResolved(name string, path string, args ...string) *Process {
	if path == "" {
		panic("NewResolved with unresolved path")
	}

	return &Process{
		names:    []string{name},
		Resolved: Resolved{Name: name, Path: path},
		args:     args,
	}
}

func (c *Process) AddArgs(args ...string) *Process {
	if len(args) == 0 {
		return c
	}

	final := make([]string, 0, len(c.args)+len(args))
	final = append(final, c.args...)
	final = append(final, args...)

	return &Process{
		names:    c.names,
		Resolved: c.Resolved,
		args:     final,
	}
}

func (c *Process) Sudo() *Process {
	tui.Check(c.Resolve(), "process.Sudo")

	return sudo.AddArgs("--", c.Resolved.Path).AddArgs(c.args...)
}

func (c *Process) Resolve() error {
	if c.Resolved.Path == "" {
		r, err := c.find()
		if err != nil {
			return err
		}
		c.Resolved = r
	}
	return nil
}

func (c *Process) Command() *exec.Cmd {
	tui.Check(c.Resolve(), "process.Command")
	tui.Run("%s %s", c.Resolved.Path, strings.Join(c.args, " "))

	return exec.Command(c.Resolved.Path, c.args...)
}

// Run executes the command with the provided arguments and returns
// combined stdout/stderr output.
func (c *Process) Run(args ...string) (*Result, error) {
	if err := c.Resolve(); err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	proc := c.AddArgs(args...).Command()
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
		return res, fmt.Errorf("%s: %w", c.Resolved.Name, err)
	}
	return res, nil
}

// Check calls tui.Fatal on any error, printing stderr first.
func Check(res *Result, err error) *Result {
	if err != nil {
		if res.Stderr != "" {
			tui.Indent(res.Stderr)
		}
		tui.Fatal("%s", err)
	}
	return res
}

// Must is like Run but calls Check on result
func (c *Process) Must(args ...string) *Result {
	return Check(c.Run(args...))
}

// Output runs the command and returns trimmed stdout.
func (c *Process) Output(args ...string) string {
	res := c.Must(args...)
	return res.Stdout
}

// Test runs the command and test if return code is zero
func (c *Process) Test(errorCode int, args ...string) bool {
	res, err := c.Run(args...)

	if IsErrorCode(err, errorCode) {
		return false
	}

	Check(res, err)
	return true
}

func (c *Process) LiveUnchecked(args ...string) error {
	if err := c.Resolve(); err != nil {
		return err
	}

	proc := c.AddArgs(args...).Command()
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	return proc.Run()
}

// Live runs the command with stdin/stdout/stderr wired directly to the controlling terminal.
func (c *Process) Live(args ...string) {
	tui.Check(c.LiveUnchecked(args...), "process.Live: %q", c.Resolved.Name)
}

func (c *Process) find() (Resolved, error) {
	var errs []string
	for _, name := range c.names {
		path, err := exec.LookPath(name)
		if err == nil {
			return Resolved{Name: name, Path: path}, nil
		}
		errs = append(errs, fmt.Sprintf("%q: %v", name, err))
	}

	if len(errs) == 0 {
		return Resolved{}, errors.New("find: no valid command names provided")
	}
	return Resolved{}, fmt.Errorf("find: none of the candidate commands were found: %s", strings.Join(errs, "; "))
}

func IsErrorCode(err error, code int) bool {
	exit, ok := errors.AsType[*exec.ExitError](err)

	return ok && exit.ExitCode() == code
}
