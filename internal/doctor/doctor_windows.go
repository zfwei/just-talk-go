//go:build windows

package doctor

import (
	"strings"

	"github.com/c/just-talk-go/config"
)

func runPlatform(cfg *config.Config, backend string) Report {
	if backend == "" {
		backend = "windows"
	}
	report := Report{
		Platform: "windows",
		Backend:  backend,
		Info: []string{
			"系统平台：Windows",
		},
	}

	// 1. Audio Recording Backend Check (Any of ffmpeg, sox)
	recCheck := commandAnyCheck(
		"录音工具安装",
		Required,
		[]string{"ffmpeg", "sox"},
		"请安装 ffmpeg 并将其路径添加到系统 PATH 环境变量。您可以在 PowerShell 中运行：\n  winget install Gnu.FFmpeg",
	)
	if recCheck.OK && strings.Contains(recCheck.Detail, "ffmpeg") {
		recCheck.Notes = append(recCheck.Notes,
			"提示：若录制无声，请在 CMD/PowerShell 运行以下命令获取您电脑的所有麦克风名称：",
			"  ffmpeg -list_devices true -f dshow -i dummy",
			"  然后在 config.toml 中配置：[voice] device = \"您的麦克风设备名称\"",
		)
	}
	report.Checks = append(report.Checks, recCheck)

	// 2. Global Hotkey Interface
	report.Checks = append(report.Checks, Check{
		Name:     "快捷键接口",
		OK:       true,
		Severity: Required,
		Detail:   "使用 WH_KEYBOARD_LL Windows 键盘钩子监听全局快捷键",
	})

	return report
}
