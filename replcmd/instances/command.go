package instances

import (
	"fmt"
	"strconv"
	"thermal/session"
)

type InstancesCommand struct{}

func New() *InstancesCommand {
	return &InstancesCommand{}
}

func toValidIndex(len int, input string) int {
	n, err := strconv.Atoi(input)
	if err != nil {
		return -1 // 数値に変換できない場合
	}
	if n >= 1 && n <= len {
		return n - 1
	}
	return -1
}

func (c *InstancesCommand) Execute(s *session.Session, args string) {
	if s.Manifest != nil {
		if args != "" {
			p := toValidIndex(len(s.Manifest.List.XBRLInstances), args)
			if p == -1 {
				fmt.Fprintln(s.Stdout, "invalid number:", args)
			} else {
				s.Instance = s.Manifest.List.XBRLInstances[p]
			}
		}
		for i, instance := range s.Manifest.List.XBRLInstances {
			prefix := " "
			if instance.Path == s.Instance.Path {
				prefix = "*"
			}
			msg := fmt.Sprintf("%s %d) %s", prefix, i+1, instance.Path)
			fmt.Fprintln(s.Stdout, msg)
		}

	} else if s.Instance != nil {
		msg := fmt.Sprintf("* %s\n", s.Instance.Path)
		fmt.Fprintln(s.Stdout, msg)
	} else {
		fmt.Fprintln(s.Stdout, "no instance.")
	}
}
