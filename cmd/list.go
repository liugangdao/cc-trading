package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"trading-journal-cli/internal/models"
	"trading-journal-cli/internal/operations"
)

var (
	listStatus     string
	listSymbol     string
	listMarketType string
	listFromDate   string
	listToDate     string
	listFormat     string
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
	listCmd.Flags().StringVar(&listFromDate, "from", "", "起始日期 (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listToDate, "to", "", "结束日期 (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "输出格式 (table, json)")

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// 解析筛选参数
	filter := operations.FilterParams{
		Status:     listStatus,
		Symbol:     listSymbol,
		MarketType: listMarketType,
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

func outputTable(positions []*models.Position) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// 表头
	fmt.Fprintln(w, "ID\t品种\t市场\t方向\t开仓价\t数量\t状态\t盈亏")
	fmt.Fprintln(w, "---\t---\t---\t---\t---\t---\t---\t---")

	// 数据行
	for _, pos := range positions {
		pnlStr := "-"
		if pos.Status == models.StatusClosed && pos.RealizedPnL != nil && pos.PnLPercentage != nil {
			pnlSign := ""
			if *pos.RealizedPnL > 0 {
				pnlSign = "+"
			}
			pnlStr = fmt.Sprintf("%s%.2f (%s%.2f%%)",
				pnlSign, *pos.RealizedPnL, pnlSign, *pos.PnLPercentage)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.4f\t%.4f\t%s\t%s\n",
			pos.PositionID[:13]+"...", // 缩短ID显示
			pos.Symbol,
			pos.MarketType,
			pos.Direction,
			pos.OpenPrice,
			pos.Quantity,
			pos.Status,
			pnlStr,
		)
	}

	return nil
}
