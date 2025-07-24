package dts

import (
	"fmt"
	"thermal/model"
	"thermal/session"

	"github.com/ddddddO/gtree"
)

type DtsCommand struct{}

func New() *DtsCommand {
	return &DtsCommand{}
}

func (c *DtsCommand) Execute(s *session.Session, args string) {
	var root *gtree.Node
	if s.Manifest != nil {
		root = manifestTree(s.Manifest)
	} else if s.Instance != nil {
		root = instanceTree(s.Instance)
	} else if s.Schema != nil {
		root = schemaTree(s.Schema)
	}
	if err := gtree.OutputFromRoot(s.Stdout, root); err != nil {
		fmt.Fprintln(s.Stderr, "error:", err)
	}
}

func manifestTree(manifest *model.Manifest) *gtree.Node {
	root := gtree.NewRoot(manifest.Path)
	for i := range manifest.List.XBRLInstances {
		if manifest.List.XBRLInstances[i] != nil {
			c := root.Add(manifest.List.XBRLInstances[i].Path)
			if manifest.List.XBRLInstances[i].SchemaRefs.Schema != nil {
				c2 := c.Add(manifest.List.XBRLInstances[i].SchemaRefs.Schema.Path)
				traverse(manifest.List.XBRLInstances[i].SchemaRefs.Schema, c2)
			}
		}
	}
	return root
}

func instanceTree(instance *model.XBRLInstance) *gtree.Node {
	root := gtree.NewRoot(instance.Path)
	if instance.SchemaRefs.Schema != nil {
		c := root.Add(instance.SchemaRefs.Schema.Path)
		traverse(instance.SchemaRefs.Schema, c)
	}
	return root
}

func schemaTree(schema *model.XBRLSchema) *gtree.Node {
	root := gtree.NewRoot(schema.Path)
	traverse(schema, root)
	return root
}

func traverse(schema *model.XBRLSchema, node *gtree.Node) {
	if schema == nil {
		return
	}

	for i := range schema.Imports {
		s := schema.Imports[i].Schema
		if s != nil {
			t := fmt.Sprintf("(S)%s", s.Path)
			c := node.Add(t)
			traverse(s, c)
		}
	}

	for i := range schema.ReferencedPresentationLinkbases {
		s := schema.ReferencedPresentationLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(P)%s", s.Path)
			node.Add(t)
		}
	}
	for i := range schema.ReferencedCalculationLinkbases {
		s := schema.ReferencedCalculationLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(C)%s", s.Path)
			node.Add(t)
		}
	}
	for i := range schema.ReferencedDefinitionLinkbases {
		s := schema.ReferencedDefinitionLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(D)%s", s.Path)
			node.Add(t)
		}
	}
	for i := range schema.ReferencedLabelLinkbases {
		s := schema.ReferencedLabelLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(L)%s", s.Path)
			node.Add(t)
		}
	}
	for i := range schema.ReferencedReferenceLinkbases {
		s := schema.ReferencedReferenceLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(R)%s", s.Path)
			node.Add(t)
		}
	}
	for i := range schema.ReferencedGenericLinkbases {
		s := schema.ReferencedGenericLinkbases[i]
		if s != nil {
			t := fmt.Sprintf("(gla)%s", s.Path)
			node.Add(t)
		}
	}
}
