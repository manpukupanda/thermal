package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"thermal/model"
)

func ParseManifest(path string) (*model.Manifest, error) {
	// マニフェストの解析
	manifest, err := ParseXML[model.Manifest](path)
	if err != nil {
		return nil, fmt.Errorf("❌ マニフェストファイルのパースに失敗:%v", err)
	}

	manifest.Path = path

	// インスタンスファイルの取得
	if len(manifest.List.Instances) == 0 {
		return nil, fmt.Errorf("❌ インスタンスファイルが見つかりません")
	}

	for i := range manifest.List.Instances {
		instanceFilename := manifest.List.Instances[i].PreferredFilename
		instanceFile := filepath.Join(filepath.Dir(manifest.Path), instanceFilename)

		// インスタンスがリモートファイルでなく、かつ、存在しなければ、InlineXBRLを代わりに読み込む
		readFromIXBRL := false
		if !IsRemoteFile(instanceFile) {
			_, err := os.Stat(instanceFile)
			if os.IsNotExist(err) {
				readFromIXBRL = true
			} else if err != nil {
				return nil, fmt.Errorf("❌ XBRLインスタンスの存在チェックに失敗:%v", err)
			}
		}

		// インスタンスの解析
		if readFromIXBRL {
			// InlineXBRLを読む処理
			inlineXBRLsPaths := make([]string, len(manifest.List.Instances[i].IXBRLFiles))
			for j, path := range manifest.List.Instances[i].IXBRLFiles {
				inlineXBRLsPaths[j] = filepath.Join(filepath.Dir(manifest.Path), path)
			}

			xbrlInstance, err := ParseInlineXBRLs(inlineXBRLsPaths, instanceFile)
			if err != nil {
				return nil, fmt.Errorf("❌ Inline XBRLのパースに失敗:%v", err)
			}
			manifest.List.XBRLInstances = append(manifest.List.XBRLInstances, xbrlInstance)
		} else {
			xbrlInstance, err := ParseInstance(instanceFile)
			if err != nil {
				return nil, fmt.Errorf("❌ XBRLインスタンスのパースに失敗:%v", err)
			}
			manifest.List.XBRLInstances = append(manifest.List.XBRLInstances, xbrlInstance)
		}
	}
	return manifest, nil
}
