package presentations

import (
	"bytes"
	"flag"
	"fmt"
	"maps"
	"os"
	"sort"
	"strconv"
	"strings"
	"thermal/model"
	"thermal/parser"
	"thermal/resolver"
	"thermal/session"

	"github.com/ddddddO/gtree"
	"gopkg.in/yaml.v3"
)

type PresentationsCommand struct{}

func New() *PresentationsCommand {
	return &PresentationsCommand{}
}

func parseArgs(args string) (string, bool, error) {
	fs := flag.NewFlagSet("presentations", flag.ContinueOnError)
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

type OutputPlesentationLink struct {
	RoleType    string   `yaml:"RoleType"`
	ElementTree []string `yaml:"ElementTree"`
}

func (c *PresentationsCommand) Execute(s *session.Session, args string) {

	rtPattern, ls, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	grouped, err := resolver.TraversePresentationLink(s.Schema)
	if err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
		return
	}

	arcRoles := make([]string, 0, len(grouped))
	for k := range grouped {
		// ロールタイプのフィルタ指定があるときはマッチしたものだけ
		if rtPattern == "" || parser.WildcardMatch(rtPattern, k) {
			arcRoles = append(arcRoles, k)
		}
	}

	// ロールタイプのフィルタ指定があり、マッチしたのがなかったらエラー
	if rtPattern != "" && len(arcRoles) == 0 {
		fmt.Fprintf(s.Stdout, "roleType not found. %s \n", rtPattern)
		return
	}

	sort.Strings(arcRoles)

	if ls {
		for _, arcRole := range arcRoles {
			fmt.Fprintln(s.Stdout, arcRole)
		}
	} else {
		var outputElements []OutputPlesentationLink

		for _, arcRole := range arcRoles {
			roots := resolver.FindRootNodes(grouped[arcRole])
			adj := resolver.BuildAdjacency(grouped[arcRole])
			visited := map[*model.XMLElement]bool{}

			var trees []string
			for _, root := range roots {
				e := root.(*model.XMLElement)
				groot := gtree.NewRoot(fmt.Sprintf("Root,%s", e.Name))
				dfs(e, visited, adj, groot)

				var buf bytes.Buffer
				if err := gtree.OutputFromRoot(&buf, groot); err != nil {
					fmt.Fprintln(s.Stderr, "error:", err)
					return
				}
				trees = append(trees, buf.String())
			}
			outputElements = append(outputElements, OutputPlesentationLink{
				RoleType:    arcRole,
				ElementTree: trees,
			})
		}
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2) // 読みやすさのためにインデント設定

		if err := encoder.Encode(outputElements); err != nil {
			fmt.Fprintf(s.Stderr, "YAML encode error: %v\n", err)
		}
	}
}

func dfs(node *model.XMLElement, visited map[*model.XMLElement]bool, adj map[any][]*resolver.ArcRelation, gnode *gtree.Node) {
	if visited[node] {
		return
	}
	visited[node] = true

	relations := adj[node]
	sort.Slice(relations, func(i, j int) bool {
		arci := relations[i].Arc.(*model.PresentationArc)
		fi, err := strconv.ParseFloat(arci.Order, 64)
		if err != nil {
			fi = 1.0
		}

		arcj := relations[j].Arc.(*model.PresentationArc)
		fj, err := strconv.ParseFloat(arcj.Order, 64)
		if err != nil {
			fj = 1.0
		}

		return fi < fj
	})

	for _, child := range relations {
		to := child.To.(*model.XMLElement)
		arc := child.Arc.(*model.PresentationArc)
		text := fmt.Sprintf("%s,%s,%s", arc.Order, to.Name, arc.PreferredLabel)
		gnodec := gnode.Add(text)

		visitedCopy := make(map[*model.XMLElement]bool)
		maps.Copy(visitedCopy, visited)

		dfs(to, visitedCopy, adj, gnodec)
	}
}
