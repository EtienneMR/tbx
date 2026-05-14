package tui

import (
	"charm.land/huh/v2"
)

// Confirm asks a yes/no question and returns true if the user confirms.
// The default value is returned when stdin is not a TTY (e.g. in scripts).
func Confirm(question string, defaultVal bool) (bool, error) {
	var answer bool = defaultVal
	err := huh.NewConfirm().
		Title(question).
		Value(&answer).
		Run()
	return answer, err
}

// Input asks for a single text value.
func Input(label, defaultVal string) (string, error) {
	var val string
	err :=
		huh.NewInput().
			Title(label).
			Placeholder(defaultVal).
			Value(&val).
			Run()

	if val == "" {
		return defaultVal, err
	}
	return val, err
}

// Select asks the user to pick one item from a list.
func Select(label string, options []string) (string, error) {
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}
	var chosen string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(label).
				Options(opts...).
				Value(&chosen),
		),
	).Run()
	return chosen, err
}
