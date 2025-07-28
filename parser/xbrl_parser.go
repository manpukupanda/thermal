package parser

import (
	"fmt"
	"maps"
	"strings"
	"sync"
	"thermal/model"
)

// 🚀 XBRLの標準スキーマかどうかを判定する関数
func IsStandardXBRLSchema(href string) bool {
	return strings.HasPrefix(href, "http://www.xbrl.org/")
}

// スキーマのキャッシュ（スレッドセーフにするために `sync.Map` を使用）
var schemaCache sync.Map

// スキーマを解析する
func ParseSchema(filename string, visited map[string]bool) (*model.XBRLSchema, error) {

	// 🔍 すでに解析済みならキャッシュを返す
	if cached, exists := schemaCache.Load(filename); exists {
		return cached.(*model.XBRLSchema), nil
	}

	// 🔍 循環 `import` の検出
	if visited[filename] {
		return nil, fmt.Errorf("🚨 循環スキーマインポート検出: %s", filename)
	}

	// 🔥 ここで `visited` をコピーし、スレッド間で共有しないようにする！
	visitedCopy := make(map[string]bool)
	maps.Copy(visitedCopy, visited)
	visitedCopy[filename] = true // 記録

	// スキーマのパース
	schema, err := ParseXML[model.XBRLSchema](filename)
	if err != nil {
		return nil, fmt.Errorf("❌ スキーマのXMLパースエラー: %s", err)
	}
	// 🔥 ファイル名を保存して、スキーマの出所を明確化
	schema.Path = filename
	for i := range schema.Elements {
		schema.Elements[i].Schema = schema
	}
	for i := range schema.RoleTypes {
		schema.RoleTypes[i].Schema = schema
	}

	// 🔥 `linkbaseRef` を確認し、対応するリンクベースの解析処理を呼び出す
	var wg1 sync.WaitGroup
	var mu1 sync.Mutex

	for _, linkbaseRef := range schema.LinkbaseRefs {
		wg1.Add(1)

		go func(linkbaseRef model.LinkbaseRef) {
			defer wg1.Done()
			href := ResolveHref(filename, linkbaseRef.Href)

			if strings.Contains(linkbaseRef.Role, "labelLinkbaseRef") {
				linkbase, err := ParseXML[model.LabelLinkBase](href)
				if err != nil {
					fmt.Printf("❌ 名称リンクベースパースエラー: %s", err)
					return
				}
				linkbase.Path = href
				for i := range linkbase.LabelLinks {
					for j := range linkbase.LabelLinks[i].Labels {
						linkbase.LabelLinks[i].Labels[j].LinkBase = linkbase
					}
				}
				mu1.Lock()
				schema.ReferencedLabelLinkbases = append(schema.ReferencedLabelLinkbases, linkbase)
				mu1.Unlock()
			} else if strings.Contains(linkbaseRef.Role, "referenceLinkbaseRef") {
				linkbase, err := ParseXML[model.ReferenceLinkBase](href)
				if err != nil {
					fmt.Printf("❌ 参照リンクベースパースエラー: %s", err)
					return
				}
				linkbase.Path = href
				for i := range linkbase.ReferenceLinks {
					for j := range linkbase.ReferenceLinks[i].References {
						linkbase.ReferenceLinks[i].References[j].LinkBase = linkbase
					}
				}
				mu1.Lock()
				schema.ReferencedReferenceLinkbases = append(schema.ReferencedReferenceLinkbases, linkbase)
				mu1.Unlock()
			} else if strings.Contains(linkbaseRef.Role, "presentationLinkbaseRef") {
				linkbase, err := ParseXML[model.PresentationLinkBase](href)
				if err != nil {
					fmt.Printf("❌ 表示リンクベースパースエラー: %s", err)
					return
				}
				linkbase.Path = href
				mu1.Lock()
				schema.ReferencedPresentationLinkbases = append(schema.ReferencedPresentationLinkbases, linkbase)
				mu1.Unlock()
			} else if strings.Contains(linkbaseRef.Role, "definitionLinkbaseRef") {
				linkbase, err := ParseXML[model.DefinitionLinkBase](href)
				if err != nil {
					fmt.Printf("❌ 定義リンクベースパースエラー: %s", err)
					return
				}
				linkbase.Path = href
				mu1.Lock()
				schema.ReferencedDefinitionLinkbases = append(schema.ReferencedDefinitionLinkbases, linkbase)
				mu1.Unlock()
			} else if strings.Contains(linkbaseRef.Role, "calculationLinkbaseRef") {
				linkbase, err := ParseXML[model.CalculationLinkBase](href)
				if err != nil {
					fmt.Printf("❌ 計算リンクベースパースエラー: %s", err)
					return
				}
				linkbase.Path = href
				mu1.Lock()
				schema.ReferencedCalculationLinkbases = append(schema.ReferencedCalculationLinkbases, linkbase)
				mu1.Unlock()
			} else if linkbaseRef.Role == "" {
				linkbase, err := ParseXML[model.GenericLinkBase](href)
				if err != nil {
					fmt.Printf("❌ ジェネリックリンクベースパース: %s", err)
					return
				}
				linkbase.Path = href
				mu1.Lock()
				schema.ReferencedGenericLinkbases = append(schema.ReferencedGenericLinkbases, linkbase)
				mu1.Unlock()
			}
		}(linkbaseRef)
	}

	// 🔥 全ての `linkbaseRef` の処理が終わるまで待機
	wg1.Wait()

	// キャッシュへ保存
	schemaCache.Store(filename, schema)

	// 🔥 インポートされたスキーマを並列処理
	var wg2 sync.WaitGroup

	for i := range schema.Imports {
		wg2.Add(1)
		go func(imp *model.XMLImport) {
			defer wg2.Done()

			importPath := ResolveHref(filename, imp.SchemaLoc)

			// 🔥 XBRL標準スキーマならスキップ
			if IsStandardXBRLSchema(importPath) {
				return
			}

			// 🚀 スレッドローカルな `visitedCopy` を使って循環をチェック
			importedSchema, err := ParseSchema(importPath, visitedCopy)
			if err != nil {
				fmt.Printf("⚠️ インポートスキーマのパースに失敗: %s\n", err)
				return
			}

			imp.Schema = importedSchema
		}(&schema.Imports[i])
	}

	// 🔥 全ての `import` の処理が終わるまで待機
	wg2.Wait()

	return schema, nil
}

// インスタンスを解析する
func ParseInstance(instanceFile string) (*model.XBRLInstance, error) {
	// インスタンスの解析
	xbrlInstance, err := ParseXML[model.XBRLInstance](instanceFile)
	if err != nil {
		return nil, fmt.Errorf("❌ XBRLインスタンスのパースに失敗:%v", err)
	}
	xbrlInstance.Path = instanceFile

	// スキーマファイルの取得
	if xbrlInstance.SchemaRefs.Href == "" {
		return nil, fmt.Errorf("❌ スキーマファイルが見つかりません")
	}
	schemaFilename := xbrlInstance.SchemaRefs.Href
	schemaFile := ResolveHref(instanceFile, schemaFilename)

	// スキーマ解析のための訪問履歴管理
	visited := make(map[string]bool)

	// スキーマの解析
	schema, err := ParseSchema(schemaFile, visited)
	if err != nil {
		return nil, fmt.Errorf("❌ スキーマのパースに失敗:%v", err)
	}
	xbrlInstance.SchemaRefs.Schema = schema
	return xbrlInstance, nil
}
