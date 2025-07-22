package exporter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"thermal/model"
	"thermal/parser"
)

// DTSのcsv形式文字列作成
func CsvDts(instance *model.XBRLInstance, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{"RefFrom", "RefType", "RefTo", "Stop"})
	}

	writer.Write([]string{instance.Path, "schemaRef", instance.SchemaRefs.Schema.Path})

	// DTSツリーを展開
	if err := writeDts(instance.SchemaRefs.Schema, writer); err != nil {
		return "", err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeDts(schema *model.XBRLSchema, writer *csv.Writer) error {
	for _, linkbase := range schema.ReferencedPresentationLinkbases {
		writer.Write([]string{schema.Path, "presentationLinkbaseRef", linkbase.Path, ""})
	}
	for _, linkbase := range schema.ReferencedDefinitionLinkbases {
		writer.Write([]string{schema.Path, "definitionLinkbaseRef", linkbase.Path, ""})
	}
	for _, linkbase := range schema.ReferencedCalculationLinkbases {
		writer.Write([]string{schema.Path, "calculationLinkbaseRef", linkbase.Path, ""})
	}
	for _, linkbase := range schema.ReferencedLabelLinkbases {
		writer.Write([]string{schema.Path, "labelLinkbaseRef", linkbase.Path, ""})
	}
	for _, linkbase := range schema.ReferencedGenericLinkbases {
		writer.Write([]string{schema.Path, "linkbaseRef", linkbase.Path, ""})
	}

	for _, child := range schema.Imports {
		if child.Schema != nil {
			writer.Write([]string{schema.Path, "import", child.Schema.Path, ""})
			if err := writeDts(child.Schema, writer); err != nil {
				return err
			}
		} else {
			writer.Write([]string{schema.Path, "import", child.SchemaLoc, "Y"})
		}
	}
	return nil
}

// 全ロールタイプcsv形式文字列作成
func CsvRoleTypes(schema *model.XBRLSchema, withheader bool) (string, error) {
	var sb strings.Builder
	if withheader {
		sb.WriteString("Path,Id,TargetNamespace,RoleURI,Definition\n")
	}

	for _, roleType := range schema.RoleTypes {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s\n", schema.Path, roleType.Id, schema.TargetNS, roleType.RoleURI, roleType.Definition.Value))
	}

	for _, child := range schema.Imports {
		if child.Schema != nil {
			csv, err := CsvRoleTypes(child.Schema, false)
			if err != nil {
				return "", err
			}
			sb.WriteString(csv)
		}
	}
	result := sb.String()
	return result, nil
}

// 全ジェネリックリンクcsv形式文字列作成
func CsvGenericLinks(schema *model.XBRLSchema, roleTypes map[string]model.RoleType, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"Path", "Id", "TargetNamespace", "RoleURI", "Definition", "Generic Label",
		})
	}

	if err := writeGenericLinks(schema, roleTypes, writer); err != nil {
		return "", err
	}

	// 子スキーマも再帰的に処理（ヘッダーなし）
	for _, child := range schema.Imports {
		if child.Schema != nil {
			csvStr, err := CsvGenericLinks(child.Schema, roleTypes, false)
			if err != nil {
				return "", err
			}
			buf.WriteString(csvStr)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeGenericLinks(schema *model.XBRLSchema, roleTypes map[string]model.RoleType, writer *csv.Writer) error {
	for _, linkbase := range schema.ReferencedGenericLinkbases {
		for _, elr := range linkbase.GenericLinks {
			locMap := make(map[string]model.Loc)
			for _, loc := range elr.Locs {
				locMap[loc.Label] = loc
			}
			labelMap := make(map[string]model.GenericLabel)
			for _, label := range elr.Labels {
				labelMap[label.Label] = label
			}
			for _, arc := range elr.Arcs {
				loc, ok := locMap[arc.From]
				if !ok {
					return fmt.Errorf("Arc invalid: from=%s", arc.From)
				}
				label, ok := labelMap[arc.To]
				if !ok {
					return fmt.Errorf("Arc invalid: to=%s", arc.To)
				}
				key := parser.ResolveHref(linkbase.Path, loc.Href)
				rt, ok := roleTypes[key]
				if !ok {
					return fmt.Errorf("Loc invalid: %s", key)
				}
				record := []string{
					rt.Schema.Path,
					rt.Id,
					rt.Schema.TargetNS,
					rt.RoleURI,
					rt.Definition.Value,
					label.Value,
				}
				writer.Write(record)
			}
		}
	}
	return nil
}

// 全要素csv形式文字列作成
func CsvElements(schema *model.XBRLSchema, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"Path", "Id", "TargetNamespace", "Name", "Type", "SubstitutionGroup",
			"Abstract", "Nillable", "PeriodType",
		})
	}

	for _, element := range schema.Elements {
		record := []string{
			schema.Path,
			element.Id,
			schema.TargetNS,
			element.Name,
			element.Type,
			element.SubstitutionGroup,
			element.Abstract,
			element.Nillable,
			element.PeriodType,
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	for _, child := range schema.Imports {
		if child.Schema != nil {
			childCSV, err := CsvElements(child.Schema, false)
			if err != nil {
				return "", err
			}
			buf.WriteString(childCSV)
		}
	}

	return buf.String(), nil
}

// DTSの全ラベルcsv形式文字列作成
func CsvLabels(schema *model.XBRLSchema, elements map[string]*model.XMLElement, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"TargetNamespace", "Name", "URI", "Id", "Lang", "Value", "Role",
		})
	}

	// メインのラベルリンクベース処理
	if err := writeLabels(schema, elements, writer); err != nil {
		return "", err
	}

	// 再帰的にインポート先を処理（ヘッダー無しで）
	for _, child := range schema.Imports {
		if child.Schema != nil {
			childCSV, err := CsvLabels(child.Schema, elements, false)
			if err != nil {
				return "", err
			}
			buf.WriteString(childCSV)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeLabels(schema *model.XBRLSchema, elements map[string]*model.XMLElement, writer *csv.Writer) error {
	for _, linkbase := range schema.ReferencedLabelLinkbases {
		for _, elr := range linkbase.LabelLinks {
			locMap := map[string]model.Loc{}
			for _, loc := range elr.Locs {
				locMap[loc.Label] = loc
			}
			labelMap := map[string]model.LabelLabel{}
			for _, label := range elr.Labels {
				labelMap[label.Label] = label
			}
			for _, arc := range elr.Arcs {
				loc, ok := locMap[arc.From]
				if !ok {
					return fmt.Errorf("Arc invalid: from=%s", arc.From)
				}
				label, ok := labelMap[arc.To]
				if !ok {
					return fmt.Errorf("Arc invalid: to=%s", arc.To)
				}
				key := parser.ResolveHref(linkbase.Path, loc.Href)
				elem, ok := elements[key]
				if !ok {
					return fmt.Errorf("Loc invalid: %s", key)
				}
				writer.Write([]string{
					elem.Schema.TargetNS,
					elem.Name,
					elem.Schema.Path,
					elem.Id,
					label.Lang,
					label.Value,
					label.Role,
				})
			}
		}
	}
	return nil
}

// DTSの表示リンクcsv形式文字列作成
func CsvPresentationLinks(schema *model.XBRLSchema, elements map[string]*model.XMLElement, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"Role", "FromTargetNamespace", "FromName", "FromURI", "FromId",
			"ToTargetNamespace", "ToName", "ToURI", "ToId", "Order", "PreferredLabel",
		})
	}

	for _, linkbase := range schema.ReferencedPresentationLinkbases {
		for _, elr := range linkbase.PresentationLinks {
			locMap := make(map[string]model.Loc)
			for _, _loc := range elr.Locs {
				locMap[_loc.Label] = _loc
			}

			for _, arc := range elr.Arcs {
				locFrom, exists := locMap[arc.From]
				if !exists {
					return "", fmt.Errorf("Arc invalid: from=%s", arc.From)
				}
				locTo, exists := locMap[arc.To]
				if !exists {
					return "", fmt.Errorf("Arc invalid: to=%s", arc.To)
				}

				keyFrom := parser.ResolveHref(linkbase.Path, locFrom.Href)
				elemFrom, exists := elements[keyFrom]
				if !exists {
					return "", fmt.Errorf("Loc invalid: %s", keyFrom)
				}
				keyTo := parser.ResolveHref(linkbase.Path, locTo.Href)
				elemTo, exists := elements[keyTo]
				if !exists {
					return "", fmt.Errorf("Loc invalid: %s", keyTo)
				}

				record := []string{
					elr.Role,
					elemFrom.Schema.TargetNS,
					elemFrom.Name,
					elemFrom.Schema.Path,
					elemFrom.Id,
					elemTo.Schema.TargetNS,
					elemTo.Name,
					elemTo.Schema.Path,
					elemTo.Id,
					arc.Order,
					arc.PreferredLabel,
				}
				writer.Write(record)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func sanitizeLongValue(input string) string {
	// 改行をすべて削除（CR, LF 両方対応）
	noNewlines := strings.ReplaceAll(input, "\r", "")
	noNewlines = strings.ReplaceAll(noNewlines, "\n", "")

	// 先頭100文字を取得（rune単位で安全に）
	runes := []rune(noNewlines)
	if len(runes) > 100 {
		runes = runes[:100]
	}
	head := string(runes)

	return head
}

// 全ファクトcsv形式文字列作成
func CsvFacts(instance *model.XBRLInstance, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"TargetNamespace", "Name", "Value", "ContextRef", "Decimals", "UnitRef", "Nil",
		})
	}

	for _, fact := range instance.Facts {
		record := []string{
			fact.XMLName.Space,
			fact.XMLName.Local,
			sanitizeLongValue(fact.Value),
			fact.ContextRef,
			fact.Decimals,
			fact.UnitRef,
			fact.Nil,
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// 全コンテキストcsv形式文字列作成
func CsvContexts(instance *model.XBRLInstance, withheader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if withheader {
		writer.Write([]string{
			"Id", "Identifier", "StartDate", "EndDate", "Instant",
			"Dimension-1", "Member-1", "Dimension-2", "Member-2", "Dimension-3", "Member-3",
		})
	}

	for _, context := range instance.Contexts {
		var dims, mems [3]string
		for i := 0; i < len(context.Scenario.Members) && i < 3; i++ {
			dims[i] = context.Scenario.Members[i].Dimension
			mems[i] = context.Scenario.Members[i].Value
		}

		record := []string{
			context.ID,
			context.Entity.Identifier.Value,
			context.Period.StartDate,
			context.Period.EndDate,
			context.Period.Instant,
			dims[0], mems[0],
			dims[1], mems[1],
			dims[2], mems[2],
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
