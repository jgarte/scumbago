package scumbag

import (
	"errors"
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

type Command interface {
	Run(args ...string)
}

type BaseCommand struct{}

// Returns the channel name from `line`.
func (cmd *BaseCommand) Channel(line *irc.Line) (string, error) {
	if len(line.Args) <= 0 {
		return "", errors.New(fmt.Sprintf("Line has no args: %v", line))
	}

	return line.Args[0], nil
}
