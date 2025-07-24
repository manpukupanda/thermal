package roletypes

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/resolver"
	"thermal/session"

	"gopkg.in/yaml.v3"
)

type RoleTypesCommand struct{}

func New() *RoleTypesCommand {
	return &RoleTypesCommand{}
}

func parseArgs(args string) (string, bool, error) {
	fs := flag.NewFlagSet("roletypes", flag.ContinueOnError)
	rt := fs.String("r", "", "Pattern to match role type URIs (* = any string)")
	ls := fs.Bool("l", false, "List role type URIs only")

	argv := strings.Fields(args)

	if err := fs.Parse(argv); err != nil {
		return "", false, err
	}

	if fs.NArg() > 0 {
		return "", false, fmt.Errorf("unknown parameter: %v", fs.Args())
	}

	return *rt, *ls, nil
}

func (c *RoleTypesCommand) Execute(s *session.Session, args string) {

	rtPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	allRoleTypes := make(map[string]*model.RoleType)
	resolver.CollectRoleTypesByHref(s.Schema, allRoleTypes)

	grouped, err := resolver.TraverseGenericLink(s.Schema, allRoleTypes)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	hrefs := make([]string, 0, len(allRoleTypes))
	for k := range allRoleTypes {
		// ロールタイプのフィルタ指定があるときはマッチしたものだけ
		if rtPattern == "" || parser.WildcardMatch(rtPattern, allRoleTypes[k].RoleURI) {
			hrefs = append(hrefs, k)
		}
	}

	// ロールタイプのフィルタ指定があり、マッチしたのがなかったらエラー
	if rtPattern != "" && len(hrefs) == 0 {
		fmt.Fprintf(s.Stdout, "roleType not found. %s \n", rtPattern)
		return
	}

	sort.Strings(hrefs)

	if ls {
		for _, href := range hrefs {
			fmt.Fprintln(s.Stdout, allRoleTypes[href].RoleURI)
		}
	} else {

		outputRoleTypes := make([]OutputRoleType, len(hrefs))
		for i, href := range hrefs {
			var rt OutputRoleType

			rt.RoleURI = allRoleTypes[href].RoleURI
			rt.Definition = allRoleTypes[href].Definition.Value

			for _, relations := range grouped {
				for _, rel := range relations {
					if rel.From == allRoleTypes[href] {
						lab := rel.To.(*model.GenericLabel)
						rt.GenLabel = append(rt.GenLabel, lab.Value)
					}
				}
			}
			for _, usedOn := range allRoleTypes[href].UsedOns {
				rt.UsedOn = append(rt.UsedOn, usedOn.Value)
			}
			rt.Href = href
			outputRoleTypes[i] = rt
		}

		encoder := yaml.NewEncoder(s.Stdout)
		encoder.SetIndent(2) // 読みやすさのためにインデント設定
		if err := encoder.Encode(outputRoleTypes); err != nil {
			fmt.Fprintf(s.Stderr, "YAML encode error: %v\n", err)
		}
	}
}

type OutputRoleType struct {
	RoleURI    string   `yaml:"RoleURI"`
	Definition string   `yaml:"Definition"`
	GenLabel   []string `yaml:"GenLabel"`
	UsedOn     []string `yaml:"UsedOn"`
	Href       string   `yaml:"Href"`
}
