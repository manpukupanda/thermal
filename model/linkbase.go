package model

import (
	"encoding/xml"
)

// リンクベース共通
type LinkBase struct {
	Path    string   // リンクベースファイル名
	XMLName xml.Name `xml:"linkbase"`
}

// Arc共通
type ArcBase struct {
	From string `xml:"from,attr"`
	To   string `xml:"to,attr"`
}

// ロケータ
type Loc struct {
	Label string `xml:"label,attr"`
	Href  string `xml:"href,attr"`
}

// /////////////////////////////////////////////////////////////
// 名称リンクベース
// /////////////////////////////////////////////////////////////
// 🔖 名称リンクベース全体の構造
type LabelLinkBase struct {
	LinkBase
	LabelLinks []LabelLink `xml:"labelLink"`
}

// 🔖 個別の名称リンク
type LabelLink struct {
	XMLName xml.Name     `xml:"labelLink"`
	Role    string       `xml:"role,attr"`
	Arcs    []LabelArc   `xml:"labelArc"`
	Locs    []Loc        `xml:"loc"`
	Labels  []LabelLabel `xml:"label"`
}

// 🔖 名称リンクのアーク
type LabelArc struct {
	ArcBase
}

// 🔖 名称リンクの名称
type LabelLabel struct {
	Label    string `xml:"label,attr"`
	Lang     string `xml:"lang,attr"`
	Role     string `xml:"role,attr"`
	Id       string `xml:"id,attr"`
	Value    string `xml:",chardata"`
	LinkBase *LabelLinkBase
}

// /////////////////////////////////////////////////////////////
// 参照リンクベース
// /////////////////////////////////////////////////////////////
// 参照リンクベース全体の構造
type ReferenceLinkBase struct {
	LinkBase
	ReferenceLinks []ReferenceLink `xml:"referenceLink"`
}

// 個別の参照リンク
type ReferenceLink struct {
	XMLName    xml.Name             `xml:"referenceLink"`
	Role       string               `xml:"role,attr"`
	Arcs       []ReferenceArc       `xml:"referenceArc"`
	Locs       []Loc                `xml:"loc"`
	References []ReferenceReference `xml:"reference"`
}

// 参照リンクのアーク
type ReferenceArc struct {
	ArcBase
}

// 参照リンクの参照
type ReferenceReference struct {
	Label                string `xml:"label,attr"`
	Role                 string `xml:"role,attr"`
	Publisher            string `xml:"Publisher"`
	Number               string `xml:"Number"`
	Name                 string `xml:"Name"`
	Article              string `xml:"Article"`
	IssueDate            string `xml:"IssueDate"`
	IndustryAbbreviation string `xml:"IndustryAbbreviation"`
	LinkBase             *ReferenceLinkBase
}

// /////////////////////////////////////////////////////////////
// 表示リンクベース
// /////////////////////////////////////////////////////////////
// 🌲 表示リンクベース全体の構造
type PresentationLinkBase struct {
	LinkBase
	PresentationLinks []PresentationLink `xml:"presentationLink"`
}

// 🌲 個別の表示リンク
type PresentationLink struct {
	XMLName xml.Name          `xml:"presentationLink"`
	Role    string            `xml:"role,attr"`
	Arcs    []PresentationArc `xml:"presentationArc"`
	Locs    []Loc             `xml:"loc"`
}

// 🌲 表示リンクのアーク（関係）
type PresentationArc struct {
	ArcBase
	Order          string `xml:"order,attr"`
	PreferredLabel string `xml:"preferredLabel,attr"`
}

// /////////////////////////////////////////////////////////////
// 定義リンクベース
// /////////////////////////////////////////////////////////////
// 🧩 定義リンクベース全体の構造
type DefinitionLinkBase struct {
	LinkBase
	DefinitionLinks []DefinitionLink `xml:"definitionLink"`
}

// 🧩 個別の定義リンク
type DefinitionLink struct {
	XMLName xml.Name        `xml:"definitionLink"`
	Role    string          `xml:"role,attr"`
	Arcs    []DefinitionArc `xml:"definitionArc"`
	Locs    []Loc           `xml:"loc"`
}

// 🧩 定義リンクのアーク（要素間の関係）
type DefinitionArc struct {
	ArcBase
	ArcRole string `xml:"arcrole,attr"`
	Order   string `xml:"order,attr"`
}

// /////////////////////////////////////////////////////////////
// 計算リンクベース
// /////////////////////////////////////////////////////////////
// ➕ 計算リンクベース全体の構造
type CalculationLinkBase struct {
	LinkBase
	CalculationLinks []CalculationLink `xml:"calculationLink"`
}

// ➕ 個別の計算リンク
type CalculationLink struct {
	XMLName xml.Name         `xml:"calculationLink"`
	Role    string           `xml:"role,attr"`
	Arcs    []CalculationArc `xml:"calculationArc"`
	Locs    []Loc            `xml:"loc"`
}

// ➕ 計算リンクのアーク（要素間の関係）
type CalculationArc struct {
	ArcBase
	ArcRole string  `xml:"arcrole,attr"`
	Order   float64 `xml:"order,attr"`
}

// /////////////////////////////////////////////////////////////
// ジェネリックリンクベース
// /////////////////////////////////////////////////////////////
// 🔗 ジェネリックリンクベース全体の構造
type GenericLinkBase struct {
	LinkBase
	GenericLinks []GenericLink `xml:"link"`
}

// 🔗 個別のジェネリックリンク
type GenericLink struct {
	XMLName xml.Name       `xml:"link"`
	Role    string         `xml:"role,attr"`
	Arcs    []GenericArc   `xml:"arc"`
	Locs    []Loc          `xml:"loc"`
	Labels  []GenericLabel `xml:"label"`
}

// 🔗 ジェネリックリンクのアーク
type GenericArc struct {
	ArcBase
}

// 🔗 ジェネリックリンクのラベル
type GenericLabel struct {
	Label string `xml:"label,attr"`
	Lang  string `xml:"lang,attr"`
	Role  string `xml:"role,attr"`
	Value string `xml:",chardata"`
}
