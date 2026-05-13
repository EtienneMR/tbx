# tbx — personal toolbox CLI

A fast, opinionated CLI for everyday dev tasks.

```
tbx git start [feature]   create + checkout a feature branch
tbx git snap              stage everything and commit
tbx system status         check toolchain presence
```

## Install

```bash
# From source
git clone https://github.com/EtienneMR/tbx
cd tbx
make install        # installs to ~/go/bin/tbx
```

## Verbosity

Add `-v` to see every shell command that runs, `-vv` for debug output:

```bash
tbx -v git start my-feature
tbx -vv git snap -m "wip: progress"
```

## Project layout

```
tbx/
├── main.go               entrypoint
├── cmd/
│   ├── root.go           global flags & subcommand registration
│   ├── git/
│   │   ├── git.go        `tbx git` group
│   │   ├── start.go      `tbx git start`
│   │   └── snap.go       `tbx git snap`
│   └── system/
│       ├── system.go     `tbx system` group
│       └── status.go     `tbx system status`
└── internal/
    ├── tlog/tlog.go      logging with colored icons
    ├── run/run.go        shell command execution
    └── ui/prompt.go      interactive prompts (huh)
```

## Adding a new command

1. Create `cmd/<group>/<verb>.go` with a `new<Verb>Cmd() *cobra.Command` function.
2. Register it in `cmd/<group>/<group>.go` via `cmd.AddCommand(new<Verb>Cmd())`.
3. Use `tlog`, `run`, and `ui` packages — no `fmt.Println` in command files.

## Build with version injection

```bash
git tag v0.1.0
make build     # produces bin/tbx reporting version v0.1.0
```
