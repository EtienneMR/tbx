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
	Names    []string
	Args     []string
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
	return &Process{Names: names}
}

func NewResolved(name string, path string, args ...string) *Process {
	return &Process{
		Names:    []string{name},
		Resolved: Resolved{Name: name, Path: path},
		Args:     args,
	}
}

func (c *Process) Sudo() *Process {
	tui.Check(sudo.Resolve(), "process.Sudo")
	tui.Check(c.Resolve(), "process.Sudo")

	args := append([]string{c.Resolved.Path}, c.Args...)
	return NewResolved(sudo.Resolved.Name, sudo.Resolved.Path, args...)
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

func (c *Process) Command(args ...string) *exec.Cmd {
	base := append([]string(nil), c.Args...)
	all := append(base, args...)
	tui.Check(c.Resolve(), "process.Command")
	return exec.Command(c.Resolved.Path, all...)
}

// Run executes the command with the provided arguments and returns
// combined stdout/stderr output.
func (c *Process) Run(args ...string) (*Result, error) {
	if err := c.Resolve(); err != nil {
		return nil, err
	}
	tui.Run("%s %s", c.Resolved.Name, strings.Join(args, " "))

	var stdout, stderr bytes.Buffer
	proc := c.Command(args...)
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
		tui.Fatal("%v", err)
	}
	return res
}

// Must is like Run but calls Check on result
func (c *Process) Must(args ...string) *Result {
	return Check(c.Run(args...))
}

// Output runs the command and returns trimmed stdout.
// Convenient for single-value reads like the current branch name.
func (c *Process) Output(args ...string) string {
	res := c.Must(args...)
	return res.Stdout
}

// Live runs the command with stdin/stdout/stderr wired directly to the
// controlling terminal. Use this for long-running commands where real-time
// output matters (git clone, npm install, make …).
func (c *Process) Live(args ...string) {
	tui.Check(c.Resolve(), "process.Live")
	tui.Run("%s %s", c.Resolved.Name, strings.Join(args, " "))

	proc := c.Command(args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	tui.Check(proc.Run(), "process.Live: %q", c.Resolved.Name)
}

func (c *Process) find() (Resolved, error) {
	var errs []string
	for _, name := range c.Names {
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
