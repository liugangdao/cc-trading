package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"trading-journal-cli/internal/models"
	"trading-journal-cli/internal/operations"
)

var (
	listStatus      string
	listSymbol      string
	listMarketType  string
	listAccountName string
	listFromDate    string
	listToDate      string
	listFormat      string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "查询交易记录",
	Long:  `查看和筛选交易记录`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "all", "筛选状态 (open, closed, all)")
	listCmd.Flags().StringVar(&listSymbol, "symbol", "", "筛选交易品种")
	listCmd.Flags().StringVar(&listMarketType, "market", "", "筛选市场类型")
	listCmd.Flags().StringVar(&listAccountName, "account", "", "筛选账户")
	listCmd.Flags().StringVar(&listFromDate, "from", "", "起始日期 (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listToDate, "to", "", "结束日期 (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "输出格式 (table, json)")

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// 解析筛选参数
	filter := operations.FilterParams{
		Status:      listStatus,
		Symbol:      listSymbol,
		MarketType:  listMarketType,
		AccountName: listAccountName,
	}

	if listFromDate != "" {
		t, err := time.Parse("2006-01-02", listFromDate)
		if err != nil {
			return fmt.Errorf("无效的起始日期格式: %w", err)
		}
		filter.FromDate = t
	}

	if listToDate != "" {
		t, err := time.Parse("2006-01-02", listToDate)
		if err != nil {
			return fmt.Errorf("无效的结束日期格式: %w", err)
		}
		filter.ToDate = t
	}

	// 查询仓位
	positions, err := ops.ListPositions(filter)
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}

	if len(positions) == 0 {
		fmt.Println("未找到匹配的记录")
		return nil
	}

	// 根据格式输出
	if listFormat == "json" {
		return outputJSON(positions)
	}
	return outputTable(positions)
}

func outputJSON(positions []*models.Position) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(positions)
}

// padRight 使用runewidth正确计算字符宽度并右填充
func padRight(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func outputTable(positions []*models.Position) error {
	printTitle("📊 交易记录")

	// 统计信息
	var openCount, closedCount int
	var totalPnL, totalPnLPercentage float64
	for _, pos := range positions {
		if pos.Status == models.StatusOpen {
			openCount++
		} else {
			closedCount++
			if pos.RealizedPnL != nil {
				totalPnL += *pos.RealizedPnL
			}
			if pos.PnLPercentage != nil {
				totalPnLPercentage += *pos.PnLPercentage
			}
		}
	}

	// 显示统计
	printInfo(fmt.Sprintf("总计: %d 条记录 | 持仓: %d | 已平仓: %d",
		len(positions), openCount, closedCount))
	if closedCount > 0 {
		avgPnL := totalPnL / float64(closedCount)
		avgPnLPct := totalPnLPercentage / float64(closedCount)
		pnlSign := ""
		if totalPnL > 0 {
			pnlSign = "+"
		}
		printInfo(fmt.Sprintf("总盈亏: %s%.2f | 平均: %.2f (%.2f%%)",
			pnlSign, totalPnL, avgPnL, avgPnLPct))
	}
	fmt.Println()
	printDivider()
	fmt.Println()

	// 列宽定义
	const (
		colPosID    = 20
		colSymbol   = 12
		colDir      = 8
		colPrice    = 12
		colQty      = 10
		colStatus   = 10
		colPnL      = 22
	)

	// 颜色定义
	colorTitle := color.New(color.FgCyan, color.Bold)
	colorMuted := color.New(color.FgHiBlack)
	colorGreen := color.New(color.FgGreen)
	colorRed := color.New(color.FgRed)
	colorYellow := color.New(color.FgYellow)
	colorBlue := color.New(color.FgBlue)
	colorGreenBold := color.New(color.FgGreen, color.Bold)

	// 表头
	fmt.Print("  ")
	colorTitle.Print(padRight("仓位ID", colPosID))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("品种", colSymbol))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("方向", colDir))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("开仓价", colPrice))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("数量", colQty))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("状态", colStatus))
	colorMuted.Print(" │ ")
	colorTitle.Print(padRight("盈亏", colPnL))
	fmt.Println()

	fmt.Print("  ")
	colorMuted.Println(strings.Repeat("─", colPosID+colSymbol+colDir+colPrice+colQty+colStatus+colPnL+18))

	// 数据行
	for _, pos := range positions {
		fmt.Print("  ")

		// 仓位ID（缩短显示，绿色高亮）
		posID := pos.PositionID
		if runewidth.StringWidth(posID) > colPosID {
			// 截断但保持正确的宽度
			for runewidth.StringWidth(posID) > colPosID-3 {
				posID = posID[:len(posID)-1]
			}
			posID = posID + "..."
		}
		colorGreen.Print(padRight(posID, colPosID))
		colorMuted.Print(" │ ")

		// 品种
		fmt.Print(padRight(pos.Symbol, colSymbol))
		colorMuted.Print(" │ ")

		// 方向（使用颜色）
		directionText := "做多"
		if pos.Direction == "short" {
			directionText = "做空"
			colorRed.Print(padRight(directionText, colDir))
		} else {
			colorGreen.Print(padRight(directionText, colDir))
		}
		colorMuted.Print(" │ ")

		// 开仓价格
		openPriceStr := fmt.Sprintf("%.4f", pos.OpenPrice)
		fmt.Print(padRight(openPriceStr, colPrice))
		colorMuted.Print(" │ ")

		// 数量 - 根据状态显示不同的数量
		var quantityStr string
		if pos.Status == models.StatusClosed && pos.CloseQuantity != nil {
			quantityStr = fmt.Sprintf("%.2f", *pos.CloseQuantity)
		} else {
			quantityStr = fmt.Sprintf("%.2f", pos.Quantity)
		}
		fmt.Print(padRight(quantityStr, colQty))
		colorMuted.Print(" │ ")

		// 状态（使用颜色）
		statusText := "持仓中"
		if pos.Status == models.StatusOpen {
			colorYellow.Print(padRight(statusText, colStatus))
		} else {
			statusText = "已平仓"
			colorBlue.Print(padRight(statusText, colStatus))
		}
		colorMuted.Print(" │ ")

		// 盈亏
		if pos.Status == models.StatusClosed && pos.RealizedPnL != nil && pos.PnLPercentage != nil {
			pnlSign := ""
			if *pos.RealizedPnL > 0 {
				pnlSign = "+"
			}
			pnlStr := fmt.Sprintf("%s%.2f (%s%.2f%%)",
				pnlSign, *pos.RealizedPnL, pnlSign, *pos.PnLPercentage)

			if *pos.RealizedPnL > 0 {
				colorGreenBold.Print(padRight(pnlStr, colPnL))
			} else {
				colorRed.Print(padRight(pnlStr, colPnL))
			}
		} else {
			colorMuted.Print(padRight("-", colPnL))
		}

		fmt.Println()
	}

	fmt.Println()
	printDivider()
	printHint("使用 --format json 可查看完整详细信息")
	fmt.Println()

	return nil
}
