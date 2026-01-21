package cmd

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
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
	fmt.Println("📉 平仓记录")
	fmt.Println()

	// 获取所有未平仓位
	openPositions, err := ops.GetOpenPositions()
	if err != nil {
		return fmt.Errorf("无法读取未平仓位: %w", err)
	}

	if len(openPositions) == 0 {
		fmt.Println("暂无未平仓位")
		return nil
	}

	// 显示未平仓位列表
	fmt.Println("未平仓位:")
	options := make([]string, len(openPositions))
	for i, pos := range openPositions {
		options[i] = fmt.Sprintf("[%s] %s (%s) @ %.4f",
			pos.PositionID, pos.Symbol, pos.Direction, pos.OpenPrice)
		fmt.Printf("%d. %s\n", i+1, options[i])
	}
	fmt.Println()

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
	fmt.Printf("\n选中仓位: %s\n", selectedPos.PositionID)
	fmt.Printf("品种: %s, 方向: %s, 开仓价: %.4f, 数量: %.4f\n\n",
		selectedPos.Symbol, selectedPos.Direction, selectedPos.OpenPrice, selectedPos.Quantity)

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
		return fmt.Errorf("平仓失败: %w", err)
	}

	// 显示成功信息
	fmt.Println()
	fmt.Printf("✓ 仓位已平仓: %s\n", pos.PositionID)
	fmt.Printf("  平仓价格: %.4f\n", *pos.ClosePrice)
	fmt.Printf("  平仓数量: %.4f\n", *pos.CloseQuantity)

	pnlSign := ""
	if *pos.RealizedPnL > 0 {
		pnlSign = "+"
	}
	fmt.Printf("  盈亏: %s%.2f (%s%.2f%%)\n",
		pnlSign, *pos.RealizedPnL, pnlSign, *pos.PnLPercentage)
	fmt.Printf("  持仓时长: %s\n", *pos.HoldingDuration)
	if pos.CloseNote != "" {
		fmt.Printf("  备注: %s\n", pos.CloseNote)
	}

	return nil
}
