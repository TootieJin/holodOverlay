package main

import (
	"fmt"
	"strings"

	"github.com/TootieJin/holodOverlay/pkg/holodOverlay"
	"github.com/lithammer/dedent"
)

func Title() {
	fmt.Printf(
		strings.TrimSpace(dedent.Dedent(`
    %s-- %sho%slo%sdO%sve%srl%say %s---------------------------------------------------------------------%s
    %sホロドリ風動画作成補助ツール / A "hololive Dreams"-style video creation tool%s
        Version: %s%s%s

        Developed by %sTootieJin (https://tootiejin.com)%s
         -> https://github.com/TootieJin/holodOverlay%s
    %s-------------------------------------------------------------------------------------%s
    `))+"\n",
		RgbColorEscape(0xff5a91), RgbColorEscape(0x17ccff), RgbColorEscape(0xff779f), RgbColorEscape(0x39e870), RgbColorEscape(0x069bff), RgbColorEscape(0xffaf50), RgbColorEscape(0xd07ef6), RgbColorEscape(0xff5a91), ResetEscape(),
		RgbColorEscape(0x00afc7), ResetEscape(),
		RgbColorEscape(0x0f6ea3), holodOverlay.Version, ResetEscape(),
		RgbColorEscape(0x48b0d5), ResetEscape(),
		ResetEscape(),
		RgbColorEscape(0xff5a91), ResetEscape(),
	)

}

func RgbColorEscape(rgb int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", (rgb>>16)&0xff, (rgb>>8)&0xff, rgb&0xff)
}

func AnsiColorEscape(color int) string {
	return fmt.Sprintf("\033[38;5;%dm", color)
}

func ResetEscape() string {
	return "\033[0m"
}
