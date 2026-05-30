package pkg

import (
	"strings"

	"github.com/EtienneMR/tbx/internal/process"
	"github.com/EtienneMR/tbx/internal/tui"
	"github.com/spf13/cobra"
)

const noFiles = cobra.ShellCompDirectiveNoFileComp

func completeLocal(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res, err := pacman.Run("-Ssq", toComplete)
	if err != nil || res.Stdout == "" {
		if err != nil {
			tui.Error("%s", err)
		}
		return nil, noFiles
	}

	already := sliceToSet(args)
	var out []string
	for name := range strings.FieldsSeq(res.Stdout) {
		if strings.Contains(name, toComplete) && !already[name] {
			out = append(out, name)
		}
	}
	return out, noFiles
}

func completeInstalled(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res, err := pm.Run("-Qqe")
	if err != nil || res.Stdout == "" {
		if err != nil {
			tui.Error("%s", err)
		}
		return nil, noFiles
	}

	already := sliceToSet(args)
	var out []string
	for name := range strings.FieldsSeq(res.Stdout) {
		if strings.Contains(name, toComplete) && !already[name] {
			out = append(out, name)
		}
	}
	return out, noFiles
}

func completeOrphans(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res, err := pm.Run("-Qdttq")
	if process.IsErrorCode(err, 1) {
		return nil, noFiles
	}
	process.Check(res, err)

	already := sliceToSet(args)
	var out []string
	for name := range strings.FieldsSeq(res.Stdout) {
		if strings.Contains(name, toComplete) && !already[name] {
			out = append(out, name)
		}
	}
	return out, noFiles
}

func sliceToSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}
