package main

import (
	"fmt"
	"os"
	"thermal/parser"
	"thermal/repl"
	"thermal/replcmd/registry"
	"thermal/session"

	"golang.org/x/term"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: thermal <manifest.xml>|<schema.xsd>|<instance.xbrl>")
		os.Exit(1)
	}

	var session session.Session
	// 対話モードか判定し、出力先を設定
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	if isTerminal {
		session.Stderr = os.Stdout
	} else {
		session.Stderr = os.Stderr
	}

	entryFile := os.Args[1]

	rootName, err := parser.PeekXMLRootElementName(entryFile)
	if err != nil {
		fmt.Fprintf(session.Stderr, "failed to load entry file: %v\n", err)
		os.Exit(1)
	}

	switch rootName {
	case "manifest":
		manifest, err := parser.ParseManifest(entryFile)
		if err != nil {
			fmt.Fprintf(session.Stderr, "failed to load manifest: %v\n", err)
			os.Exit(1)
		}

		session.Manifest = manifest
		session.Instance = manifest.List.XBRLInstances[0]
		session.Schema = manifest.List.XBRLInstances[0].SchemaRefs.Schema
	case "xbrl":
		instance, err := parser.ParseInstance(entryFile)
		if err != nil {
			fmt.Fprintf(session.Stderr, "failed to load XBRL: %v\n", err)
			os.Exit(1)
		}
		session.Instance = instance
		session.Schema = instance.SchemaRefs.Schema
	case "schema":
		visited := make(map[string]bool)
		schema, err := parser.ParseSchema(entryFile, visited)
		if err != nil {
			fmt.Fprintf(session.Stderr, "failed to load schema: %v\n", err)
			os.Exit(1)
		}
		session.Schema = schema
	default:
		fmt.Fprintf(session.Stderr, "file unknown\n")
		os.Exit(1)
	}

	registry.RegisterAll()
	repl.Start(&session)
}
