package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"trading-journal-cli/internal/models"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "账户管理",
	Long:  `管理交易账户，包括添加、查看、更新和删除账户`,
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有账户",
	RunE:  runAccountList,
}

var accountAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新账户",
	RunE:  runAccountAdd,
}

var accountUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新账户余额",
	RunE:  runAccountUpdate,
}

var accountDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除账户",
	RunE:  runAccountDelete,
}

var accountTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "设置账户开仓模板",
	RunE:  runAccountTemplate,
}

func init() {
	accountCmd.AddCommand(accountListCmd)
	accountCmd.AddCommand(accountAddCmd)
	accountCmd.AddCommand(accountUpdateCmd)
	accountCmd.AddCommand(accountDeleteCmd)
	accountCmd.AddCommand(accountTemplateCmd)
	rootCmd.AddCommand(accountCmd)
}

func getAccountManager() *models.AccountManager {
	am := models.NewAccountManager(dataDir)
	if err := am.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	return am
}

func runAccountList(cmd *cobra.Command, args []string) error {
	am := getAccountManager()
	accounts := am.ListAccounts()

	printTitle("💼 账户管理")

	if len(accounts) == 0 {
		printWarning("暂无账户")
		printHint("使用 'trading-cli account add' 添加新账户")
		return nil
	}

	for i, acc := range accounts {
		if i > 0 {
			printDivider()
		}

		currency := acc.Currency
		if currency == "" {
			currency = "USD"
		}

		// 账户名称
		printHighlightField("账户", acc.Name)
		printField("余额", fmt.Sprintf("%.2f %s", acc.Balance, currency))

		// 模板信息
		if acc.Template != nil {
			fmt.Print("  ")
			colorMuted.Print("模板           ")
			colorSuccess.Print("✓ 已设置 ")
			colorMuted.Print("(")

			templateParts := []string{}
			if acc.Template.DefaultSymbol != "" {
				templateParts = append(templateParts, acc.Template.DefaultSymbol)
			}
			if acc.Template.DefaultMarketType != "" {
				templateParts = append(templateParts, string(acc.Template.DefaultMarketType))
			}
			if acc.Template.DefaultDirection != "" {
				templateParts = append(templateParts, string(acc.Template.DefaultDirection))
			}

			if len(templateParts) > 0 {
				colorMuted.Print(strings.Join(templateParts, ", "))
			}
			colorMuted.Println(")")
		} else {
			fmt.Print("  ")
			colorMuted.Print("模板           ")
			colorWarning.Println("未设置")
		}
	}

	fmt.Println()
	printHint("使用 'trading-cli account template' 设置账户模板")
	fmt.Println()

	return nil
}

func runAccountAdd(cmd *cobra.Command, args []string) error {
	am := getAccountManager()

	printTitle("➕ 添加新账户")

	var name string
	namePrompt := &survey.Input{
		Message: "账户名称 (如 \"黄金账户\", \"BTC账户\"):",
	}
	if err := survey.AskOne(namePrompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var balanceStr string
	balancePrompt := &survey.Input{
		Message: "账户余额:",
	}
	if err := survey.AskOne(balancePrompt, &balanceStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var balance float64
	if _, err := fmt.Sscanf(balanceStr, "%f", &balance); err != nil {
		return fmt.Errorf("无效的余额格式: %w", err)
	}

	var currency string
	currencyPrompt := &survey.Input{
		Message: "币种 (可选，默认 USD):",
		Default: "USD",
	}
	survey.AskOne(currencyPrompt, &currency)

	account := models.Account{
		Name:     name,
		Balance:  balance,
		Currency: currency,
	}

	if err := am.AddAccount(account); err != nil {
		printError(fmt.Sprintf("添加账户失败: %v", err))
		return err
	}

	fmt.Println()
	printSuccess("账户已添加")
	printHighlightField("账户名称", name)
	printField("余额", fmt.Sprintf("%.2f %s", balance, currency))
	fmt.Println()

	return nil
}

func runAccountUpdate(cmd *cobra.Command, args []string) error {
	am := getAccountManager()
	accounts := am.ListAccounts()

	if len(accounts) == 0 {
		fmt.Println("暂无账户")
		return nil
	}

	// 选择账户
	options := make([]string, len(accounts))
	for i, acc := range accounts {
		options[i] = fmt.Sprintf("%s (%.2f %s)", acc.Name, acc.Balance, acc.Currency)
	}

	var selectedIndex int
	selectPrompt := &survey.Select{
		Message: "选择要更新的账户:",
		Options: options,
	}
	if err := survey.AskOne(selectPrompt, &selectedIndex); err != nil {
		return err
	}

	selectedAccount := accounts[selectedIndex]

	// 输入新余额
	var balanceStr string
	balancePrompt := &survey.Input{
		Message: "新余额:",
		Default: fmt.Sprintf("%.2f", selectedAccount.Balance),
	}
	if err := survey.AskOne(balancePrompt, &balanceStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var newBalance float64
	if _, err := fmt.Sscanf(balanceStr, "%f", &newBalance); err != nil {
		return fmt.Errorf("无效的余额格式: %w", err)
	}

	if err := am.UpdateAccount(selectedAccount.Name, newBalance); err != nil {
		return fmt.Errorf("更新账户失败: %w", err)
	}

	fmt.Printf("\n✓ 账户已更新: %s (%.2f -> %.2f %s)\n",
		selectedAccount.Name, selectedAccount.Balance, newBalance, selectedAccount.Currency)
	return nil
}

func runAccountDelete(cmd *cobra.Command, args []string) error {
	am := getAccountManager()
	accounts := am.ListAccounts()

	if len(accounts) == 0 {
		fmt.Println("暂无账户")
		return nil
	}

	// 选择账户
	options := make([]string, len(accounts))
	for i, acc := range accounts {
		options[i] = fmt.Sprintf("%s (%.2f %s)", acc.Name, acc.Balance, acc.Currency)
	}

	var selectedIndex int
	selectPrompt := &survey.Select{
		Message: "选择要删除的账户:",
		Options: options,
	}
	if err := survey.AskOne(selectPrompt, &selectedIndex); err != nil {
		return err
	}

	selectedAccount := accounts[selectedIndex]

	// 确认删除
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("确认删除账户 \"%s\"?", selectedAccount.Name),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("已取消")
		return nil
	}

	if err := am.DeleteAccount(selectedAccount.Name); err != nil {
		return fmt.Errorf("删除账户失败: %w", err)
	}

	fmt.Printf("\n✓ 账户已删除: %s\n", selectedAccount.Name)
	return nil
}

func runAccountTemplate(cmd *cobra.Command, args []string) error {
	am := getAccountManager()
	accounts := am.ListAccounts()

	printTitle("🎯 设置账户模板")

	if len(accounts) == 0 {
		printWarning("暂无账户")
		printHint("请先使用 'trading-cli account add' 添加账户")
		return nil
	}

	// 选择账户
	options := make([]string, len(accounts))
	for i, acc := range accounts {
		options[i] = fmt.Sprintf("%s (%.2f %s)", acc.Name, acc.Balance, acc.Currency)
	}

	var selectedIndex int
	selectPrompt := &survey.Select{
		Message: "选择要设置模板的账户:",
		Options: options,
	}
	if err := survey.AskOne(selectPrompt, &selectedIndex); err != nil {
		return err
	}

	selectedAccount := accounts[selectedIndex]

	// 显示当前模板（如果有）
	if selectedAccount.Template != nil {
		fmt.Println()
		printInfo("当前模板配置:")
		if selectedAccount.Template.DefaultMarketType != "" {
			printField("市场类型", selectedAccount.Template.DefaultMarketType)
		}
		if selectedAccount.Template.DefaultSymbol != "" {
			printField("品种", selectedAccount.Template.DefaultSymbol)
		}
		if selectedAccount.Template.DefaultDirection != "" {
			printField("方向", selectedAccount.Template.DefaultDirection)
		}
	}

	fmt.Println()
	printDivider()
	fmt.Println()

	template := &models.AccountTemplate{}

	// 设置默认市场类型
	var marketTypeStr string
	marketTypePrompt := &survey.Select{
		Message: "默认市场类型 (可选，按ESC跳过):",
		Options: []string{"crypto", "forex", "gold", "silver", "futures", "(不设置)"},
	}
	if err := survey.AskOne(marketTypePrompt, &marketTypeStr); err != nil {
		return err
	}
	if marketTypeStr != "(不设置)" {
		template.DefaultMarketType = models.MarketType(marketTypeStr)
	}

	// 设置默认品种
	var symbol string
	symbolPrompt := &survey.Input{
		Message: "默认品种 (可选，如 BTC/USDT，留空跳过):",
	}
	if err := survey.AskOne(symbolPrompt, &symbol); err != nil {
		return err
	}
	template.DefaultSymbol = symbol

	// 设置默认方向
	var directionStr string
	directionPrompt := &survey.Select{
		Message: "默认方向 (可选):",
		Options: []string{"long", "short", "(不设置)"},
	}
	if err := survey.AskOne(directionPrompt, &directionStr); err != nil {
		return err
	}
	if directionStr != "(不设置)" {
		template.DefaultDirection = models.Direction(directionStr)
	}

	// 保存模板
	if err := am.UpdateAccountTemplate(selectedAccount.Name, template); err != nil {
		printError(fmt.Sprintf("保存模板失败: %v", err))
		return err
	}

	fmt.Println()
	printSuccess("账户模板已更新")
	printHighlightField("账户", selectedAccount.Name)
	fmt.Println()

	if template.DefaultMarketType != "" || template.DefaultSymbol != "" || template.DefaultDirection != "" {
		printInfo("模板配置:")
		if template.DefaultMarketType != "" {
			printField("市场类型", template.DefaultMarketType)
		}
		if template.DefaultSymbol != "" {
			printField("品种", template.DefaultSymbol)
		}
		if template.DefaultDirection != "" {
			printField("方向", template.DefaultDirection)
		}
	} else {
		printWarning("未设置任何默认值")
	}
	fmt.Println()

	return nil
}
