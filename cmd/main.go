package main

import (
	"fmt"
	"os"
	"thermal/parser"
	"thermal/repl"
	"thermal/replcmd/registry"
	"thermal/session"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: thermal <manifest.xml>|<schema.xsd>|<instance.xbrl>")
		os.Exit(1)
	}

	entryFile := os.Args[1]

	rootName, err := parser.PeekXMLRootElementName(entryFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load entry file: %v\n", err)
		os.Exit(1)
	}

	var session session.Session

	switch rootName {
	case "manifest":
		manifest, err := parser.ParseManifest(entryFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load manifest: %v\n", err)
			os.Exit(1)
		}

		session.Manifest = manifest
		session.Instance = manifest.List.XBRLInstances[0]
		session.Schema = manifest.List.XBRLInstances[0].SchemaRefs.Schema
	case "xbrl":
		instance, err := parser.ParseInstance(entryFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load XBRL: %v\n", err)
			os.Exit(1)
		}
		session.Instance = instance
		session.Schema = instance.SchemaRefs.Schema
	case "schema":
		visited := make(map[string]bool)
		schema, err := parser.ParseSchema(entryFile, visited)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load schema: %v\n", err)
			os.Exit(1)
		}
		session.Schema = schema
	default:
		fmt.Fprintf(os.Stderr, "file unknown\n")
		os.Exit(1)
	}

	registry.RegisterAll()
	repl.Start(&session)

	/*

		// コマンドライン引数でエントリーファイルを指定できるようにする
		//entryFile := flag.String("entry", "./data/manifest_PublicDoc.xml", "エントリーポイントとなるXMLファイル")
		//flag.Parse()

		//manifest, err := parser.ParseManifest(*entryFile)
		xbrlInstance := session.Manifest.List.XBRLInstances[0]

		// 全要素をマップに保持（リンクベースのロケータからの参照用）
		allElements := make(map[string]*model.XMLElement)
		resolver.CollectElementsByHref(xbrlInstance.SchemaRefs.Schema, allElements)

		allRoleTypes := make(map[string]model.RoleType)
			resolver.CollectRoleTypesByHref(xbrlInstance.SchemaRefs.Schema, allRoleTypes)

			fmt.Println("\n DTS出力:")
			csv, err := exporter.CsvDts(xbrlInstance, true)
			if err != nil {
				fmt.Println("❌ DTS失敗:", err)
				return
			}
			fmt.Print(csv)

			// ラベルを表示
			fmt.Println("\n 全ラベル出力:")
			csv, err = exporter.CsvLabels(xbrlInstance.SchemaRefs.Schema, allElements, true)
			if err != nil {
				fmt.Println("❌ ラベル失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n 表示リンク出力:")
			csv, err = exporter.CsvPresentationLinks(xbrlInstance.SchemaRefs.Schema, allElements, true)
			if err != nil {
				fmt.Println("❌ 表示リンク失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n 要素リスト出力:")
			csv, err = exporter.CsvElements(xbrlInstance.SchemaRefs.Schema, true)
			if err != nil {
				fmt.Println("❌ 要素リスト失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n ロールタイプリスト出力:")
			csv, err = exporter.CsvRoleTypes(xbrlInstance.SchemaRefs.Schema, true)
			if err != nil {
				fmt.Println("❌ ロールタイプリスト失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n ジェネリックリンクリスト出力:")
			csv, err = exporter.CsvGenericLinks(xbrlInstance.SchemaRefs.Schema, allRoleTypes, true)
			if err != nil {
				fmt.Println("❌ ジェネリックリンクリスト失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n ファクトリスト出力:")
			csv, err = exporter.CsvFacts(xbrlInstance, true)
			if err != nil {
				fmt.Println("❌ ファクトリスト失敗:", err)
				return
			}
			fmt.Print(csv)

			fmt.Println("\n コンテキストリスト出力:")
			csv, err = exporter.CsvContexts(xbrlInstance, true)
			if err != nil {
				fmt.Println("❌ コンテキストリスト失敗:", err)
				return
			}
			fmt.Print(csv)
	*/
}
