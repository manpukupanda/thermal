package model

import (
	"encoding/xml"
)

// マニフェスト構造体の定義
type Manifest struct {
	Path    string         // マニフェストファイル名
	XMLName xml.Name       `xml:"manifest"`
	Toc     TocComposition `xml:"tocComposition"`
	List    List           `xml:"list"`
}

type TocComposition struct {
	Title []Title `xml:"title"`
	Item  []Item  `xml:"item"`
}

type Title struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

type Item struct {
	In      string `xml:"in,attr"`
	Ref     string `xml:"ref,attr"`
	ExtRole string `xml:"extrole,attr"`
}

type List struct {
	Instances     []Instance      `xml:"instance"`
	XBRLInstances []*XBRLInstance // パースしたXBRLInstance
}

type Instance struct {
	ID                string   `xml:"id,attr"`
	Type              string   `xml:"type,attr"`
	PreferredFilename string   `xml:"preferredFilename,attr"`
	IXBRLFiles        []string `xml:"ixbrl"`
}
