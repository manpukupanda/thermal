package labels

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/resolver"
	"thermal/session"

	"gopkg.in/yaml.v3"
)

type LabelsCommand struct{}

func New() *LabelsCommand {
	return &LabelsCommand{}
}

func parseArgs(args string) (string, string, bool, error) {
	fs := flag.NewFlagSet("elements", flag.ContinueOnError)
	el := fs.String("e", "", "Pattern to match element names (* = any string)")
	tx := fs.String("t", "", "Pattern to match labels (* = any string)")
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

type OutputLabel struct {
	ArcRole string `yaml:"ArcRole"`
	Element string `yaml:"Element"`
	Lang    string `yaml:"Lang"`
	Role    string `yaml:"Role"`
	Value   string `yaml:"Label"`
	Href    string `yaml:"Href"`
}

func (c *LabelsCommand) Execute(s *session.Session, args string) {

	elPattern, txPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	grouped, err := resolver.TraverseLabelLink(s.Schema)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var outputLabels []OutputLabel

	for arcRole, relations := range grouped {
		for _, relation := range relations {
			elem, ok := relation.From.(*model.XMLElement)
			if !ok {
				panic("unreachable: REPL command dispatch should never reach this point")
			}

			label, ok := relation.To.(*model.LabelLabel)
			if !ok {
				panic("unreachable: REPL command dispatch should never reach this point")
			}

			if elPattern != "" && !parser.WildcardMatch(elPattern, elem.Name) {
				continue
			}
			if txPattern != "" && !parser.WildcardMatch(txPattern, label.Value) {
				continue
			}

			outputLabel := OutputLabel{
				ArcRole: arcRole,
				Lang:    label.Lang,
				Role:    label.Role,
				Value:   label.Value,
			}
			outputLabel.Element = fmt.Sprintf("{%s}%s", elem.Schema.TargetNS, elem.Name)
			if label.Id != "" {
				outputLabel.Href = fmt.Sprintf("%s#%s", label.LinkBase.Path, label.Id)
			}
			outputLabels = append(outputLabels, outputLabel)
		}
	}
	if ls {
		for _, outputLabel := range outputLabels {
			fmt.Println(outputLabel.Value)
		}
	} else {
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2) // 読みやすさのためにインデント設定

		if err := encoder.Encode(outputLabels); err != nil {
			log.Fatalf("YAML encode error: %v", err)
		}
	}
}
