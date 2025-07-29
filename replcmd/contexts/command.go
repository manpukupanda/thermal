package contexts

import (
	"flag"
	"fmt"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/session"

	"gopkg.in/yaml.v3"
)

type ContextsCommand struct{}

func New() *ContextsCommand {
	return &ContextsCommand{}
}

func parseArgs(args string) (string, bool, error) {
	fs := flag.NewFlagSet("facts", flag.ContinueOnError)
	cx := fs.String("c", "", "Pattern to match context IDs (* = any string)")
	ls := fs.Bool("l", false, "List role type URIs only")

	argv := strings.Fields(args)

	if err := fs.Parse(argv); err != nil {
		return "", false, err
	}

	if fs.NArg() > 0 {
		return "", false, fmt.Errorf("unknown parameter: %v", fs.Args())
	}

	return *cx, *ls, nil
}

type OutputContext struct {
	ID       string         `yaml:"ID"`
	Entity   model.Entity   `yaml:"Entity"`
	Period   model.Period   `yaml:"Period"`
	Scenario model.Scenario `yaml:"Scenario"`
}

func (c *ContextsCommand) Execute(s *session.Session, args string) {
	if s.Instance == nil {
		return
	}

	cxPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	var outputContexts []OutputContext

	for _, context := range s.Instance.Contexts {
		if cxPattern != "" && !parser.WildcardMatch(cxPattern, context.ID) {
			continue
		}

		if ls {
			fmt.Fprintln(s.Stdout, context.ID)
		} else {
			outCxt := OutputContext{
				ID:       context.ID,
				Entity:   context.Entity,
				Period:   context.Period,
				Scenario: context.Scenario,
			}
			outputContexts = append(outputContexts, outCxt)

			encoder := yaml.NewEncoder(s.Stdout)
			encoder.SetIndent(2) // 読みやすさのためにインデント設定

			if err := encoder.Encode(outputContexts); err != nil {
				fmt.Fprintf(s.Stderr, "YAML encode error: %v\n", err)
			}
		}
	}
}
