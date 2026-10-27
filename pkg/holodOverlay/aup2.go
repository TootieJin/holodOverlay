package holodOverlay

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed main_16-9.aup2
var rawBaseAup2 []byte

func WriteAup2Files(assets string, destDir string, textAndEffects []string, font []string, chartId string) error {
	baseAup2 := string(rawBaseAup2)

	/*
		textAndEffects = {
			diffCol,
			textDifficulty,
			textAuthor, stellarAuthor
			textVocals,
			textLyrics, stellarLyrics
			textComposer, stellarComposer
			textArranger, stellarArranger
			textExtra,
			textTitle
		}
	*/

	mapping := []string{
		"{assets}", strings.ReplaceAll(assets, "\\", "/"),
		"{dist}", strings.ReplaceAll(destDir, "\\", "/"),
		"{diffCol}", textAndEffects[0],
		"{text:difficulty}", textAndEffects[1],
		"{text:author}", textAndEffects[2], "{stellar:author}", textAndEffects[3],
		"{text:vocals}", textAndEffects[4],
		"{text:lyrics}", textAndEffects[5], "{stellar:lyrics}", textAndEffects[6],
		"{text:composer}", textAndEffects[7], "{stellar:composer}", textAndEffects[8],
		"{text:arranger}", textAndEffects[9], "{stellar:arranger}", textAndEffects[10],
		"{text:extra}", textAndEffects[11],
		"{text:title}", textAndEffects[12],
		"{font-1}", font[0], "{font-2}", font[1],
	}

	for i := range mapping {
		if i%2 == 0 {
			continue
		}
		if !strings.Contains(baseAup2, mapping[i-1]) {
			panic(fmt.Sprintf("aup2ファイルの生成に失敗しました (Failed to generate aup2 file) [Missing: %s]", mapping[i-1]))
		}
		baseAup2 = strings.ReplaceAll(baseAup2, mapping[i-1], mapping[i])
	}

	if err := os.WriteFile(filepath.Join(destDir, "main_16-9_"+chartId+".aup2"),
		[]byte(baseAup2),
		0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました (Failed to write file) [%w]", err)
	}
	return nil
}
