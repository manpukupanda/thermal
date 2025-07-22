package repl

import (
	"fmt"
	"thermal/session"
)

type Command interface {
	Execute(*session.Session, string)
}

var commandRegistry = map[string]Command{}

func RegisterCommand(name string, cmd Command) {
	commandRegistry[name] = cmd
}

func ExecuteCommand(name string, s *session.Session, args string) {
	if cmd, ok := commandRegistry[name]; ok {
		cmd.Execute(s, args)
	} else {
		fmt.Println("unknown command:", name)
	}
}
