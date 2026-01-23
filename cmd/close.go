package cmd

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"trading-journal-cli/internal/models"
	"trading-journal-cli/internal/operations"
)

var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "平仓记录",
	Long:  `选择未平仓的仓位并记录平仓信息`,
	RunE:  runClose,
}

func init() {
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	printTitle("📉 平仓记录")

	// 获取所有未平仓位
	openPositions, err := ops.GetOpenPositions()
	if err != nil {
		printError(fmt.Sprintf("无法读取未平仓位: %v", err))
		return err
	}

	if len(openPositions) == 0 {
		printWarning("暂无未平仓位")
		return nil
	}

	// 显示未平仓位列表
	printInfo(fmt.Sprintf("找到 %d 个未平仓位", len(openPositions)))
	fmt.Println()

	options := make([]string, len(openPositions))
	for i, pos := range openPositions {
		options[i] = fmt.Sprintf("[%s] %s (%s) @ %.4f x %.4f",
			pos.PositionID[:13]+"...", pos.Symbol, pos.Direction, pos.OpenPrice, pos.Quantity)
	}

	// 选择仓位
	var selectedIndex int
	selectPrompt := &survey.Select{
		Message: "选择要平仓的仓位:",
		Options: options,
	}
	if err := survey.AskOne(selectPrompt, &selectedIndex); err != nil {
		return err
	}

	selectedPos := openPositions[selectedIndex]
	fmt.Println()
	printDivider()
	printHighlightField("仓位ID", selectedPos.PositionID)
	printField("品种", fmt.Sprintf("%s (%s)", selectedPos.Symbol, selectedPos.MarketType))
	printField("方向", selectedPos.Direction)
	printField("开仓价格", fmt.Sprintf("%.4f", selectedPos.OpenPrice))
	printField("数量", fmt.Sprintf("%.4f", selectedPos.Quantity))
	printDivider()
	fmt.Println()

	var params operations.CloseParams

	// 平仓价格
	var closePriceStr string
	closePricePrompt := &survey.Input{
		Message: "平仓价格:",
	}
	if err := survey.AskOne(closePricePrompt, &closePriceStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(closePriceStr, "%f", &params.ClosePrice); err != nil {
		return fmt.Errorf("无效的价格格式: %w", err)
	}

	// 平仓数量
	var closeQuantityStr string
	closeQuantityPrompt := &survey.Input{
		Message: fmt.Sprintf("平仓数量 (最大 %.4f):", selectedPos.Quantity),
		Default: fmt.Sprintf("%.4f", selectedPos.Quantity),
	}
	if err := survey.AskOne(closeQuantityPrompt, &closeQuantityStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(closeQuantityStr, "%f", &params.CloseQuantity); err != nil {
		return fmt.Errorf("无效的数量格式: %w", err)
	}

	// 询问是否手动输入盈亏
	var useManualPnL bool
	manualPnLPrompt := &survey.Confirm{
		Message: "是否手动输入盈亏金额?",
		Default: false,
	}
	if err := survey.AskOne(manualPnLPrompt, &useManualPnL); err != nil {
		return err
	}

	// 如果选择手动输入盈亏
	if useManualPnL {
		var pnlStr string
		pnlPrompt := &survey.Input{
			Message: "盈亏金额 (正数为盈利，负数为亏损):",
		}
		if err := survey.AskOne(pnlPrompt, &pnlStr, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		var pnl float64
		if _, err := fmt.Sscanf(pnlStr, "%f", &pnl); err != nil {
			return fmt.Errorf("无效的盈亏格式: %w", err)
		}
		params.ManualPnL = &pnl
	}

	// 平仓原因
	var closeReasonStr string
	closeReasonPrompt := &survey.Select{
		Message: "平仓原因:",
		Options: []string{"stop_loss", "take_profit", "manual"},
	}
	if err := survey.AskOne(closeReasonPrompt, &closeReasonStr); err != nil {
		return err
	}
	params.CloseReason = models.CloseReason(closeReasonStr)

	// 平仓备注（可选）
	closeNotePrompt := &survey.Input{
		Message: "平仓备注 (可选):",
	}
	survey.AskOne(closeNotePrompt, &params.CloseNote)

	// 平仓时间（可选）
	var useCurrentTime bool
	timePrompt := &survey.Confirm{
		Message: "使用当前时间?",
		Default: true,
	}
	if err := survey.AskOne(timePrompt, &useCurrentTime); err != nil {
		return err
	}

	if !useCurrentTime {
		var timeStr string
		customTimePrompt := &survey.Input{
			Message: "平仓时间 (格式: 2006-01-02 15:04:05):",
		}
		if err := survey.AskOne(customTimePrompt, &timeStr); err != nil {
			return err
		}
		if timeStr != "" {
			t, err := time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				return fmt.Errorf("无效的时间格式: %w", err)
			}
			params.CloseTime = &t
		}
	}

	// 执行平仓操作
	pos, err := ops.ClosePosition(selectedPos.PositionID, params)
	if err != nil {
		printError(fmt.Sprintf("平仓失败: %v", err))
		return err
	}

	// 显示成功信息
	fmt.Println()
	printSuccess("仓位已平仓")
	printHighlightField("仓位ID", pos.PositionID)
	printDivider()
	printField("平仓价格", fmt.Sprintf("%.4f", *pos.ClosePrice))
	printField("平仓数量", fmt.Sprintf("%.4f", *pos.CloseQuantity))

	// 盈亏（使用颜色）
	pnlColor := color.New(color.FgRed)
	pnlSign := ""
	if *pos.RealizedPnL > 0 {
		pnlColor = color.New(color.FgGreen, color.Bold)
		pnlSign = "+"
	}
	fmt.Print("  ")
	colorMuted := color.New(color.FgHiBlack)
	colorMuted.Printf("%-15s ", "盈亏:")
	pnlColor.Printf("%s%.2f (%s%.2f%%)\n",
		pnlSign, *pos.RealizedPnL, pnlSign, *pos.PnLPercentage)

	printField("持仓时长", *pos.HoldingDuration)
	if pos.CloseNote != "" {
		printField("备注", pos.CloseNote)
	}
	fmt.Println()

	return nil
}
