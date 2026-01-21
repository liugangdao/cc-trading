# 快速开始

5分钟快速上手交易日志CLI系统。

## 安装

### 选项1：直接编译（推荐）

```bash
# 克隆或下载项目
cd trading-journal-cli

# 编译
go build -o trading-cli

# 测试安装
./trading-cli --help
```

### 选项2：直接运行

```bash
# 不编译，直接运行
go run main.go --help
```

## 第一笔交易

### 1. 记录开仓

```bash
./trading-cli open
```

按提示输入信息，例如：
- 品种: `BTC/USDT`
- 市场: `crypto`
- 方向: `long`
- 开仓价: `42500`
- 数量: `0.5`
- 止损: `41000` ⚠️ **必填**
- 止盈: `45000` ⚠️ **必填**
- 保证金: `5000`

### 2. 查看未平仓位

```bash
./trading-cli list --status open
```

### 3. 平仓

```bash
./trading-cli close
```

选择要平仓的仓位，输入平仓信息：
- 平仓价: `44200`
- 数量: `0.5`
- 原因: `take_profit`

系统会自动计算盈亏和持仓时长。

### 4. 查看历史

```bash
# 查看所有交易
./trading-cli list

# 只看已平仓
./trading-cli list --status closed

# 导出为JSON
./trading-cli list --format json > trades.json
```

## 数据分析

### 使用Claude Code

1. 在Claude Code中打开项目目录
2. 复制 `prompts/risk-assessment.md` 或 `prompts/trade-optimization.md`
3. 粘贴到对话框，Claude Code会自动分析数据

或者直接提问：
- "分析我当前的持仓风险"
- "统计我本月的总盈亏"
- "找出表现最好的交易品种"

## 常用命令

```bash
# 开仓
./trading-cli open

# 平仓
./trading-cli close

# 查看所有交易
./trading-cli list

# 只看未平仓
./trading-cli list --status open

# 只看已平仓
./trading-cli list --status closed

# 按品种筛选
./trading-cli list --symbol BTC/USDT

# 按市场筛选
./trading-cli list --market crypto

# 按日期筛选
./trading-cli list --from 2025-01-01 --to 2025-01-31

# JSON格式输出
./trading-cli list --format json

# 自定义数据目录
./trading-cli --data-dir /path/to/data open
```

## 数据位置

交易数据保存在 `./trading-data/` 目录：
```
trading-data/
├── trades-2025-01.jsonl
├── trades-2025-02.jsonl
└── ...
```

## 运行测试

```bash
# 运行所有测试
go test ./... -v

# 测试特定包
go test ./internal/models -v
go test ./internal/validator -v
```

## 下一步

- 📖 阅读 [README.md](README.md) 了解完整功能
- 💡 查看 [EXAMPLES.md](EXAMPLES.md) 学习实际使用场景
- 🔍 探索 [prompts/](prompts/) 目录了解数据分析功能

## 需要帮助？

- 查看 `./trading-cli --help`
- 查看子命令帮助: `./trading-cli open --help`
- 阅读文档: [README.md](README.md)

## 核心特性

✅ 强制风险管理（止损止盈必填）
✅ 自动盈亏计算
✅ 部分平仓支持
✅ 灵活查询筛选
✅ JSONL文件存储，易于备份
✅ Claude Code智能分析
✅ 跨平台支持

开始记录你的第一笔交易吧！🚀
