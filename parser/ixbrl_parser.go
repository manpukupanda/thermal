package parser

import (
	"encoding/xml"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"thermal/model"

	"github.com/antchfx/xmlquery"
)

// 名前空間情報を抽出する
func extractNamespaceMap(doc *xmlquery.Node) map[string]string {
	nsMap := make(map[string]string)
	root := doc.SelectElement("*") // 最初のルート要素
	for _, attr := range root.Attr {
		if attr.Name.Local == "xmlns" {
			nsMap["(default)"] = attr.Value
		} else if attr.Name.Space == "xmlns" {
			nsMap[attr.Name.Local] = attr.Value
		} else {
			continue
		}
	}
	return nsMap
}

func getPrefixByNamespaceURI(nsMap map[string]string, targetURI string) string {
	for prefix, uri := range nsMap {
		if uri == targetURI {
			return prefix
		}
	}
	return ""
}

func resolveXMLName(tag string, nsMap map[string]string) xml.Name {
	var space, local string
	parts := strings.SplitN(tag, ":", 2)

	if len(parts) == 2 {
		// プレフィックスあり (例: ns1:hogehoge)
		prefix := parts[0]
		local = parts[1]
		if uri, ok := nsMap[prefix]; ok {
			space = uri
		}
	} else {
		// プレフィックスなし → デフォルト名前空間
		local = tag
		if uri, ok := nsMap["(default)"]; ok {
			space = uri
		}
	}

	return xml.Name{
		Space: space,
		Local: local,
	}
}

func ShiftDecimal(value string, n int) string {
	// 文字列 → float64
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "NaN"
	}

	// 10のn乗をかける
	shifted := f * math.Pow10(n)

	// 小数点以下が不要なら .0 を消す
	result := strconv.FormatFloat(shifted, 'f', -1, 64)
	return result
}

func warekiToSeireki(date string) (string, error) {
	eras := map[string]int{
		"明治": 1868,
		"大正": 1912,
		"昭和": 1926,
		"平成": 1989,
		"令和": 2019,
	}

	toHalfWidth := func(s string) string {
		r := strings.NewReplacer(
			"０", "0", "１", "1", "２", "2", "３", "3", "４", "4",
			"５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
		)
		return r.Replace(s)
	}

	// 年月日 or 年月 に対応
	re := regexp.MustCompile(`(明治|大正|昭和|平成|令和)(元|[0-9０-９]+)年([0-9０-９]+)月(?:([0-9０-９]+)日)?`)
	matches := re.FindStringSubmatch(date)
	if matches == nil {
		return "", fmt.Errorf("不正な形式: %s", date)
	}

	era := matches[1]
	yearStr := matches[2]
	monthStr := matches[3]
	dayStr := matches[4] // 日付は空の場合も

	if yearStr == "元" {
		yearStr = "1"
	}

	y, _ := strconv.Atoi(toHalfWidth(yearStr))
	m, _ := strconv.Atoi(toHalfWidth(monthStr))
	seireki := eras[era] + y - 1

	if dayStr == "" {
		return fmt.Sprintf("%04d-%02d", seireki, m), nil
	}

	d, _ := strconv.Atoi(toHalfWidth(dayStr))
	return fmt.Sprintf("%04d-%02d-%02d", seireki, m, d), nil
}

func jpDateToISO(jp string) (string, error) {
	toHalf := func(s string) string {
		return strings.NewReplacer(
			"０", "0", "１", "1", "２", "2", "３", "3", "４", "4",
			"５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
		).Replace(s)
	}

	// 年月日 or 年月 のパターン対応
	re := regexp.MustCompile(`([0-9０-９]+)年([0-9０-９]+)月(?:([0-9０-９]+)日)?`)
	matches := re.FindStringSubmatch(jp)
	if matches == nil {
		return "", fmt.Errorf("不正な形式: %s", jp)
	}

	y, _ := strconv.Atoi(toHalf(matches[1]))
	m, _ := strconv.Atoi(toHalf(matches[2]))

	if matches[3] == "" {
		return fmt.Sprintf("%04d-%02d", y, m), nil // 年月だけ
	}

	d, _ := strconv.Atoi(toHalf(matches[3]))
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d), nil // 年月日
}

func parseInlineXBRL(inlineXBRLFile string, instance *model.XBRLInstance) error {

	r, err := GetXMLReader(inlineXBRLFile)
	if err != nil {
		return err
	}

	doc, err := xmlquery.Parse(r)
	if err != nil {
		return err
	}

	// 名前空間対応表作成
	nsMap := extractNamespaceMap(doc)

	for _, node := range xmlquery.Find(doc, "//*") {
		if (node.Data == "nonNumeric" || node.Data == "nonFraction") && node.NamespaceURI == "http://www.xbrl.org/2008/inlineXBRL" {
			// Fact
			name := node.SelectAttr("name")
			xsi := getPrefixByNamespaceURI(nsMap, "http://www.w3.org/2001/XMLSchema-instance")
			xsinil := fmt.Sprintf("%s:nil", xsi)
			escape := node.SelectAttr("escape")
			sign := node.SelectAttr("sign")
			text := ""
			if xsinil != "true" {
				if escape == "true" {
					// テキストブロックの場合（タグごと）
					innerXML := ""
					for child := node.FirstChild; child != nil; child = child.NextSibling {
						innerXML += child.OutputXML(true)
					}
					text = innerXML
				} else {
					text = node.InnerText()
					format := node.SelectAttr("format")
					if strings.HasSuffix(format, ":numdotdecimal") {
						text = strings.ReplaceAll(text, ",", "")
					} else if strings.HasSuffix(format, ":dateerayearmonthdayjp") {
						// 和暦年月日変換
						s, err := warekiToSeireki(text)
						if err != nil {
							text = s
						}
					} else if strings.HasSuffix(format, ":dateerayearmonthjp") {
						// 和暦年月変換
						s, err := warekiToSeireki(text)
						if err != nil {
							text = s
						}
					} else if strings.HasSuffix(format, ":dateyearmonthdaycjk") {
						// 年月日変換
						s, err := jpDateToISO(text)
						if err != nil {
							text = s
						}
					} else if strings.HasSuffix(format, ":dateyearmonthcjk") {
						// 年月変換
						s, err := jpDateToISO(text)
						if err != nil {
							text = s
						}
					}

					scale := node.SelectAttr("scale")
					scaleNum, err := strconv.Atoi(scale)
					if err == nil && scaleNum != 0 {
						text = ShiftDecimal(text, scaleNum)
					}
					text = sign + text
				}
			}
			// TODO:トランスフォーメーションルールの実装
			instance.Facts = append(instance.Facts, model.Fact{
				XMLName:    resolveXMLName(name, nsMap),
				ContextRef: node.SelectAttr("contextRef"),
				UnitRef:    node.SelectAttr("unitRef"),
				Decimals:   node.SelectAttr("decimals"),
				Nil:        node.SelectAttr(xsinil),
				Value:      text,
			})
		} else if node.Data == "schemaRef" && node.NamespaceURI == "http://www.xbrl.org/2003/linkbase" {
			// schemaRef要素の xlink:href 属性の値を取得
			xlink := getPrefixByNamespaceURI(nsMap, "http://www.w3.org/1999/xlink")
			xlinkhref := fmt.Sprintf("%s:href", xlink)
			instance.SchemaRefs.Href = node.SelectAttr(xlinkhref)
		} else if node.Data == "context" && node.NamespaceURI == "http://www.xbrl.org/2003/instance" {
			xmlstr := node.OutputXML(true)
			reader := strings.NewReader(xmlstr)
			decoder := xml.NewDecoder(reader)
			var context model.Context
			if err := decoder.Decode(&context); err != nil {
				return err
			}
			instance.Contexts = append(instance.Contexts, context)
		} else if node.Data == "unit" && node.NamespaceURI == "http://www.xbrl.org/2003/instance" {
			xmlstr := node.OutputXML(true)
			reader := strings.NewReader(xmlstr)
			decoder := xml.NewDecoder(reader)
			var unit model.Unit
			if err := decoder.Decode(&unit); err != nil {
				return err
			}
			instance.Units = append(instance.Units, unit)
		} else if node.Data == "roleRef" && node.NamespaceURI == "http://www.xbrl.org/2003/linkbase" {
			xlink := getPrefixByNamespaceURI(nsMap, "http://www.w3.org/1999/xlink")
			xlinkhref := fmt.Sprintf("%s:href", xlink)
			instance.RoleRefs = append(instance.RoleRefs, model.RoleRef{
				RoleURI: node.SelectAttr("roleURI"),
				Href:    node.SelectAttr(xlinkhref),
			})
		}
	}

	return nil
}

func ParseInlineXBRLs(inlineXBRLFiles []string, instanceFile string) (*model.XBRLInstance, error) {

	xbrlInstance := &model.XBRLInstance{}

	for _, inlineXBRLFile := range inlineXBRLFiles {
		err := parseInlineXBRL(inlineXBRLFile, xbrlInstance)
		if err != nil {
			return nil, fmt.Errorf("❌ InlineXBRLのパースに失敗:%v", err)
		}
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
