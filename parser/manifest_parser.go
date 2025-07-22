package parser

import (
	"fmt"
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

		// インスタンスの解析
		xbrlInstance, err := ParseInstance(instanceFile)
		if err != nil {
			return nil, fmt.Errorf("❌ XBRLインスタンスのパースに失敗:%b", err)
		}
		manifest.List.XBRLInstances = append(manifest.List.XBRLInstances, xbrlInstance)
	}
	return manifest, nil
}
