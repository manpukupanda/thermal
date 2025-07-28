package model

import (
	"encoding/xml"
)

// XBRLインスタンスのトップレベル構造
type XBRLInstance struct {
	Path         string         // インスタンスファイル名
	XMLName      xml.Name       `xml:"xbrl"`
	SchemaRefs   SchemaRef      `xml:"schemaRef"`
	RoleRefs     []RoleRef      `xml:"roleRef"`
	Contexts     []Context      `xml:"context"`
	Units        []Unit         `xml:"unit"`
	Facts        []Fact         `xml:",any"`
	FootnoteLink []FootnoteLink `xml:"footnoteLink"`
}

// スキーマ定義
type SchemaRef struct {
	Href   string `xml:"href,attr"`
	Schema *XBRLSchema
}

// roleRefタグ
type RoleRef struct {
	RoleURI string `xml:"roleURI,attr"`
	Href    string `xml:"href,attr"`
}

// コンテキスト情報
type Context struct {
	ID       string   `xml:"id,attr"`
	Entity   Entity   `xml:"entity"`
	Period   Period   `xml:"period"`
	Scenario Scenario `xml:"scenario"`
}

// 企業識別情報
type Entity struct {
	Identifier Identifier `xml:"identifier"`
}

type Identifier struct {
	Scheme string `xml:"scheme,attr"`
	Value  string `xml:",chardata"`
}

// 期間時点
type Period struct {
	StartDate string `xml:"startDate"`
	EndDate   string `xml:"endDate"`
	Instant   string `xml:"instant"`
}

// シナリオ情報（セグメントや補足情報）
type Scenario struct {
	Members []Member `xml:"explicitMember"`
}

type Member struct {
	Dimension string `xml:"dimension,attr"`
	Value     string `xml:",chardata"`
}

// 単位情報
type Unit struct {
	ID      string `xml:"id,attr"`
	Measure string `xml:"measure"`
}

// 財務データ（可変要素）
type Fact struct {
	XMLName    xml.Name `xml:""`
	ContextRef string   `xml:"contextRef,attr"`
	UnitRef    string   `xml:"unitRef,attr"`
	Decimals   string   `xml:"decimals,attr"`
	Nil        string   `xml:"nil,attr"`
	Value      string   `xml:",chardata"`
}

type FootnoteLink struct {
	Value string `xml:",chardata"`
}
