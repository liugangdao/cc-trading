package cmd

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"trading-journal-cli/internal/models"
	"trading-journal-cli/internal/operations"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "开仓记录",
	Long:  `通过交互式提示记录新的开仓信息`,
	RunE:  runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	printTitle("📈 开仓记录")

	var params operations.OpenParams

	// 选择账户
	am := getAccountManager()
	accounts := am.ListAccounts()

	if len(accounts) == 0 {
		printWarning("未找到账户配置")
		printHint("请先添加账户: trading-cli account add")
		return fmt.Errorf("no accounts configured")
	}

	var selectedAccountIndex int
	accountOptions := make([]string, len(accounts))
	for i, acc := range accounts {
		currency := acc.Currency
		if currency == "" {
			currency = "USD"
		}
		accountOptions[i] = fmt.Sprintf("%s (%.2f %s)", acc.Name, acc.Balance, currency)
	}

	accountPrompt := &survey.Select{
		Message: "选择账户:",
		Options: accountOptions,
	}
	if err := survey.AskOne(accountPrompt, &selectedAccountIndex); err != nil {
		return err
	}

	selectedAccount := accounts[selectedAccountIndex]
	params.AccountName = selectedAccount.Name
	params.AccountBalance = selectedAccount.Balance

	fmt.Println()
	printHighlightField("账户", fmt.Sprintf("%s (%.2f %s)", selectedAccount.Name, selectedAccount.Balance, selectedAccount.Currency))

	// 显示模板信息（如果有）
	if selectedAccount.Template != nil {
		printHint("使用账户模板默认值")
	}
	fmt.Println()
	printDivider()
	fmt.Println()

	// 交易品种
	symbolDefault := ""
	if selectedAccount.Template != nil && selectedAccount.Template.DefaultSymbol != "" {
		symbolDefault = selectedAccount.Template.DefaultSymbol
	}
	symbolPrompt := &survey.Input{
		Message: "交易品种 (如 BTC/USDT):",
		Default: symbolDefault,
	}
	if err := survey.AskOne(symbolPrompt, &params.Symbol, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// 市场类型
	var marketTypeStr string
	marketTypeOptions := []string{"crypto", "forex", "gold", "silver", "futures"}
	marketTypeDefault := 0
	if selectedAccount.Template != nil && selectedAccount.Template.DefaultMarketType != "" {
		// 找到默认值的索引
		for i, opt := range marketTypeOptions {
			if opt == string(selectedAccount.Template.DefaultMarketType) {
				marketTypeDefault = i
				break
			}
		}
	}
	marketTypePrompt := &survey.Select{
		Message: "市场类型:",
		Options: marketTypeOptions,
		Default: marketTypeDefault,
	}
	if err := survey.AskOne(marketTypePrompt, &marketTypeStr); err != nil {
		return err
	}
	params.MarketType = models.MarketType(marketTypeStr)

	// 方向
	var directionStr string
	directionOptions := []string{"long", "short"}
	directionDefault := 0
	if selectedAccount.Template != nil && selectedAccount.Template.DefaultDirection != "" {
		// 找到默认值的索引
		for i, opt := range directionOptions {
			if opt == string(selectedAccount.Template.DefaultDirection) {
				directionDefault = i
				break
			}
		}
	}
	directionPrompt := &survey.Select{
		Message: "方向:",
		Options: directionOptions,
		Default: directionDefault,
	}
	if err := survey.AskOne(directionPrompt, &directionStr); err != nil {
		return err
	}
	params.Direction = models.Direction(directionStr)

	// 开仓价格
	var openPriceStr string
	openPricePrompt := &survey.Input{
		Message: "开仓价格:",
	}
	if err := survey.AskOne(openPricePrompt, &openPriceStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(openPriceStr, "%f", &params.OpenPrice); err != nil {
		return fmt.Errorf("无效的价格格式: %w", err)
	}

	// 仓位大小
	var quantityStr string
	quantityPrompt := &survey.Input{
		Message: "仓位大小:",
	}
	if err := survey.AskOne(quantityPrompt, &quantityStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(quantityStr, "%f", &params.Quantity); err != nil {
		return fmt.Errorf("无效的数量格式: %w", err)
	}

	// 止损价格
	var stopLossStr string
	stopLossPrompt := &survey.Input{
		Message: "止损价格 (必填):",
	}
	if err := survey.AskOne(stopLossPrompt, &stopLossStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(stopLossStr, "%f", &params.StopLoss); err != nil {
		return fmt.Errorf("无效的止损价格格式: %w", err)
	}

	// 止盈价格
	var takeProfitStr string
	takeProfitPrompt := &survey.Input{
		Message: "止盈价格 (必填):",
	}
	if err := survey.AskOne(takeProfitPrompt, &takeProfitStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(takeProfitStr, "%f", &params.TakeProfit); err != nil {
		return fmt.Errorf("无效的止盈价格格式: %w", err)
	}

	// 保证金/成本
	var marginStr string
	marginPrompt := &survey.Input{
		Message: "保证金/成本:",
	}
	if err := survey.AskOne(marginPrompt, &marginStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(marginStr, "%f", &params.Margin); err != nil {
		return fmt.Errorf("无效的保证金格式: %w", err)
	}

	// 交易理由（可选）
	reasonPrompt := &survey.Input{
		Message: "交易理由 (可选):",
	}
	survey.AskOne(reasonPrompt, &params.Reason)

	fmt.Println()
	printDivider()
	fmt.Println()

	// 市场背景判断
	printInfo("市场背景判断（可选）")
	var judgeMarket bool
	judgeMarketPrompt := &survey.Confirm{
		Message: "是否需要判断市场背景?",
		Default: false,
	}
	if err := survey.AskOne(judgeMarketPrompt, &judgeMarket); err != nil {
		return err
	}

	if judgeMarket {
		// 询问市场类型
		var marketContextStr string
		marketContextPrompt := &survey.Select{
			Message: "当前市场背景:",
			Options: []string{"牛市", "熊市", "震荡"},
		}
		if err := survey.AskOne(marketContextPrompt, &marketContextStr); err != nil {
			return err
		}

		// 只有牛市才进行进一步判断
		if marketContextStr == "牛市" {
			params.MarketContext = models.MarketContextBull

			fmt.Println()
			printWarning("牛市末期信号判断（满足2项或以上为牛市末期）")
			fmt.Println()

			// ① 日线是否跌破 EMA20，并反抽失败
			var ema20Broken bool
			ema20Prompt := &survey.Confirm{
				Message: "① 日线是否跌破 EMA20，并反抽失败?",
				Default: false,
			}
			survey.AskOne(ema20Prompt, &ema20Broken)
			params.EMA20Broken = ema20Broken

			// ② 创新高但成交量明显低于前高
			var volumeDecrease bool
			volumePrompt := &survey.Confirm{
				Message: "② 创新高但成交量明显低于前高?",
				Default: false,
			}
			survey.AskOne(volumePrompt, &volumeDecrease)
			params.VolumeDecrease = volumeDecrease

			// ③ 连续两次回调都打穿前低
			var consecutiveLowBreak bool
			lowBreakPrompt := &survey.Confirm{
				Message: "③ 连续两次回调都打穿前低?",
				Default: false,
			}
			survey.AskOne(lowBreakPrompt, &consecutiveLowBreak)
			params.ConsecutiveLowBreak = consecutiveLowBreak

			// 计算满足的条件数量
			signalCount := 0
			var signals []string
			if ema20Broken {
				signalCount++
				signals = append(signals, "日线跌破EMA20并反抽失败")
			}
			if volumeDecrease {
				signalCount++
				signals = append(signals, "创新高但成交量明显低于前高")
			}
			if consecutiveLowBreak {
				signalCount++
				signals = append(signals, "连续两次回调都打穿前低")
			}

			// 判断是否为牛市末期
			if signalCount >= 2 {
				params.MarketPhase = "牛市末期"
				noteBuilder := fmt.Sprintf("⚠️  检测到 %d 个牛市末期信号：", signalCount)
				for i, sig := range signals {
					noteBuilder += fmt.Sprintf("\n  %d. %s", i+1, sig)
				}
				noteBuilder += "\n建议：谨慎做多，关注趋势反转风险"
				params.MarketNote = noteBuilder

				fmt.Println()
				printWarning(fmt.Sprintf("检测到 %d 个牛市末期信号", signalCount))
				printError("市场阶段: 牛市末期")
				fmt.Println()
			} else if signalCount > 0 {
				params.MarketPhase = "牛市"
				noteBuilder := fmt.Sprintf("检测到 %d 个末期信号，暂未达到警戒线(2个)：", signalCount)
				for i, sig := range signals {
					noteBuilder += fmt.Sprintf("\n  %d. %s", i+1, sig)
				}
				params.MarketNote = noteBuilder
			} else {
				params.MarketPhase = "牛市"
				params.MarketNote = "未检测到牛市末期信号，市场处于健康状态"
			}
		} else if marketContextStr == "熊市" {
			params.MarketContext = models.MarketContextBear
			params.MarketPhase = "熊市"
		} else {
			params.MarketContext = models.MarketContextNone
			params.MarketPhase = "震荡"
		}
	}

	fmt.Println()
	printDivider()
	fmt.Println()

	// 开仓时间（可选）
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
			Message: "开仓时间 (格式: 2006-01-02 15:04:05):",
		}
		if err := survey.AskOne(customTimePrompt, &timeStr); err != nil {
			return err
		}
		if timeStr != "" {
			t, err := time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				return fmt.Errorf("无效的时间格式: %w", err)
			}
			params.OpenTime = &t
		}
	}

	// 执行开仓操作
	pos, err := ops.OpenPosition(params)
	if err != nil {
		printError(fmt.Sprintf("开仓失败: %v", err))
		return err
	}

	// 显示成功信息
	fmt.Println()
	printSuccess("仓位已记录")
	printHighlightField("仓位ID", pos.PositionID)
	printDivider()
	printField("品种", fmt.Sprintf("%s (%s)", pos.Symbol, pos.MarketType))
	printField("方向", pos.Direction)
	printField("开仓价格", fmt.Sprintf("%.4f", pos.OpenPrice))
	printField("仓位大小", fmt.Sprintf("%.4f", pos.Quantity))
	printField("止损", fmt.Sprintf("%.4f", pos.StopLoss))
	printField("止盈", fmt.Sprintf("%.4f", pos.TakeProfit))
	printField("保证金", fmt.Sprintf("%.2f", pos.Margin))
	if pos.Reason != "" {
		printField("理由", pos.Reason)
	}

	// 显示市场背景信息
	if pos.MarketContext != "" && pos.MarketContext != models.MarketContextNone {
		fmt.Println()
		printDivider()
		if pos.MarketPhase == "牛市末期" {
			printError(fmt.Sprintf("市场背景: %s - %s", pos.MarketContext, pos.MarketPhase))
		} else {
			printInfo(fmt.Sprintf("市场背景: %s - %s", pos.MarketContext, pos.MarketPhase))
		}
		if pos.MarketNote != "" {
			fmt.Println()
			printField("备注", pos.MarketNote)
		}
	}
	fmt.Println()

	return nil
}
