package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/TootieJin/holodOverlay/pkg/holodOverlay"
	"github.com/TootieJin/holodOverlay/pkg/sonolus"
	"github.com/fatih/color"
	"github.com/google/go-github/v57/github"
	"github.com/srinathh/gokilo/rawmode"
	"golang.org/x/sys/windows"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func checkSubstrings(str []string, subs ...string) string {
	for _, s := range str {
		for _, sub := range subs {
			if strings.Contains(s, sub) {
				return sub
			}
		}
	}
	return ""
}

func matchedStringIndex(str []string, target string) int {
	for i, s := range str {
		if s == target {
			return i
		}
	}
	return -1
}

func shouldCheckUpdate() bool {
	executablePath, err := os.Executable()
	if err != nil {
		return false
	}
	updateCheckFile, err := os.OpenFile(filepath.Join(filepath.Dir(executablePath), ".update-check"), os.O_RDONLY, 0666)
	if err != nil {
		return os.IsNotExist(err)
	}
	defer updateCheckFile.Close()

	scanner := bufio.NewScanner(updateCheckFile)
	scanner.Scan()
	lastCheckTime, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		return false
	}

	return time.Now().Unix()-lastCheckTime > 60*60*24*3
}

func checkUpdate() bool {
	githubClient := github.NewClient(nil)
	release, _, err := githubClient.Repositories.GetLatestRelease(context.Background(), "TootieJin", "holodOverlay")
	if err != nil {
		return false
	}

	executablePath, err := os.Executable()
	updateCheckFile, err := os.OpenFile(filepath.Join(filepath.Dir(executablePath), ".update-check"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return false
	}
	defer updateCheckFile.Close()
	updateCheckFile.WriteString(strconv.FormatInt(time.Now().Unix(), 10))

	latestVersion := strings.TrimPrefix(release.GetTagName(), "v")
	if latestVersion == holodOverlay.Version || holodOverlay.Version == "0.0.0" {
		return false
	}

	fmt.Printf(color.HiCyanString("新しいバージョンがリリースされています (New version released):v%s -> v%s\n"), holodOverlay.Version, latestVersion)
	fmt.Printf(color.HiCyanString("ダウンロード (Download Here) -> %s\n"), release.GetHTMLURL())
	return true
}

func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}

	return nil
}

func checkLocale() (string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-WinSystemLocale | Select-Object -ExpandProperty Name")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func checkUILang() (string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-UICulture | Select-Object -ExpandProperty Name")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func setLocale(locale string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Not on Windows.")
	}

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", `$process = Start-Process -FilePath 'powershell.exe' -Verb RunAs -Wait -PassThru -ArgumentList '-NoProfile -NonInteractive -Command "if ((Get-WinSystemLocale).Name -ne ''`+locale+`'') { Set-WinSystemLocale -SystemLocale `+locale+` }; exit 0"'; exit $process.ExitCode`)
	return cmd.Run()
}

func isAdminPerm(path string) bool {
	created := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return true
		}
		created = true
	}

	testFile := filepath.Join(path, ".test_access")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return true
	}

	// cleanup test file
	_ = os.Remove(testFile)

	if created {
		_ = os.Remove(path)
	}

	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func localizedChartField(value string, useEnglish bool) string {
	const prefix = "##LOCALIZE:"
	if !strings.HasPrefix(value, prefix) {
		return value
	}

	var localized map[string]string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(value, prefix)), &localized); err != nil {
		return value
	}

	language := "ja"
	if useEnglish {
		language = "en"
	}
	if translated := localized[language]; translated != "" {
		return translated
	}
	return value
}

func localizedChartDescriptionField(description string, field string, useEnglish bool) string {
	marker := "#" + field + ":#SEPARATOR_COLON:"
	for line := range strings.SplitSeq(description, "\n") {
		value, found := strings.CutPrefix(line, marker)
		if !found {
			continue
		}
		return localizedChartField(value, useEnglish)
	}
	return "-"
}

//go:embed assets/holodOverlay-logo-mark.png
var markImg []byte

func origMain(isOptionSpecified bool) {
	Title()

	var skipAviutlInstall bool
	flag.BoolVar(&skipAviutlInstall, "skip-obj-install", false, "AviUtlオブジェクトのインストールをスキップします。(Skip installing AviUtl objects.)")

	var skipAviutlScriptInstall bool
	flag.BoolVar(&skipAviutlScriptInstall, "skip-script-install", false, "AviUtlスクリプトのインストールをスキップします。(Skip installing AviUtl scripts.)")

	var noExplorerAutoOpen bool
	flag.BoolVar(&noExplorerAutoOpen, "no-explorer-auto-open", false, "出力先ディレクトリを自動で開くのを無効にします。(Disable auto-opening output directory in Explorer.)")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "./dist/_chartId_", "出力先ディレクトリを指定します。_chartId_ は譜面IDに置き換えられます。\nEnter the output path. _chartId_ will be replaced with the chart ID.")

	var chartInstance string
	flag.StringVar(&chartInstance, "instance", "", "サーバーインスタンス（またはソースURL）を指定します。\nSpecify the server instance (or source URL).")

	var unitScore float64
	flag.Float64Var(&unitScore, "unit-score", 1000000, "ユニットスコアを指定します。(Specify the team's unit score.)")

	var enUI bool
	flag.BoolVar(&enUI, "en-ui", false, "英語UIを使う。(Use English UI.)")

	var ignoreDamage bool
	flag.BoolVar(&ignoreDamage, "ignore-dmg", true, "ダメージノーツを無視します。(Ignore damage notes.)")

	var difficulty string
	flag.StringVar(&difficulty, "difficulty", "", "難易度を指定します。\nSpecify the difficulty.\n[EASY, NORMAL, HARD, EXPERT, CREATE]")

	flag.Usage = func() {
		fmt.Println("Usage: holodOverlay [オプション (Options)] [譜面ID (Chart ID)]")
		flag.PrintDefaults()
	}

	flag.Parse()

	var listing = false

	fmt.Printf("- 前提条件を確認中 (Checking prerequisites)... ")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}
	if isAdminPerm(cwd) {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: ディレクトリには管理者権限が必要です。holodOverlayを「C:\\」または別の場所に移動してください。\nYour directory requires administrative permissions. Please move holodOverlay to \"C:\\\" or somewhere else.\n\n出力先ディレクトリ (Output path): %s", cwd)))
		return
	}
	if !isASCII(cwd) {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: ディレクトリに非ASCII文字が含まれています。holodOverlayを「C:\\」または別の場所に移動してください。\nYour directory contains non-ASCII characters. Please move holodOverlay to \"C:\\\" or somewhere else.\n\n出力先ディレクトリ (Output path): %s", cwd)))
		return
	}

	locale, err := checkLocale()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if locale != "ja-JP" {
<<<<<<< HEAD
		fmt.Println(color.CyanString("\nINFO: お使いのシステムロケールが「日本語（日本）」に設定されていません。変更しますか？[Y/N]\nYour system locale is not set to \"Japanese (Japan)\". Do you want to change it? [Y/N]"))
=======
		fmt.Print(color.CyanString("\nINFO: お使いのシステムロケールが「日本語（日本）」に設定されていません。変更しますか？[Y/N]\nYour system locale is not set to \"Japanese (Japan)\". Do you want to change it? [Y/N]"))
>>>>>>> b643252e043259b856fd7b3bb8718fba4d226818
		before, _ := rawmode.Enable()
		tmpChangeLocaleByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpChangeLocale := string(tmpChangeLocaleByte)
		rawmode.Restore(before)
		if tmpChangeLocale == "Y" || tmpChangeLocale == "y" {
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpChangeLocale))
			if err := setLocale("ja-JP"); err != nil {
				fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
				return
			} else {
				fmt.Println(color.HiYellowString("WARN: AviUtl ExEdit2を使用できるように、システムのロケールを「日本語（日本）」に変更しました。\n変更を有効にするには、PCを再起動する必要があります。\nSystem locale has been changed to \"Japanese (Japan)\" to be able to use AviUtl ExEdit2.\nYou must restart your computer for it to take effect.\n\n- 何かキーを押すと続行します...\n- Press any key to continue..."))
				before, _ := rawmode.Enable()
				bufio.NewReader(os.Stdin).ReadByte()
				rawmode.Restore(before)
			}
		} else {
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpChangeLocale))
		}
	}

	fmt.Println(color.GreenString("OK"))

	if shouldCheckUpdate() {
		if !isOptionSpecified && checkUpdate() {
<<<<<<< HEAD
			fmt.Println("\n最新リリースのページを開きますか？[Y/N]\nDo you want to open the latest release page? [Y/N]")
=======
			fmt.Print("\n最新リリースのページを開きますか？[Y/N]\nDo you want to open the latest release page? [Y/N]")
>>>>>>> b643252e043259b856fd7b3bb8718fba4d226818
			before, _ := rawmode.Enable()
			tmpOpenBrowserByte, _ := bufio.NewReader(os.Stdin).ReadByte()
			tmpOpenBrowser := string(tmpOpenBrowserByte)
			rawmode.Restore(before)
			if tmpOpenBrowser == "Y" || tmpOpenBrowser == "y" {
				fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpOpenBrowser))
				if err := openBrowser("https://github.com/TootieJin/holodOverlay/releases/latest"); err != nil {
					fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
					return
				}
			} else {
				fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpOpenBrowser))
			}
		}
	}

	uiLang, err := checkUILang()
	var font = []string{"LINE Seed Sans App ExtraBold", "TT-Ro GSan Serif StdN B"}
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if uiLang == "ja-JP" {
		font = []string{"LINE Seed Sans App ExtraBold", "TT-RoG2サンセリフStdN-B"}
	}

	var aviutlPath, aviutlProcess, aviutlName string
	aviutlProcess = "aviutl2.exe"
	aviutlName = "AviUtl ExEdit2"
	aviutlPath, _ = filepath.Abs("C:\\ProgramData\\aviutl2")

	var successInstall = false
	if !skipAviutlInstall {
		success := holodOverlay.TryInstallObject(aviutlPath, font)
		if success {
			fmt.Println(color.GreenString(aviutlName + "オブジェクトのインストールに成功しました。(" + aviutlName + " object successfully installed.)"))
			successInstall = true
		}
	}
	if !skipAviutlScriptInstall {
		success := holodOverlay.TryInstallAviUtlDependencies(aviutlPath, aviutlProcess)
		if success {
			fmt.Println(color.GreenString(aviutlName + "依存関係スクリプトのインストールに成功しました。(" + aviutlName + " dependency scripts successfully installed.)"))
			successInstall = true
		}
	}
	if successInstall {
		fmt.Println(color.HiYellowString("変更を適用するには、" + aviutlName + "を再起動してください。(Please restart " + aviutlName + " to apply changes.)\n"))
	}

	var chartId string
	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		fmt.Printf("譜面ID (Chart ID): %s\n", color.GreenString(chartId))
	} else {
		fmt.Print("譜面IDを接頭辞込みで入力して下さい。（リストについては、GitHubのREADMEをご確認ください。）\nEnter the chart ID including the prefix. (Check README in GitHub for the list.)\n> ")
		fmt.Scanln(&chartId)
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
	}

	// Instance section
	if chartInstance == "" && strings.HasPrefix(chartId, "chcy-") {
		fmt.Printf("\nChart Cyanvasインスタンスを選択してください。(Please choose Chart Cyanvas instance.)\n\n[インスタンス一覧 - List of instance(s)]\n'0': アーカイブ/Archive - cc.milkbun.org\n'1': 分岐サーバー/Offshoot server - chart-cyanvas.com\n> ")
		var chartInput string
		fmt.Scanln(&chartInput)

		chartInstance = chartInput
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartInput))

	} else if chartInstance == "" && strings.HasPrefix(chartId, "custom-") {
		fmt.Printf("\nサーバーのソースURLを入力してください。(Enter the server source URL.)\n> ")
		var chartInput string
		fmt.Scanln(&chartInput)
		chartInput = strings.TrimPrefix(chartInput, "http://")
		chartInput = strings.TrimPrefix(chartInput, "https://")

		chartInstance = chartInput
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartInput))
	}

	var chartSource holodOverlay.Source
	if strings.HasPrefix(chartId, "sync") {
		chartSource, err = holodOverlay.DetectLocalChartSource()
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
		if strings.Contains(chartId, "-") {
			parts := strings.SplitN(chartId, "-", 2)
			if len(parts) == 2 {
				chartId = parts[1]
			}
		} else {
			fmt.Print("ローカルサーバーの譜面を入力してください。(Enter chart ID for the local server.)\n> ")
			fmt.Scanln(&chartId)
		}
	} else {
		chartSource, err = holodOverlay.DetectChartSource(chartId, chartInstance)
		if err != nil {
			fmt.Println(color.RedString("FAIL: 譜面が見つかりません。接頭辞も込め、正しい譜面IDを入力して下さい。\nChart not found. Please enter the correct chart ID including the prefix."))
			return
		}

		switch chartSource.Status {
		case 1:
			fmt.Printf(color.RedString("FAIL: %sは対応されなくなりました。ご利用ありがとうございました。\n%s is no longer supported. Thank you for using the service.\n"), chartSource.Name, chartSource.Name)
			return
		case 2:
			fmt.Printf(color.HiYellowString("WARN: %sは現在開発中であり、正常に動作しない可能性があります。\n%s is currently in development and may not work.\n"), chartSource.Name, chartSource.Name)
		}
	}

	if !isOptionSpecified {
		fmt.Print("\n英語UIを使う？[Y/N]\nUse English UI? [Y/N]\n> ")
		before, _ := rawmode.Enable()
		tmpEnableENByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpEnableEN := string(tmpEnableENByte)
		rawmode.Restore(before)
		if tmpEnableEN == "Y" || tmpEnableEN == "y" {
			enUI = true
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableEN))
			fmt.Println(color.GreenString("TOGGLE: ON"))
		} else {
			enUI = false
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpEnableEN))
			fmt.Println(color.RedString("TOGGLE: OFF"))
		}
	}

	fmt.Printf("- 譜面を取得中 (Getting chart): %s%s%s ", RgbColorEscape(chartSource.Color), chartSource.Name, ResetEscape())

	var chart sonolus.LevelInfo
	var description string
	prefixTrim := checkSubstrings([]string{chartId}, "custom-")
	chart, description, err = holodOverlay.FetchChart(chartSource, strings.TrimPrefix(chartId, prefixTrim))

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	source, err := holodOverlay.Listing("4834734941414141414141414177584255524B4145425146304231357A7969616C74423347304443544E46775453322F63784C77394A556F356734524D394A776F34666D6130456F454C3765744E654B484C5A63614A4B4850767A436939576E4D737A73316179636B5534474F55397371646D586E43326A58514966667052774B57396341414141", chart.Source)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if source {
		listing = true
	}
	list, err := holodOverlay.Listing("4834734941414141414141414178584D32773241494177463049327344314C4645667832675A59416B696759755562484E3534427A6761636453614B71614B4A43647574642F57584B786B2B6F33486C6F4C55554A4C2B6B54494D3474734C61427862726A50672B54454537566A5857634476534A512B7035503350384F49446F7A33526A3130414141413D", chart.Author)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if list {
		listing = true
	}

	var chartInfo struct {
		Title    string
		Artists  string
		Lyrics   string
		Composer string
		Arranger string
		Author   string
	}
	if !strings.HasPrefix(description, "#") {
		chartInfo.Title = chart.Title
		chartInfo.Artists = "-"
		chartInfo.Lyrics = "-"
		chartInfo.Composer = chart.Artists
		chartInfo.Arranger = "-"
		chartInfo.Author = chart.Author
	} else {
		chartInfo.Title = localizedChartField(chart.Title, enUI)
		chartInfo.Artists = localizedChartField(chart.Artists, enUI)
		chartInfo.Lyrics = localizedChartDescriptionField(description, "LYRICIST", enUI)
		chartInfo.Composer = localizedChartDescriptionField(description, "COMPOSER", enUI)
		chartInfo.Arranger = localizedChartDescriptionField(description, "ARRANGER", enUI)
		chartInfo.Author = localizedChartField(chart.Author, enUI)
	}

	fmt.Println(color.GreenString("OK"))

	fmt.Printf("  %s / %s - %s (Lv. %s)\n",
		color.CyanString(chartInfo.Title),
		color.CyanString(chartInfo.Composer),
		color.CyanString(chartInfo.Author),
		color.MagentaString(strconv.Itoa(chart.Rating)),
	)

	fmt.Printf("- exeのパスを取得中 (Getting executable path)... ")
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	formattedOutDir := filepath.Join(cwd, strings.ReplaceAll(outDir, "_chartId_", chartId))
	resultDir := filepath.Dir(formattedOutDir) + "\\" + chartId

	fmt.Println(color.GreenString("OK"))
	fmt.Printf("- 出力先ディレクトリ (Output path): %s\n", color.CyanString(resultDir))

	fmt.Print("- ジャケットをダウンロード中 (Downloading jacket)... ")
	err = holodOverlay.DownloadJacket(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	fmt.Print("- 譜面を解析中 (Analyzing chart)... ")

	levelData, err := holodOverlay.FetchLevelData(chartSource, chart)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	if !isOptionSpecified {
		fmt.Print("\nユニットスコアを指定してください。 (Input your unit score.)\n\n- 小数と科学的記数法が使える (Accepts decimals & scientific notation)\n- おすすめ (Recommended): 1000000 - 2000000\n- 例 (Example): 1234567; 1e+20; -500000\n> ")
		var tmpUnitScore string
		fmt.Scanln(&tmpUnitScore)
		if tmpUnitScore == "" {
			tmpUnitScore = "1000000"
		}
		unitScore, err = strconv.ParseFloat(tmpUnitScore, 64)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpUnitScore))
	}

	if !isOptionSpecified {
		if damageNote := holodOverlay.CheckDamageNote(levelData); damageNote {
			fmt.Print("\nこの譜面にはダメージノーツが含まれています。ダメージノーツを無視しますか？[Y/N]\nYour chart contains damage notes. Do you want ignore damage notes? [Y/N]\n> ")
			before, _ := rawmode.Enable()
			tmpIgnoreDmgByte, _ := bufio.NewReader(os.Stdin).ReadByte()
			tmpIgnoreDmg := string(tmpIgnoreDmgByte)
			rawmode.Restore(before)
			if tmpIgnoreDmg == "Y" || tmpIgnoreDmg == "y" {
				ignoreDamage = true
				fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpIgnoreDmg))
				fmt.Println(color.GreenString("TOGGLE: ON"))
			} else {
				ignoreDamage = false
				fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpIgnoreDmg))
				fmt.Println(color.RedString("TOGGLE: OFF"))
			}
		}
	}

	fmt.Print("- スコアを計算中 (Calculating score)... ")
	scoreData := holodOverlay.CalculateScore(chart, levelData, unitScore, ignoreDamage, listing)

	fmt.Println(color.GreenString("OK"))

	executableDir := filepath.Dir(executablePath)
	assets := filepath.Join(executableDir, "assets")

	if len(markImg) > 0 {
		if err := os.WriteFile(filepath.Join(assets, "holodOverlay-logo-mark.png"), markImg, 0644); err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
	}

	fmt.Print("- YAGOOファイルを生成中 (Generating YAGOO file)... ")

	err = holodOverlay.WriteYagooFile(scoreData, assets, filepath.Join(formattedOutDir, "data.yagoo"), sonolus.LevelInfo{Rating: chart.Rating}, levelData, listing)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	fmt.Print("- aup2プロジェクトファイルを生成中 (Generating aup2 project file)... ")

	difficultyStrings := []string{"EASY", "NORMAL", "HARD", "EXPERT", "CREATE"}
	difficultyCol := []int{0x7ad89c, 0xf4be67, 0xef98b6, 0x9a92ff, 0x83daef}

	if difficulty == "" {
		for i := range chart.Tags {
			tags := checkSubstrings([]string{strings.ToUpper(strings.TrimPrefix(chart.Tags[i].Title, "#"))}, difficultyStrings...)
			if tags != "" {
				difficulty = tags
				break
			}
		}
	}

	if difficulty == "" {
		if title := checkSubstrings(strings.Fields(strings.ToUpper(chart.Title)), difficultyStrings...); title != "" {
			difficulty = title
		} else {
			if !isOptionSpecified {
				fmt.Print("\n難易度を指定してください。(Input the difficulty.)\n[EASY, NORMAL, HARD, EXPERT, CREATE]\n> ")
				var tmpDifficulty string
				fmt.Scanln(&tmpDifficulty)

				difficulty = checkSubstrings(strings.Fields(strings.ToUpper(tmpDifficulty)), difficultyStrings...)
				if difficulty == "" {
					difficulty = "CREATE"
				}
				fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(difficulty))
			}
		}
	}

	if stringTest := checkSubstrings(strings.Fields(strings.ToUpper(difficulty)), difficultyStrings...); stringTest == "" {
		difficulty = "CREATE"
	}

	composerAndVocals := []string{chartInfo.Artists, "-"}
	if separateAttempt := strings.Split(chartInfo.Artists, " / "); chartSource.Id == "chart_cyanvas" && len(separateAttempt) <= 2 {
		composerAndVocals = separateAttempt
		chartInfo.Composer = composerAndVocals[0]
		chartInfo.Artists = composerAndVocals[1]
	}

	charter := []string{chartInfo.Author, "-"}
	if charterTag := strings.Split(chartInfo.Author, "#"); len(charterTag) <= 2 {
		charter = charterTag
	}

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

	caser := cases.Title(language.English)
	var textAndEffects = []string{
		strconv.FormatInt(int64(difficultyCol[matchedStringIndex(difficultyStrings, difficulty)]), 16),
		caser.String(difficulty),
		"<s-3.5>譜面制作<s><gw+14> <gw><gw+0.6>" + charter[0], strconv.Itoa(5),
		chartInfo.Artists,
		"<s-3.5>作詞<s><gw+14> <gw><gw+0.6>" + chartInfo.Lyrics, strconv.Itoa(3),
		"<s-3.5>作曲<s><gw+14> <gw><gw+0.6>" + chartInfo.Composer, strconv.Itoa(3),
		"<s-3.5>編曲<s><gw+14> <gw><gw+0.6>" + chartInfo.Arranger, strconv.Itoa(3),
		"【追加情報】",
		chartInfo.Title,
	}

	if enUI {
		textAndEffects = []string{
			strconv.FormatInt(int64(difficultyCol[matchedStringIndex(difficultyStrings, difficulty)]), 16),
			caser.String(difficulty),
			"<s-3.5>Chart Design<s><gw+14> <gw><gw+0.6>" + charter[0], strconv.Itoa(13),
			chartInfo.Artists,
			"<s-3.5>Lyrics<s><gw+14> <gw><gw+0.6>" + chartInfo.Lyrics, strconv.Itoa(7),
			"<s-3.5>Composer<s><gw+14> <gw><gw+0.6>" + chartInfo.Composer, strconv.Itoa(9),
			"<s-3.5>Arranger<s><gw+14> <gw><gw+0.6>" + chartInfo.Arranger, strconv.Itoa(9),
			"【Additional Info】",
			chartInfo.Title,
		}
	}

	err = holodOverlay.WriteAup2Files(assets, formattedOutDir, textAndEffects, font, chartId)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	message := "\n全ての処理が完了しました！READMEの規約を確認した上で、aup2プロジェクトファイルを開いてください。\nExecution complete! Please open the aup2 project file after reviewing the README Terms of Use."
	fmt.Println(color.GreenString(message))

	if !isOptionSpecified || !noExplorerAutoOpen {
		cmd := exec.Command(`explorer`, `/select,`, resultDir)
		cmd.Run()

		time.Sleep(2000 * time.Millisecond)
	}
}

func main() {
	isOptionSpecified := len(os.Args) > 1
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	origMain(isOptionSpecified)

	if !isOptionSpecified {
		fmt.Print(color.CyanString("\n- 何かキーを押すと終了します...\n- Press any key to exit..."))
		before, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
		rawmode.Restore(before)
	}
}
