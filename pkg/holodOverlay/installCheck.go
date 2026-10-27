package holodOverlay

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed hldr.obj2
var hldrObj2 []byte

//go:embed English.holodOverlay.aul2
var enAul2 []byte

func TryInstallObject(aviutl2Path string, font []string) bool {
	if aviutl2Path == "" {
		return false
	}

	// object
	hldrObj2Path := filepath.Join(aviutl2Path, "Script", "@holodOverlay.obj2")
	if err := os.MkdirAll(filepath.Join(aviutl2Path, "Script"), 0755); err != nil {
		return false
	}

	if data, err := os.ReadFile(hldrObj2Path); err == nil {
		if strings.Contains(string(data), "-- holodOverlay by TootieJin (v"+Version+")") && Version != "0.0.0" {
			return false
		}
	}

	hldrObj2File, err := os.Create(hldrObj2Path)
	if err != nil {
		return false
	}
	defer hldrObj2File.Close()

	hldrObj2Writer := strings.NewReplacer(
		"\r\n", "\r\n",
		"\r", "\r\n",
		"\n", "\r\n",
		"{version}", Version,
		"{font-1}", font[0],
		"{font-2}", font[1],
	).Replace(string(hldrObj2))
	if _, err := io.WriteString(hldrObj2File, hldrObj2Writer); err != nil {
		return false
	}

	// language
	enAul2Path := filepath.Join(aviutl2Path, "Language", "English.holodOverlay.aul2")
	if err := os.MkdirAll(filepath.Join(aviutl2Path, "Language"), 0755); err != nil {
		return false
	}

	if data, err := os.ReadFile(enAul2Path); err == nil {
		if strings.Contains(string(data), "; holodOverlay by TootieJin (v"+Version+")") && Version != "0.0.0" {
			return false
		}
	}

	enAul2File, err := os.Create(enAul2Path)
	if err != nil {
		return false
	}
	defer enAul2File.Close()

	enAul2Writer := strings.NewReplacer(
		"\r\n", "\r\n",
		"\r", "\r\n",
		"\n", "\r\n",
		"{version}", Version,
	).Replace(string(enAul2))
	if _, err := io.WriteString(enAul2File, enAul2Writer); err != nil {
		return false
	}

	return true
}

func DownloadENLang(destPath string) bool {
	resp, err := http.Get("https://raw.githubusercontent.com/aviutl2/aviutl2_community_translation/refs/heads/main/locales/community_en.aul2")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}

	bodyStr := string(body)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "4") || strings.HasPrefix(strings.TrimSpace(bodyStr), "5") {
		return false
	}

	file, err := os.Create(destPath)
	if err != nil {
		return false
	}
	defer file.Close()

	_, err = file.WriteString(bodyStr)
	if err != nil {
		return false
	}

	return true
}

func TryInstallAviUtlDependencies(aviutl2Path string, aviutlProcess string) bool {
	if aviutl2Path == "" {
		return false
	}

	var langDest string
	langDest = filepath.Join(aviutl2Path, "Language")

	var depDir2 string
	depDir2 = filepath.Join("dependencies", "aviutl2")

	var copyDir func(src, dest string) error
	copyDir = func(src, dest string) error {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		for _, entry := range entries {
			srcPath := filepath.Join(src, entry.Name())
			destPath := filepath.Join(dest, entry.Name())
			if entry.IsDir() {
				if err := copyDir(srcPath, destPath); err != nil {
					continue
				}
			} else {
				if _, err := os.Stat(destPath); err == nil { // if file already exists
					continue
				}

				srcFile, err := os.Open(srcPath)
				if err != nil {
					continue
				}
				defer srcFile.Close()

				destFile, err := os.Create(destPath)
				if err != nil {
					srcFile.Close()
					continue
				}
				defer destFile.Close()

				_, err = io.Copy(destFile, srcFile)
				if err != nil {
					continue
				}
			}
		}
		return nil
	}

	if err := copyDir(depDir2, aviutl2Path); err != nil {
		return false
	}
	if installENLang := DownloadENLang(filepath.Join(langDest, "English.aul2")); !installENLang {
		return false
	}

	return true
}

func decryptStr(encoded string) (string, error) {
	str, err := hex.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(str))
	if err != nil {
		return "", err
	}

	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}

func Listing(url string, str string) (bool, error) {
	url, err := decryptStr(url)
	if err != nil {
		return false, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	bodyStr := string(body)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "4") || strings.HasPrefix(strings.TrimSpace(bodyStr), "5") {
		return false, fmt.Errorf("listing error: %s", bodyStr)
	}

	list := strings.Split(string(body), "\n")
	for _, name := range list {
		hashtagCount := strings.Count(str, "#")
		suffix := "#" + strings.Split(str, "#")[int(math.Max(0, float64(hashtagCount)-1))]

		if strings.TrimSpace(name) == str {
			return true, nil
		} else if strings.HasSuffix(strings.TrimSpace(name), suffix) {
			return true, nil
		} else if strings.EqualFold(strings.TrimSpace(name), strings.TrimSuffix(str, suffix)) {
			return true, nil
		}
	}

	return false, nil
}
