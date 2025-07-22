package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"thermal/replcmd/registry"
	"thermal/session"
)

func Start(s *session.Session) {
	fmt.Println("REPL started. Type 'exit' to quit.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("\nbye") // Ctrl+D で退出
			break
		} else if err != nil {
			fmt.Println("input error:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("bye")
			break
		}

		registry.Execute(input, s)
	}
}
