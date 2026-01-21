package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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

func init() {
	accountCmd.AddCommand(accountListCmd)
	accountCmd.AddCommand(accountAddCmd)
	accountCmd.AddCommand(accountUpdateCmd)
	accountCmd.AddCommand(accountDeleteCmd)
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

	if len(accounts) == 0 {
		fmt.Println("暂无账户，使用 'trading-cli account add' 添加账户")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "账户名称\t余额\t币种")
	fmt.Fprintln(w, "---\t---\t---")

	for _, acc := range accounts {
		currency := acc.Currency
		if currency == "" {
			currency = "USD"
		}
		fmt.Fprintf(w, "%s\t%.2f\t%s\n", acc.Name, acc.Balance, currency)
	}

	return nil
}

func runAccountAdd(cmd *cobra.Command, args []string) error {
	am := getAccountManager()

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
		return fmt.Errorf("添加账户失败: %w", err)
	}

	fmt.Printf("\n✓ 账户已添加: %s (%.2f %s)\n", name, balance, currency)
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
