package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"trading-journal-cli/internal/operations"
	"trading-journal-cli/internal/storage"
	"trading-journal-cli/internal/validator"
)

var (
	dataDir string
	ops     *operations.Operations
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "trading-cli",
	Short: "交易日志CLI系统",
	Long:  `通过命令行管理交易记录，包括开仓、平仓、查询和分析功能`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "d", "./trading-data", "数据目录")

	// 初始化操作实例
	cobra.OnInitialize(initOperations)
}

func initOperations() {
	store := storage.NewJSONLStorage(dataDir)
	valid := validator.NewPositionValidator()
	ops = operations.NewOperations(store, valid)
}
