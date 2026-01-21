package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Account 账户信息
type Account struct {
	Name     string  `json:"name"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency,omitempty"`
	Template *AccountTemplate `json:"template,omitempty"` // 开仓模板
}

// AccountTemplate 账户开仓模板
type AccountTemplate struct {
	DefaultMarketType MarketType `json:"defaultMarketType,omitempty"` // 默认市场类型
	DefaultSymbol     string     `json:"defaultSymbol,omitempty"`     // 默认品种
	DefaultDirection  Direction  `json:"defaultDirection,omitempty"`  // 默认方向
}

// AccountConfig 账户配置文件结构
type AccountConfig struct {
	Accounts []Account `json:"accounts"`
}

// AccountManager 账户管理器
type AccountManager struct {
	configPath string
	config     *AccountConfig
}

// NewAccountManager 创建账户管理器
func NewAccountManager(dataDir string) *AccountManager {
	configPath := filepath.Join(dataDir, "accounts.json")
	return &AccountManager{
		configPath: configPath,
		config:     &AccountConfig{Accounts: []Account{}},
	}
}

// Load 加载账户配置
func (am *AccountManager) Load() error {
	if _, err := os.Stat(am.configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		return am.Save()
	}

	data, err := os.ReadFile(am.configPath)
	if err != nil {
		return fmt.Errorf("failed to read accounts config: %w", err)
	}

	if err := json.Unmarshal(data, &am.config); err != nil {
		return fmt.Errorf("failed to parse accounts config: %w", err)
	}

	return nil
}

// Save 保存账户配置
func (am *AccountManager) Save() error {
	// 确保目录存在
	dir := filepath.Dir(am.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(am.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts config: %w", err)
	}

	if err := os.WriteFile(am.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write accounts config: %w", err)
	}

	return nil
}

// ListAccounts 列出所有账户
func (am *AccountManager) ListAccounts() []Account {
	return am.config.Accounts
}

// GetAccount 获取指定账户
func (am *AccountManager) GetAccount(name string) (*Account, error) {
	for _, acc := range am.config.Accounts {
		if acc.Name == name {
			return &acc, nil
		}
	}
	return nil, fmt.Errorf("account not found: %s", name)
}

// AddAccount 添加账户
func (am *AccountManager) AddAccount(account Account) error {
	// 检查是否已存在
	for _, acc := range am.config.Accounts {
		if acc.Name == account.Name {
			return fmt.Errorf("account already exists: %s", account.Name)
		}
	}

	am.config.Accounts = append(am.config.Accounts, account)
	return am.Save()
}

// UpdateAccount 更新账户
func (am *AccountManager) UpdateAccount(name string, balance float64) error {
	found := false
	for i := range am.config.Accounts {
		if am.config.Accounts[i].Name == name {
			am.config.Accounts[i].Balance = balance
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", name)
	}

	return am.Save()
}

// DeleteAccount 删除账户
func (am *AccountManager) DeleteAccount(name string) error {
	found := false
	newAccounts := []Account{}
	for _, acc := range am.config.Accounts {
		if acc.Name != name {
			newAccounts = append(newAccounts, acc)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", name)
	}

	am.config.Accounts = newAccounts
	return am.Save()
}

// UpdateAccountTemplate 更新账户模板
func (am *AccountManager) UpdateAccountTemplate(name string, template *AccountTemplate) error {
	found := false
	for i := range am.config.Accounts {
		if am.config.Accounts[i].Name == name {
			am.config.Accounts[i].Template = template
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", name)
	}

	return am.Save()
}
