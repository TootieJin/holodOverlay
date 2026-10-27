package holodOverlay

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/TootieJin/holodOverlay/pkg/sonolus"
	"golang.org/x/image/draw"
)

type Source struct {
	Id     string
	Name   string
	Color  int
	Host   string
	Status int
}

func DetectLocalChartSource() (Source, error) {
	// ScoreSync（3939ポート）に接続を試行
	resp, err := http.Get("http://localhost:3939/")
	if err != nil {
		return Source{}, errors.New("ScoreSyncに接続できませんでした。サーバーが起動していることを確認してください。(Could not connect to ScoreSync. Please make sure the server is up and running.)")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Source{}, errors.New("ScoreSyncからの応答が正常ではありません。(Invalid response from ScoreSync.)")
	}

	source := Source{
		Id:     "local_server",
		Name:   "Local Server (ScoreSync + ScoreSync Modern)",
		Color:  0xa9cd6a,
		Host:   "localhost:3939",
		Status: 0,
	}
	return source, nil
}

func FetchChart(source Source, chartId string) (sonolus.LevelInfo, string, error) {
	var url string
	if source.Id == "local_server" {
		// ローカルサーバーの場合はchartIdをタイトルとして使用
		url = "http://" + source.Host + "/sonolus/levels/" + chartId
	} else {
		url = "https://" + source.Host + "/sonolus/levels/" + chartId
	}

	resp, err := http.Get(url)

	if err != nil {
		if source.Id == "local_server" {
			return sonolus.LevelInfo{}, "", errors.New("ScoreSyncに接続できませんでした。(Could not connect to ScoreSync.)")
		}
		return sonolus.LevelInfo{}, "", errors.New("サーバーに接続できませんでした。(Could not connect to server.)")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if source.Id == "local_server" {
			return sonolus.LevelInfo{}, "", errors.New("指定されたタイトルの譜面が見つかりませんでした。(Chart not found.)")
		}
		return sonolus.LevelInfo{}, "", errors.New("譜面が見つかりませんでした。(Unable to search chart.)")
	}

	var chart sonolus.InfoResponse[sonolus.LevelInfo]
	json.NewDecoder(resp.Body).Decode(&chart)

	return chart.Item, chart.Description, nil
}

func DetectChartSource(chartId string, chartInstance string) (Source, error) {
	var source Source
	if strings.HasPrefix(chartId, "holodori-") {
		source = Source{
			Id:     "hldr-db",
			Name:   "hololive Dreams Unofficial Database",
			Color:  0x12244f,
			Host:   "sonolus.holodori.best",
			Status: 0,
		}
	} else if strings.HasPrefix(chartId, "sekai-best-") {
		source = Source{
			Id:     "sekai_viewer",
			Name:   "Sekai Viewer",
			Color:  0x02cbbd,
			Host:   "sonolus.sekai.best",
			Status: 0,
		}
	} else if strings.HasPrefix(chartId, "sss-") {
		source = Source{
			Id:     "sbuga_sonolus",
			Name:   "Sbuga's Sonolus Server",
			Color:  0xe0f2fe,
			Host:   "sonolus.sbuga.com",
			Status: 0,
		}
	} else if strings.HasPrefix(chartId, "chcy-") {
		switch chartInstance {
		case "0":
			source = Source{
				Id:     "chart_cyanvas",
				Name:   "Chart Cyanvas Archive",
				Color:  0x83ccd2,
				Host:   "cc.milkbun.org",
				Status: 0,
			}
		case "1":
			source = Source{
				Id:     "chart_cyanvas",
				Name:   "Chart Cyanvas Offshoot Server",
				Color:  0x83ccd2,
				Host:   "chart-cyanvas.com",
				Status: 0,
			}
		}
	} else if strings.HasPrefix(chartId, "ptlv-") {
		source = Source{
			Id:     "potato_leaves",
			Name:   "Potato Leaves",
			Color:  0x88cb7f,
			Host:   "ptlv.milkbun.org",
			Status: 0,
		}
	} else if strings.HasPrefix(chartId, "utsk-") {
		source = Source{
			Id:     "untitled_sekai",
			Name:   "Untitled Sekai",
			Color:  0x6a6a6a,
			Host:   "us.pim4n-net.com",
			Status: 1,
		}
	} else if strings.HasPrefix(chartId, "UnCh-") {
		source = Source{
			Id:     "untitledcharts",
			Name:   "UntitledCharts",
			Color:  0x7765da,
			Host:   "untitledcharts.com",
			Status: 0,
		}
	} else if strings.HasPrefix(chartId, "custom-") {
		source = Source{
			Id:     "custom_server",
			Name:   "Custom Server (" + chartInstance + ")",
			Color:  0xffffff,
			Host:   chartInstance,
			Status: 0,
		}
	}
	if source.Id == "" {
		return Source{
			Id:     chartId,
			Name:   "",
			Color:  0,
			Host:   "",
			Status: 0,
		}, errors.New("unknown chart source")
	}
	return source, nil
}

func FetchLevelData(source Source, level sonolus.LevelInfo) (sonolus.LevelData, error) {
	var url string
	var err error

	if source.Id == "local_server" {
		url, err = sonolus.JoinUrl("http://"+source.Host, level.Data.Url)
	} else {
		url, err = sonolus.JoinUrl("https://"+source.Host, level.Data.Url)
	}

	if err != nil {
		return sonolus.LevelData{}, fmt.Errorf("URLの解析に失敗しました。(URL parsing failed.) [%s]", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return sonolus.LevelData{}, fmt.Errorf("サーバーに接続できませんでした。(Could not connect to server.) [%s]", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return sonolus.LevelData{}, fmt.Errorf("譜面データが見つかりませんでした。(No chart data found.) [%d]", resp.StatusCode)
	}

	var data sonolus.LevelData
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return sonolus.LevelData{}, fmt.Errorf("譜面データの読み込みに失敗しました。(Loading chart data failed.) [%s]", err)
	}

	err = json.NewDecoder(gzipReader).Decode(&data)

	if err != nil {
		return sonolus.LevelData{}, fmt.Errorf("譜面データの読み込みに失敗しました。(Loading chart data failed.) [%s]", err)
	}

	return data, nil
}

func DownloadJacket(source Source, level sonolus.LevelInfo, destPath string) error {
	var url string
	var err error

	if source.Id == "local_server" {
		url, err = sonolus.JoinUrl("http://"+source.Host, level.Cover.Url)
	} else {
		url, err = sonolus.JoinUrl("https://"+source.Host, level.Cover.Url)
	}

	if err != nil {
		return fmt.Errorf("URLの解析に失敗しました。(URL parsing failed.) [%s]", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("サーバーに接続できませんでした。(Could not connect to server.) [%s]", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("サーバーに接続できませんでした。(Could not connect to server.) [%d]", resp.StatusCode)
	}

	os.MkdirAll(destPath, 0755)
	imageData, _, err := image.Decode(resp.Body)

	if err != nil {
		return fmt.Errorf("ジャケットの読み込みに失敗しました。(Loading jacket failed.) [%s]", err)
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 2048, 2048))

	draw.ApproxBiLinear.Scale(newImage, newImage.Bounds(), imageData, imageData.Bounds(), draw.Over, nil)

	file, err := os.Create(path.Join(destPath, "jacket.png"))

	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました。(Failed to create file.) [%s]", err)
	}

	defer file.Close()

	err = png.Encode(file, newImage)

	if err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました。(Failed to write file.) [%s]", err)
	}

	return nil
}

func DownloadPreview(source Source, level sonolus.LevelInfo, destPath string) error {
	var previewUrl string
	var err error

	if source.Id == "local_server" {
		previewUrl, err = sonolus.JoinUrl("http://"+source.Host, level.Preview.Url)
	} else {
		previewUrl, err = sonolus.JoinUrl("https://"+source.Host, level.Preview.Url)
	}

	if err != nil {
		return fmt.Errorf("URLの解析に失敗しました。(URL parsing failed.) [%s]", err)
	}

	resp, err := http.Get(previewUrl)
	if err != nil {
		return fmt.Errorf("サーバーに接続できませんでした。(Could not connect to server.) [%s]", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("音声が見つかりませんでした。(Audio not found.) [%d]", resp.StatusCode)
	}

	var file *os.File
	file, err = os.Create(path.Join(destPath, "preview.mp3"))

	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました。(Failed to create file.) [%s]", err)
	}

	defer file.Close()

	if file != nil {
		if _, err := io.Copy(file, resp.Body); err != nil {
			return fmt.Errorf("ファイルの書き込みに失敗しました。(Failed to write file.) [%s]", err)
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	return nil
}
