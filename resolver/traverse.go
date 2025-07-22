package resolver

import (
	"fmt"
	"thermal/model"
	"thermal/parser"
)

type ArcRelation struct {
	ArcRole string
	Arc     any
	From    any
	To      any
}

func TraverseLabelLink(schema *model.XBRLSchema) (map[string][]ArcRelation, error) {
	return traverseLink(schema, dfsLabelLink)
}

func TraverseReferenceLink(schema *model.XBRLSchema) (map[string][]ArcRelation, error) {
	return traverseLink(schema, dfsReferenceLink)
}

func TraversePresentationLink(schema *model.XBRLSchema) (map[string][]ArcRelation, error) {
	return traverseLink(schema, dfsPresentationLink)
}

func TraverseDefinitionLink(schema *model.XBRLSchema) (map[string][]ArcRelation, error) {
	return traverseLink(schema, dfsDefinitionLink)
}

func traverseLink(
	schema *model.XBRLSchema,
	traverseFn func(*model.XBRLSchema, map[string]*model.XMLElement, map[string]bool, []ArcRelation) ([]ArcRelation, error),
) (map[string][]ArcRelation, error) {

	allElements := make(map[string]*model.XMLElement)
	CollectElementsByHref(schema, allElements)

	var relations []ArcRelation
	visited := make(map[string]bool)

	relations, err := traverseFn(schema, allElements, visited, relations)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]ArcRelation)
	for _, r := range relations {
		grouped[r.ArcRole] = append(grouped[r.ArcRole], r)
	}

	return grouped, nil
}

func dfsLink(
	schema *model.XBRLSchema,
	elements map[string]*model.XMLElement,
	visited map[string]bool,
	getLinkbasePathsFn func(*model.XBRLSchema) []string,
	collectRetationsFn func(*model.XBRLSchema, string) ([]ArcRelation, error),
	relations []ArcRelation) ([]ArcRelation, error) {

	if schema == nil || visited[schema.Path] {
		return relations, nil
	}
	visited[schema.Path] = true

	for _, linkbasePath := range getLinkbasePathsFn(schema) {
		if visited[linkbasePath] {
			continue
		}
		visited[linkbasePath] = true

		relationsCurr, err := collectRetationsFn(schema, linkbasePath)
		if err != nil {
			return nil, err
		}
		relations = append(relations, relationsCurr...)
	}

	for i := range schema.Imports {
		if schema.Imports[i].Schema != nil {
			x, err := dfsLink(schema.Imports[i].Schema, elements, visited, getLinkbasePathsFn, collectRetationsFn, relations)
			if err != nil {
				return nil, err
			}
			relations = x
		}
	}
	return relations, nil
}

// ロケータのスライスから、ロケータのラベルをキーにしたmapを作成する
func makeLocsMap(locs *[]model.Loc) map[string]*model.Loc {
	locMap := make(map[string]*model.Loc, len(*locs))
	for _, loc := range *locs {
		locMap[loc.Label] = &loc
	}
	return locMap
}

func dfsLabelLink(schema *model.XBRLSchema, elements map[string]*model.XMLElement, visited map[string]bool, relations []ArcRelation) ([]ArcRelation, error) {

	return dfsLink(
		schema, elements, visited,
		func(s *model.XBRLSchema) []string {
			linkbases := make([]string, len(s.ReferencedLabelLinkbases))
			for i, v := range s.ReferencedLabelLinkbases {
				linkbases[i] = v.Path
			}
			return linkbases
		},
		func(s *model.XBRLSchema, path string) ([]ArcRelation, error) {
			var llb *model.LabelLinkBase
			for i := range s.ReferencedLabelLinkbases {
				if s.ReferencedLabelLinkbases[i].Path == path {
					llb = s.ReferencedLabelLinkbases[i]
					break
				}
			}
			if llb == nil {
				return nil, fmt.Errorf("Linkbase not found.")
			}

			var rels []ArcRelation
			for _, elr := range llb.LabelLinks {
				locMap := makeLocsMap(&elr.Locs)
				labelMap := make(map[string]*model.LabelLabel, len(elr.Labels))
				for _, label := range elr.Labels {
					labelMap[label.Label] = &label
				}
				for i, arc := range elr.Arcs {
					loc, ok := locMap[arc.From]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: from=%s", arc.From)
					}
					label, ok := labelMap[arc.To]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: to=%s", arc.To)
					}
					key := parser.ResolveHref(path, loc.Href)
					elem, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}
					var r ArcRelation
					r.ArcRole = elr.Role
					r.Arc = &elr.Arcs[i]
					r.From = elem
					r.To = label
					rels = append(rels, r)
				}
			}
			return rels, nil
		}, relations)
}

func dfsReferenceLink(schema *model.XBRLSchema, elements map[string]*model.XMLElement, visited map[string]bool, relations []ArcRelation) ([]ArcRelation, error) {

	return dfsLink(
		schema, elements, visited,
		func(s *model.XBRLSchema) []string {
			linkbases := make([]string, len(s.ReferencedReferenceLinkbases))
			for i, v := range s.ReferencedReferenceLinkbases {
				linkbases[i] = v.Path
			}
			return linkbases
		},
		func(s *model.XBRLSchema, path string) ([]ArcRelation, error) {
			var rels []ArcRelation
			var rlb *model.ReferenceLinkBase
			for i := range s.ReferencedReferenceLinkbases {
				if s.ReferencedReferenceLinkbases[i].Path == path {
					rlb = s.ReferencedReferenceLinkbases[i]
					break
				}
			}
			if rlb == nil {
				return nil, fmt.Errorf("Linkbase not found.")
			}

			for _, elr := range rlb.ReferenceLinks {
				locMap := makeLocsMap(&elr.Locs)
				rerefenceMap := make(map[string]*model.ReferenceReference, len(elr.References))
				for _, reference := range elr.References {
					rerefenceMap[reference.Label] = &reference
				}
				for i, arc := range elr.Arcs {
					loc, ok := locMap[arc.From]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: from=%s", arc.From)
					}
					ref, ok := rerefenceMap[arc.To]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: to=%s", arc.To)
					}
					key := parser.ResolveHref(path, loc.Href)
					elem, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}
					var r ArcRelation
					r.ArcRole = elr.Role
					r.Arc = &elr.Arcs[i]
					r.From = elem
					r.To = ref
					rels = append(rels, r)
				}
			}
			return rels, nil
		}, relations)
}

func dfsPresentationLink(schema *model.XBRLSchema, elements map[string]*model.XMLElement, visited map[string]bool, relations []ArcRelation) ([]ArcRelation, error) {
	return dfsLink(
		schema, elements, visited,
		func(s *model.XBRLSchema) []string {
			linkbases := make([]string, len(s.ReferencedPresentationLinkbases))
			for i, v := range s.ReferencedPresentationLinkbases {
				linkbases[i] = v.Path
			}
			return linkbases
		},
		func(s *model.XBRLSchema, path string) ([]ArcRelation, error) {
			var rels []ArcRelation
			var plb *model.PresentationLinkBase
			for i := range s.ReferencedPresentationLinkbases {
				if s.ReferencedPresentationLinkbases[i].Path == path {
					plb = s.ReferencedPresentationLinkbases[i]
					break
				}
			}
			if plb == nil {
				return nil, fmt.Errorf("Linkbase not found.")
			}

			for _, elr := range plb.PresentationLinks {
				locMap := makeLocsMap(&elr.Locs)

				for i, arc := range elr.Arcs {
					locFrom, ok := locMap[arc.From]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: from=%s", arc.From)
					}
					key := parser.ResolveHref(plb.Path, locFrom.Href)
					elemFrom, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}

					locTo, ok := locMap[arc.To]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: to=%s", arc.To)
					}
					key = parser.ResolveHref(path, locTo.Href)
					elemTo, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}

					var r ArcRelation
					r.ArcRole = elr.Role
					r.Arc = &elr.Arcs[i]
					r.From = elemFrom
					r.To = elemTo
					rels = append(rels, r)
				}
			}
			return rels, nil
		}, relations)
}

func dfsDefinitionLink(schema *model.XBRLSchema, elements map[string]*model.XMLElement, visited map[string]bool, relations []ArcRelation) ([]ArcRelation, error) {
	return dfsLink(
		schema, elements, visited,
		func(s *model.XBRLSchema) []string {
			linkbases := make([]string, len(s.ReferencedDefinitionLinkbases))
			for i, v := range s.ReferencedDefinitionLinkbases {
				linkbases[i] = v.Path
			}
			return linkbases
		},
		func(s *model.XBRLSchema, path string) ([]ArcRelation, error) {
			var rels []ArcRelation
			var dlb *model.DefinitionLinkBase
			for i := range s.ReferencedDefinitionLinkbases {
				if s.ReferencedDefinitionLinkbases[i].LinkBase.Path == path {
					dlb = s.ReferencedDefinitionLinkbases[i]
					break
				}
			}
			if dlb == nil {
				return nil, fmt.Errorf("Linkbase not found.")
			}

			for _, elr := range dlb.DefinitionLinks {
				locMap := makeLocsMap(&elr.Locs)

				for i, arc := range elr.Arcs {
					locFrom, ok := locMap[arc.From]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: from=%s", arc.From)
					}
					key := parser.ResolveHref(dlb.Path, locFrom.Href)
					elemFrom, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}

					locTo, ok := locMap[arc.To]
					if !ok {
						return nil, fmt.Errorf("Arc invalid: to=%s", arc.To)
					}
					key = parser.ResolveHref(path, locTo.Href)
					elemTo, ok := elements[key]
					if !ok {
						return nil, fmt.Errorf("Loc invalid: %s", key)
					}

					var r ArcRelation
					r.ArcRole = elr.Role
					r.Arc = &elr.Arcs[i]
					r.From = elemFrom
					r.To = elemTo
					rels = append(rels, r)
				}
			}
			return rels, nil
		}, relations)
}

func TraverseGenericLink(schema *model.XBRLSchema, roleTypes map[string]*model.RoleType) (map[string][]ArcRelation, error) {

	var relations []ArcRelation
	visited := make(map[string]bool)

	relations, err := dfsGenericLink(schema, roleTypes, visited, relations)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]ArcRelation)
	for _, r := range relations {
		grouped[r.ArcRole] = append(grouped[r.ArcRole], r)
	}

	return grouped, nil
}

func dfsGenericLink(schema *model.XBRLSchema, roleTypes map[string]*model.RoleType, visited map[string]bool, relations []ArcRelation) ([]ArcRelation, error) {

	if schema == nil || visited[schema.Path] {
		return relations, nil
	}
	visited[schema.Path] = true

	for _, linkbase := range schema.ReferencedGenericLinkbases {
		if visited[linkbase.Path] {
			// 同じリンクベースは2回処理しない
			continue
		}
		visited[linkbase.Path] = true

		for _, elr := range linkbase.GenericLinks {
			locMap := makeLocsMap(&elr.Locs)
			labelMap := make(map[string]*model.GenericLabel, len(elr.Labels))
			for _, label := range elr.Labels {
				labelMap[label.Label] = &label
			}
			for i, arc := range elr.Arcs {
				loc, ok := locMap[arc.From]
				if !ok {
					return nil, fmt.Errorf("Arc invalid: from=%s", arc.From)
				}
				label, ok := labelMap[arc.To]
				if !ok {
					return nil, fmt.Errorf("Arc invalid: to=%s", arc.To)
				}
				key := parser.ResolveHref(linkbase.Path, loc.Href)
				roleType, ok := roleTypes[key]
				if !ok {
					return nil, fmt.Errorf("Loc invalid: %s", key)
				}
				var r ArcRelation
				r.ArcRole = elr.Role
				r.Arc = &elr.Arcs[i]
				r.From = roleType
				r.To = label
				relations = append(relations, r)
			}
		}
	}

	for i := range schema.Imports {
		if schema.Imports[i].Schema != nil {
			x, err := dfsGenericLink(schema.Imports[i].Schema, roleTypes, visited, relations)
			if err != nil {
				return nil, err
			}
			relations = x
		}
	}
	return relations, nil
}

func FindRootNodes(relations []ArcRelation) []any {
	fromSet := map[any]struct{}{}
	toSet := map[any]struct{}{}

	for _, rel := range relations {
		fromSet[rel.From] = struct{}{}
		if to, ok := rel.To.(*model.XMLElement); ok {
			toSet[to] = struct{}{}
		}
	}

	roots := []any{}
	for from := range fromSet {
		if _, exists := toSet[from]; !exists {
			roots = append(roots, from)
		}
	}

	return roots
}

func BuildAdjacency(relations []ArcRelation) map[any][]*ArcRelation {
	adj := make(map[any][]*ArcRelation)
	for i, rel := range relations {
		adj[rel.From] = append(adj[rel.From], &relations[i])
	}
	return adj
}
