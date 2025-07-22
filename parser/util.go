package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func IsRemoteFile(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// ローカル又はリモートのXMLデータをメモリにロードし、`bytes.Reader` を返す共通関数
func GetXMLReader(filename string) (*bytes.Reader, error) {
	var data []byte

	// EDINETのタクソノミはローカルのキャッシュパスから取得するため置き換える
	if IsRemoteFile(filename) {
		if strings.Contains(filename, "http://disclosure.edinet-fsa.go.jp/taxonomy/") {
			base := os.Getenv("EDINET_TAXONOMY_DIR")
			if base == "" {
				base = "/app/taxonomy/all/taxonomy/" // デフォルト値
			}
			// 📂 ローカルファイルが存在する場合は、そちらを優先
			_cachedfilePath := strings.Replace(filename, "http://disclosure.edinet-fsa.go.jp/taxonomy/", base, 1)
			if _, err := os.Stat(_cachedfilePath); err == nil {
				filename = _cachedfilePath
			}
		}
	}

	// 🌐 リモート URL の場合
	if IsRemoteFile(filename) {
		resp, err := http.Get(filename)
		if err != nil {
			return nil, fmt.Errorf("❌ XML取得失敗: %s", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("❌ HTTPレスポンスエラー: %d", resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("❌ HTTPレスポンスのデータ読み込み失敗: %s", err)
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("❌ ファイルを開けません: %s", err)
		}
		defer file.Close()
		data, err = io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("❌ ファイルのデータ読み込み失敗: %s", err)
		}
	}

	// 🔍 `data` が空の場合はエラー
	if len(data) == 0 {
		return nil, fmt.Errorf("❌ 読み込んだデータが空です")
	}

	return bytes.NewReader(data), nil
}

// 参照先URL（`linkbaseRef.Href`等）を適切な URL やローカルパスに変換する関数
func ResolveHref(baseFilename, href string) string {
	// 絶対 URL はそのまま返す
	if IsRemoteFile(href) {
		return href
	}

	// ベースがリモートのファイルであればUrlをパースして相対パスを解釈する
	if IsRemoteFile(baseFilename) {
		// 🌐 URLの相対パスを解釈
		baseURLParsed, _ := url.Parse(baseFilename)
		newURLParsed, _ := baseURLParsed.Parse(href)
		return newURLParsed.String()
	}

	// 📂 ローカルファイルの相対パス処理
	return filepath.Join(filepath.Dir(baseFilename), href)
}

// ジェネリックなXMLパーサー
func ParseXML[T any](filename string) (*T, error) {
	reader, err := GetXMLReader(filename)
	if err != nil {
		return nil, err
	}

	var result T
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("XMLのパースに失敗: %w", err)
	}
	return &result, nil
}

func PeekXMLRootElementName(filename string) (string, error) {
	reader, err := GetXMLReader(filename)
	if err != nil {
		return "", err
	}

	decoder := xml.NewDecoder(reader)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

// シンプルな * ワイルドカードマッチ
func WildcardMatch(pattern, str string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	parts := strings.Split(pattern, "*")

	// 最初の部分が一致するか（前方一致）
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}
	str = str[len(parts[0]):]

	// 中間部分を順に探す
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str, parts[i])
		if idx == -1 {
			return false
		}
		str = str[idx+len(parts[i]):]
	}

	// 最後の部分が一致するか（後方一致）
	return strings.HasSuffix(str, parts[len(parts)-1])
}
