package registry

import (
	"fmt"
	"strings"
	"thermal/replcmd/definitions"
	"thermal/replcmd/dts"
	"thermal/replcmd/elements"
	"thermal/replcmd/facts"
	"thermal/replcmd/labels"
	"thermal/replcmd/presentations"
	"thermal/replcmd/references"
	"thermal/replcmd/roletypes"
	"thermal/session"
)

type Command interface {
	Execute(*session.Session, string)
}

var commandMap = map[string]Command{}

func RegisterAll() {
	commandMap["elements"] = elements.New()
	commandMap["facts"] = facts.New()
	commandMap["labels"] = labels.New()
	commandMap["references"] = references.New()
	commandMap["presentations"] = presentations.New()
	commandMap["definitions"] = definitions.New()
	commandMap["dts"] = dts.New()
	commandMap["roletypes"] = roletypes.New()

	// エイリアス
	commandMap["pr"] = commandMap["presentations"]
	commandMap["df"] = commandMap["definitions"]
	commandMap["rt"] = commandMap["roletypes"]
	commandMap["el"] = commandMap["elements"]
	commandMap["lb"] = commandMap["labels"]
	commandMap["rf"] = commandMap["references"]
}

func Execute(input string, s *session.Session) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return
	}

	cmd, args := fields[0], strings.Join(fields[1:], " ")
	if c, ok := commandMap[cmd]; ok {
		c.Execute(s, args)
	} else {
		fmt.Fprintln(s.Stderr, "Unknown command:", cmd)
	}
}
