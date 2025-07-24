package facts

import (
	"fmt"
	"strings"
	"thermal/session"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type FactsCommand struct{}

func New() *FactsCommand {
	return &FactsCommand{}
}

func sanitizeLongValue(input string) string {
	// 改行をすべて削除（CR, LF 両方対応）
	noNewlines := strings.ReplaceAll(input, "\r", "")
	noNewlines = strings.ReplaceAll(noNewlines, "\n", "")

	// 先頭100文字を取得（rune単位で安全に）
	runes := []rune(noNewlines)
	if len(runes) > 100 {
		runes = runes[:100]
		return string(runes) + "…"
	}

	return noNewlines
}

type OutputFact struct {
	Element    string `yaml:"Element"`
	ContextRef string `yaml:"Context"`
	UnitRef    string `yaml:"Unit"`
	Decimals   string `yaml:"Decimals"`
	Nil        string `yaml:"Nil"`
	Length     int    `yaml:"Length"`
	Value      string `yaml:"Value"`
}

func (c *FactsCommand) Execute(s *session.Session, args string) {
	if s.Instance == nil {
		return
	}

	outputFacts := make([]OutputFact, len(s.Instance.Facts))

	for i, fact := range s.Instance.Facts {

		val := sanitizeLongValue(fact.Value)
		name := fmt.Sprintf("{%s}%s", fact.XMLName.Space, fact.XMLName.Local)
		outputFacts[i] = OutputFact{
			Element:    name,
			ContextRef: fact.ContextRef,
			UnitRef:    fact.UnitRef,
			Decimals:   fact.Decimals,
			Nil:        fact.Nil,
			Length:     utf8.RuneCountInString(fact.Value),
			Value:      val,
		}
	}

	encoder := yaml.NewEncoder(s.Stdout)
	encoder.SetIndent(2) // 読みやすさのためにインデント設定

	if err := encoder.Encode(outputFacts); err != nil {
		fmt.Fprintf(s.Stderr, "YAML encode error: %v\n", err)
	}
}
