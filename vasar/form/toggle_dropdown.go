package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// newToggleDropdown returns a new form dropdown with the options Enabled and Disabled.
func newToggleDropdown(name string, enabled bool) form.Dropdown {
	return form.NewDropdown(name, []string{
		text.Colourf("<green>Enabled</green>"),
		text.Colourf("<red>Disabled</red>"),
	}, boolIndex(enabled))
}

// boolIndex converts a boolean to an index.
func boolIndex(b bool) int {
	if b {
		return 0
	}
	return 1
}

// indexBool converts an index to a bool.
func indexBool(dropdown form.Dropdown) bool {
	return dropdown.Value() == 0
}
