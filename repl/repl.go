package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"thermal/replcmd/registry"
	"thermal/session"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

func Start(s *session.Session) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		startInteractive(s)
	} else {
		startNonInteractive(s)
	}
}

// 終了コマンド
var exitCommands = []string{"exit", "quit", "bye"}

func startInteractive(s *session.Session) {
	historyPath := filepath.Join(os.TempDir(), "thermal_history.tmp")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     historyPath,
		InterruptPrompt: "^C",
		EOFPrompt:       "bye",
		Stdin:           io.NopCloser(s.Stdin),
		Stdout:          s.Stdout,
		Stderr:          s.Stderr,
	})
	if err != nil {
		fmt.Fprintln(s.Stderr, "readline init error:", err)
		return
	}
	defer rl.Close()

	if s.Manifest != nil {
		fmt.Fprintln(s.Stdout, "[manifest]")
		fmt.Fprintln(s.Stdout, "*", s.Manifest.Path)
	}
	if s.Instance != nil {
		fmt.Fprintln(s.Stdout, "[instance(s)]")
		for _, instance := range s.Manifest.List.XBRLInstances {
			prefix := " "
			if instance.Path == s.Instance.Path {
				prefix = "*"
			}
			msg := fmt.Sprintf("%s %s", prefix, instance.Path)
			fmt.Fprintln(s.Stdout, msg)
		}
	}
	fmt.Fprintln(s.Stdout, "")
	fmt.Fprintln(s.Stdout, "thermal started. Type 'exit' to quit.")

	for {
		input, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			fmt.Fprintln(s.Stdout, rl.Config.EOFPrompt)
			break
		} else if err != nil {
			fmt.Fprintln(s.Stderr, "input error:", err)
			continue
		}

		input = strings.TrimSpace(input)

		if slices.Contains(exitCommands, input) {
			fmt.Fprintln(s.Stdout, rl.Config.EOFPrompt)
			break
		}

		registry.Execute(input, s)
	}
}

func startNonInteractive(s *session.Session) {
	reader := bufio.NewReader(s.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(s.Stderr, "input error:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if slices.Contains(exitCommands, input) {
			break
		}

		registry.Execute(input, s)
	}
}
