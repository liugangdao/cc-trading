# 交易日志 CLI 系统

一个基于 Go 的命令行工具，用于记录和管理交易操作（开仓、平仓），并通过 Claude Code 进行风险评估和交易优化分析。

## 特性

- 💼 **多账户管理** - 支持管理多个交易账户（如黄金账户、BTC账户）
- 📊 **交互式开仓记录** - 通过友好的提示界面记录新仓位
- 📉 **智能平仓管理** - 选择未平仓位并记录平仓信息，自动计算盈亏
- 🔍 **灵活查询筛选** - 按状态、品种、市场类型、账户、日期范围筛选交易记录
- 💰 **双重盈亏指标** - 同时显示账户盈亏比（真实收益）和保证金ROI（资金效率）
- ⚠️ **强制风险管理** - 开仓时必须设置止损和止盈
- 📈 **数据分析** - 通过 Claude Code 进行风险评估和表现分析
- 💾 **简单存储** - JSONL 文件存储，无需数据库
- 🚀 **跨平台支持** - 编译为单一可执行文件，支持 Windows/Mac/Linux

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone <repository-url>
cd trading-journal-cli

# 编译
go build -o trading-cli

# 或者安装到 $GOPATH/bin
go install
```

### 直接运行

```bash
go run main.go [command]
```

## 使用指南

### 账户管理

在开始记录交易之前，需要先添加账户：

```bash
# 添加账户
trading-cli account add
? 账户名称: 黄金账户
? 账户余额: 10000
? 币种: USD

# 列出所有账户
trading-cli account list

# 更新账户余额
trading-cli account update

# 删除账户
trading-cli account delete
```

系统支持多账户管理，适用于：
- 不同资产类别（如黄金账户、BTC账户）
- 不同交易平台
- 不同风险等级的资金划分

### 开仓记录

```bash
trading-cli open
```

交互式提示会引导你填写以下信息：
- **选择账户**（从已配置的账户中选择）
- 交易品种（如 BTC/USDT）
- 市场类型（crypto, forex, gold, silver, futures）
- 方向（long/short）
- 开仓价格
- 仓位大小
- 止损价格（必填）
- 止盈价格（必填）
- 保证金/成本
- 交易理由（可选）

系统会自动生成唯一的仓位 ID（格式：`YYYYMMDD-HHMMSS-XXXX`）。

### 平仓记录

```bash
trading-cli close
```

系统会：
1. 显示所有未平仓位列表
2. 让你选择要平仓的仓位
3. 引导你填写平仓信息（价格、数量、原因、备注）
4. 自动计算盈亏、盈亏百分比和持仓时长

支持部分平仓（平仓数量小于持仓数量）。

### 查询交易记录

```bash
# 查看所有交易
trading-cli list

# 只查看未平仓位
trading-cli list --status open

# 只查看已平仓位
trading-cli list --status closed

# 按账户筛选
trading-cli list --account "黄金账户"
trading-cli list --account "BTC账户"

# 按品种筛选
trading-cli list --symbol BTC/USDT

# 按市场类型筛选
trading-cli list --market crypto

# 按日期范围筛选
trading-cli list --from 2025-01-01 --to 2025-01-31

# JSON 格式输出
trading-cli list --format json

# 组合筛选
trading-cli list --status closed --account "黄金账户" --from 2025-01-01
```

### 数据分析（通过 Claude Code）

项目包含预设的分析 Prompt 模板，位于 `prompts/` 目录：

1. **风险评估** (`prompts/risk-assessment.md`)
   - 评估当前持仓风险
   - 计算风险回报比
   - 检查仓位集中度
   - 提供风险预警

2. **交易优化** (`prompts/trade-optimization.md`)
   - 分析历史交易表现
   - 计算胜率和盈亏比
   - 识别最佳/最差品种
   - 提供优化建议

**使用方法**：
1. 在 Claude Code 中打开项目目录
2. 复制 Prompt 内容
3. 粘贴到 Claude Code 对话框
4. Claude Code 会自动读取 `trading-data/` 中的数据并分析

你也可以直接向 Claude Code 提问：
- "分析我在 BTC/USDT 上的交易表现"
- "计算我本月的总盈亏"
- "找出我最常犯的交易错误"

## 数据存储

### 文件格式

- **格式**：JSONL (JSON Lines)
- **位置**：`./trading-data/`
- **命名**：`trades-YYYY-MM.jsonl`（按月份存储）

### 数据结构

每条记录包含以下字段：

```json
{
  "positionId": "20250120-143022-A7B3",
  "accountName": "黄金账户",
  "accountBalance": 10000.00,
  "symbol": "BTC/USDT",
  "marketType": "crypto",
  "openTime": "2025-01-20T14:30:22Z",
  "direction": "long",
  "openPrice": 42500.00,
  "quantity": 0.5,
  "stopLoss": 41000.00,
  "takeProfit": 45000.00,
  "margin": 5000.00,
  "reason": "突破关键阻力位",
  "status": "closed",
  "closeTime": "2025-01-21T10:15:30Z",
  "closePrice": 44200.00,
  "closeQuantity": 0.5,
  "realizedPnL": 850.00,
  "pnlPercentage": 8.5,
  "marginROI": 17.0,
  "holdingDuration": "19h 45m",
  "closeReason": "take_profit",
  "closeNote": "达到止盈目标"
}
```

**说明**：
- `pnlPercentage`: 占账户余额的百分比（真实收益率）
- `marginROI`: 保证金回报率（资金使用效率）

### 更新策略

JSONL 文件是追加式存储：
- 平仓时，系统会追加更新后的记录到文件末尾
- 同一 `positionId` 的最后一条记录代表最新状态
- 保留完整历史记录，支持审计追踪

## 验证规则

### 开仓验证

- 所有价格和数量必须为正数
- **止损和止盈必填**（强制风险管理）
- 做多仓位：`止损 < 开仓价 < 止盈`
- 做空仓位：`止盈 < 开仓价 < 止损`

### 平仓验证

- 仓位必须存在且状态为 "open"
- 平仓价格必须为正数
- 平仓数量不能超过持仓数量

## 盈亏计算

系统提供两种盈亏指标：

### 1. 账户盈亏比（pnlPercentage）

反映交易对整个账户的真实影响。

**做多（Long）**：
```
realizedPnL = (closePrice - openPrice) * closeQuantity
pnlPercentage = (realizedPnL / accountBalance) * 100
```

**做空（Short）**：
```
realizedPnL = (openPrice - closePrice) * closeQuantity
pnlPercentage = (realizedPnL / accountBalance) * 100
```

**示例**：账户余额 $10,000，盈利 $850，账户盈亏比 = 8.5%

### 2. 保证金回报率（marginROI）

反映资金使用效率。

```
marginROI = (realizedPnL / margin) * 100
```

**示例**：保证金 $5,000，盈利 $850，保证金ROI = 17%

**推荐用途**：
- 用**账户盈亏比**评估真实收益和风险
- 用**保证金ROI**评估资金使用效率

## 项目结构

```
trading-journal-cli/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令
│   ├── open.go            # 开仓命令
│   ├── close.go           # 平仓命令
│   └── list.go            # 查询命令
├── internal/
│   ├── models/            # 数据模型
│   ├── storage/           # JSONL 存储
│   ├── validator/         # 数据验证
│   └── operations/        # 业务操作
├── prompts/               # Claude Code 分析 Prompt
├── trading-data/          # 交易数据存储目录
├── main.go               # 程序入口
└── README.md             # 本文档
```

## 支持的市场类型

- `crypto` - 加密货币
- `forex` - 外汇
- `gold` - 黄金
- `silver` - 白银
- `futures` - 期货

## 常见问题

### Q: 数据存储在哪里？
A: 默认存储在 `./trading-data/` 目录，可以通过 `--data-dir` 参数修改。

### Q: 可以部分平仓吗？
A: 可以，平仓时可以指定少于持仓数量的平仓数量。

### Q: 如何备份数据？
A: 直接复制 `trading-data/` 目录下的 JSONL 文件即可。

### Q: 如果忘记设置止损怎么办？
A: 系统会拒绝没有止损和止盈的开仓请求，确保风险管理。

### Q: 如何查看特定品种的历史表现？
A: 使用 `trading-cli list --symbol <品种名> --status closed --format json` 或直接在 Claude Code 中询问。

## 技术栈

- Go 1.21+
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Survey](https://github.com/AlecAivazis/survey) - 交互式提示
- JSONL - 数据存储格式

## 许可证

MIT License
