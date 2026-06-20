//go:build windows

package doctor

import (
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
