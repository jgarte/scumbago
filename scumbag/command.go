package scumbag

import (
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

// Command is an interface containing methods each command should have.
type Command interface {
	Run(args ...string)
}

// BaseCommand contains common functions for all commands.
type BaseCommand struct{}

// Channel returns the channel name from `line`.
func (cmd *BaseCommand) Channel(line *irc.Line) (string, error) {
	if len(line.Args) <= 0 {
		return "", fmt.Errorf("Line has no args: %v", line)
	}

	return line.Args[0], nil
}
