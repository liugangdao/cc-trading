package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// UI颜色定义
var (
	colorTitle   = color.New(color.FgCyan, color.Bold)
	colorSuccess = color.New(color.FgGreen, color.Bold)
	colorWarning = color.New(color.FgYellow, color.Bold)
	colorError   = color.New(color.FgRed, color.Bold)
	colorInfo    = color.New(color.FgBlue)
	colorMuted   = color.New(color.FgHiBlack)
	colorHighlight = color.New(color.FgGreen)
	colorValue   = color.New(color.FgWhite, color.Bold)
)

// 打印标题
func printTitle(title string) {
	fmt.Println()
	colorTitle.Println("╔═══════════════════════════════════════════════════════════════")
	colorTitle.Printf("║  %s\n", title)
	colorTitle.Println("╚═══════════════════════════════════════════════════════════════")
	fmt.Println()
}

// 打印分割线
func printDivider() {
	colorMuted.Println(strings.Repeat("─", 65))
}

// 打印成功消息
func printSuccess(msg string) {
	fmt.Print("  ")
	colorSuccess.Print("✓ ")
	fmt.Println(msg)
}

// 打印信息消息
func printInfo(msg string) {
	fmt.Print("  ")
	colorInfo.Print("ℹ ")
	fmt.Println(msg)
}

// 打印警告消息
func printWarning(msg string) {
	fmt.Print("  ")
	colorWarning.Print("⚠ ")
	fmt.Println(msg)
}

// 打印错误消息
func printError(msg string) {
	fmt.Print("  ")
	colorError.Print("✗ ")
	fmt.Println(msg)
}

// 打印字段
func printField(label string, value interface{}) {
	fmt.Print("  ")
	colorMuted.Printf("%-15s ", label+":")
	colorValue.Printf("%v\n", value)
}

// 打印高亮字段
func printHighlightField(label string, value interface{}) {
	fmt.Print("  ")
	colorMuted.Printf("%-15s ", label+":")
	colorHighlight.Printf("%v\n", value)
}

// 打印表格头部
func printTableHeader(headers ...string) {
	fmt.Print("  ")
	for i, header := range headers {
		if i > 0 {
			colorMuted.Print(" │ ")
		}
		colorTitle.Print(header)
	}
	fmt.Println()
	fmt.Print("  ")
	colorMuted.Println(strings.Repeat("─", 65))
}

// 打印提示信息
func printHint(msg string) {
	fmt.Print("  ")
	colorMuted.Printf("💡 %s\n", msg)
}
