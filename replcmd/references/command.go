package references

import (
	"flag"
	"fmt"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/resolver"
	"thermal/session"

	"gopkg.in/yaml.v3"
)

type ReferencesCommand struct{}

func New() *ReferencesCommand {
	return &ReferencesCommand{}
}

func parseArgs(args string) (string, string, bool, error) {
	fs := flag.NewFlagSet("references", flag.ContinueOnError)
	el := fs.String("e", "", "Pattern to match element names (* = any string)")
	tx := fs.String("t", "", "Pattern to match publisher names etc. (* = any string)")
	ls := fs.Bool("l", false, "List labels only")

	argv := strings.Fields(args)

	if err := fs.Parse(argv); err != nil {
		return "", "", false, err
	}

	if fs.NArg() > 0 {
		return "", "", false, fmt.Errorf("unknown parameter: %v", fs.Args())
	}

	return *el, *tx, *ls, nil
}

type OutputReference struct {
	ArcRole              string `yaml:"ArcRole"`
	Element              string `yaml:"Element"`
	ElementLocal         string `yaml:"-"` // YAMLでは出力しない
	Role                 string `yaml:"Role"`
	Publisher            string `yaml:"Publisher"`
	Number               string `yaml:"Number"`
	Name                 string `yaml:"Name"`
	IssueDate            string `yaml:"IssueDate"`
	Article              string `yaml:"Article"`
	IndustryAbbreviation string `yaml:"IndustryAbbreviation"`
}

func (c *ReferencesCommand) Execute(s *session.Session, args string) {

	elPattern, txPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	grouped, err := resolver.TraverseReferenceLink(s.Schema)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	var outputLabels []OutputReference

	for arcRole, relations := range grouped {
		for _, relation := range relations {
			elem, ok := relation.From.(*model.XMLElement)
			if !ok {
				panic("unreachable: REPL command dispatch should never reach this point")
			}

			reference, ok := relation.To.(*model.ReferenceReference)
			if !ok {
				panic("unreachable: REPL command dispatch should never reach this point")
			}

			if elPattern != "" && !parser.WildcardMatch(elPattern, elem.Name) {
				continue
			}
			if txPattern != "" &&
				!parser.WildcardMatch(txPattern, reference.Publisher) &&
				!parser.WildcardMatch(txPattern, reference.Number) &&
				!parser.WildcardMatch(txPattern, reference.Name) &&
				!parser.WildcardMatch(txPattern, reference.Article) &&
				!parser.WildcardMatch(txPattern, reference.IssueDate) &&
				!parser.WildcardMatch(txPattern, reference.IndustryAbbreviation) {
				continue
			}

			outputLabel := OutputReference{
				ArcRole:              arcRole,
				Role:                 reference.Role,
				Publisher:            reference.Publisher,
				Number:               reference.Number,
				Name:                 reference.Name,
				Article:              reference.Article,
				IssueDate:            reference.IssueDate,
				IndustryAbbreviation: reference.IndustryAbbreviation,
			}
			outputLabel.Element = fmt.Sprintf("{%s}%s", elem.Schema.TargetNS, elem.Name)
			outputLabel.ElementLocal = elem.Name
			outputLabels = append(outputLabels, outputLabel)
		}
	}
	if ls {
		for _, outputLabel := range outputLabels {
			fmt.Fprintln(s.Stdout, outputLabel.ElementLocal)
		}
	} else {
		encoder := yaml.NewEncoder(s.Stdout)
		encoder.SetIndent(2) // 読みやすさのためにインデント設定

		if err := encoder.Encode(outputLabels); err != nil {
			fmt.Fprintf(s.Stderr, "YAML encode error: %v\n", err)
		}
	}
}
