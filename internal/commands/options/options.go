package options

import (
	"fmt"
	"strings"
)

// Option represents a command option
type Option struct {
	Name         string
	Description  string
	Incompatible []string // List of option names that are incompatible with this option
}

// Options represents a set of command options
type Options struct {
	options map[string]Option
	active  map[string]bool
}

// NewOptions creates a new Options instance
func NewOptions() *Options {
	return &Options{
		options: make(map[string]Option),
		active:  make(map[string]bool),
	}
}

// RegisterOption registers a new option with its incompatibility rules
func (o *Options) RegisterOption(name, description string, incompatible []string) {
	o.options[name] = Option{
		Name:         name,
		Description:  description,
		Incompatible: incompatible,
	}
}

// Set activates an option
func (o *Options) Set(name string) error {
	name = strings.ToUpper(name)
	if _, exists := o.options[name]; !exists {
		return fmt.Errorf("unknown option: %s", name)
	}

	// Check incompatibility with already active options
	for activeOpt := range o.active {
		for _, incompatible := range o.options[name].Incompatible {
			if strings.ToUpper(incompatible) == activeOpt {
				return fmt.Errorf("option %s is incompatible with %s", name, activeOpt)
			}
		}
	}

	o.active[name] = true
	return nil
}

// IsSet checks if an option is active
func (o *Options) IsSet(name string) bool {
	return o.active[strings.ToUpper(name)]
}

// Clear clears all active options
func (o *Options) Clear() {
	o.active = make(map[string]bool)
}

// GetActive returns all active options
func (o *Options) GetActive() []string {
	active := make([]string, 0, len(o.active))
	for opt := range o.active {
		active = append(active, opt)
	}
	return active
}
