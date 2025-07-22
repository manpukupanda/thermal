package elements

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/session"

	"gopkg.in/yaml.v3"
)

type ElementsCommand struct{}

func New() *ElementsCommand {
	return &ElementsCommand{}
}

func parseArgs(args string) (string, bool, error) {
	fs := flag.NewFlagSet("elements", flag.ContinueOnError)
	el := fs.String("e", "", "Pattern to match element names (* = any string)")
	ls := fs.Bool("l", false, "List element names only")

	argv := strings.Fields(args)

	if err := fs.Parse(argv); err != nil {
		return "", false, err
	}

	if fs.NArg() > 0 {
		return "", false, fmt.Errorf("unknown parameter: %v", fs.Args())
	}

	return *el, *ls, nil
}

type OutputElement struct {
	Name       string `yaml:"Name"`
	Namespace  string `yaml:"Namespace"`
	Type       string `yaml:"Type"`
	PeriodType string `yaml:"PeriodType"`
	Abstract   string `yaml:"Abstract"`
	Nillable   string `yaml:"Nillable"`
	Href       string `yaml:"Href"`
}

func (c *ElementsCommand) Execute(s *session.Session, args string) {

	elPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	elements, err := schemaTree(s.Schema)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	sort.Slice(elements, func(i, j int) bool {
		return strings.ToLower(elements[i].Name) < strings.ToLower(elements[j].Name)
	})

	var outputElements []OutputElement

	for _, element := range elements {
		if elPattern != "" && !parser.WildcardMatch(elPattern, element.Name) {
			continue
		}
		outputElement := OutputElement{
			Name:       element.Name,
			Namespace:  element.Schema.TargetNS,
			Type:       element.Type,
			PeriodType: element.PeriodType,
			Abstract:   element.Abstract,
			Nillable:   element.Nillable,
		}
		outputElement.Href = fmt.Sprintf("%s#%s", element.Schema.Path, element.Id)
		outputElements = append(outputElements, outputElement)
	}
	if ls {
		for _, outputElement := range outputElements {
			fmt.Println(outputElement.Name)
		}
	} else {
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2) // 読みやすさのためにインデント設定

		if err := encoder.Encode(outputElements); err != nil {
			log.Fatalf("YAML encode error: %v", err)
		}
	}
}

func schemaTree(schema *model.XBRLSchema) ([]*model.XMLElement, error) {
	visited := make(map[string]bool)
	var elemens []*model.XMLElement
	traverse(schema, &elemens, visited)
	return elemens, nil
}

func traverse(schema *model.XBRLSchema, elements *[]*model.XMLElement, visited map[string]bool) error {
	if schema == nil {
		return nil
	}

	// 同じスキーマファイルを2度処理しない
	if visited[schema.Path] {
		return nil
	}
	visited[schema.Path] = true

	for i := range schema.Elements {
		*elements = append(*elements, &schema.Elements[i])
	}

	//	*elements = append(*elements, schema.Elements...)

	for i := range schema.Imports {
		s := schema.Imports[i].Schema
		if s != nil {
			if err := traverse(s, elements, visited); err != nil {
				return err
			}
		}
	}
	return nil
}
