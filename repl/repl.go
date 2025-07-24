package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"thermal/replcmd/registry"
	"thermal/session"

	"golang.org/x/term"
)

func Start(s *session.Session) {
	// 対話モードか判定
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))
	if isTerminal {
		fmt.Fprintln(s.Stdout, "thermal started. Type 'exit' to quit.")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		if isTerminal {
			fmt.Fprint(s.Stdout, ">>> ")
		}
		input, err := reader.ReadString('\n')
		if err == io.EOF {
			if isTerminal {
				fmt.Fprintln(s.Stdout, "\nbye")
			}
			break
		} else if err != nil {
			fmt.Fprintln(s.Stderr, "input error:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			if isTerminal {
				fmt.Fprintln(s.Stdout, "bye")
			}
			break
		}

		registry.Execute(input, s)
	}
}
