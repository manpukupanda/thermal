package model

import (
	"encoding/xml"
)

// XBRLスキーマのルート構造
type XBRLSchema struct {
	Path                            string                  // スキーマファイル名
	XMLName                         xml.Name                `xml:"schema"`
	TargetNS                        string                  `xml:"targetNamespace,attr"`
	Elements                        []XMLElement            `xml:"element"`
	Imports                         []XMLImport             `xml:"import"`
	LinkbaseRefs                    []LinkbaseRef           `xml:"annotation>appinfo>linkbaseRef"`
	RoleTypes                       []RoleType              `xml:"annotation>appinfo>roleType"`
	ReferencedLabelLinkbases        []*LabelLinkBase        // LinkbaseRef で参照している名称リンク
	ReferencedReferenceLinkbases    []*ReferenceLinkBase    // LinkbaseRef で参照している参照リンク
	ReferencedPresentationLinkbases []*PresentationLinkBase // LinkbaseRef で参照している表示リンク
	ReferencedDefinitionLinkbases   []*DefinitionLinkBase   // LinkbaseRef で参照している定義リンク
	ReferencedCalculationLinkbases  []*CalculationLinkBase  // LinkbaseRef で参照している計算リンク
	ReferencedGenericLinkbases      []*GenericLinkBase      // LinkbaseRef で参照しているジェネリックリンク
}

// スキーマ定義の要素
type XMLElement struct {
	Id                string `xml:"id,attr"`
	Name              string `xml:"name,attr"`
	Type              string `xml:"type,attr"`
	SubstitutionGroup string `xml:"substitutionGroup,attr"`
	Abstract          string `xml:"abstract,attr"`
	Nillable          string `xml:"nillable,attr"`
	PeriodType        string `xml:"periodType,attr"`
	Schema            *XBRLSchema
}

// インポート情報
type XMLImport struct {
	Namespace string `xml:"namespace,attr"`
	SchemaLoc string `xml:"schemaLocation,attr"`
	Schema    *XBRLSchema
}

// 🔗 リンクベース参照構造
type LinkbaseRef struct {
	Href    string `xml:"href,attr"`    // 参照先
	Role    string `xml:"role,attr"`    // リンクベースの種類
	ArcRole string `xml:"arcrole,attr"` // アークロール情報
}

// 🔗 ロールタイプ構造
type RoleType struct {
	RoleURI    string             `xml:"roleURI,attr"`
	Id         string             `xml:"id,attr"`
	Definition RoleTypeDefinition `xml:"definition"`
	UsedOns    []RoleTypeUsedOn   `xml:"usedOn"`
	Schema     *XBRLSchema
}

type RoleTypeDefinition struct {
	Value string `xml:",chardata"`
}

type RoleTypeUsedOn struct {
	Value string `xml:",chardata"`
}
