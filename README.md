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

**账户余额自动更新**：每次平仓后，系统会自动将盈亏金额加到账户余额上，无需手动更新。

### 开仓记录

```bash
trading-cli open
```

交互式提示会引导你填写以下信息：
- **选择账户**（从已配置的账户中选择）
- 交易品种（如 BTC/USDT）
- 市场类型（crypto, forex, gold, silver, futures, cn_stocks, us_stocks）
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
4. **手动输入盈亏**（可选）：适用于外汇等需要汇率转换的复杂计算场景
5. 自动计算盈亏、盈亏百分比和持仓时长
6. 自动更新账户余额

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

**列表显示**：
- 表格格式会显示所有关键信息，包括仓位ID、品种、方向、价格、数量、状态、盈亏
- 对于已平仓记录，会显示"**平仓后余额**"列，按时间顺序累积计算每笔交易后的账户余额
- 使用颜色区分盈利（绿色）和亏损（红色）

### 数据分析（通过 Claude Code）

#### 快速分析 - 使用 Skill（推荐）

项目内置了 `/anasisly-trading` 技能，可快速生成约 300 字的精简诊断报告：

```bash
# 在 Claude Code 中运行（自动分析当前月份）
/anasisly-trading

# 或指定具体月份文件
/anasisly-trading ./trading-data/trades-2025-12.jsonl
```

**分析内容**：
- 核心数据：总盈亏、胜率、止盈/止损率、平均持仓时间
- 问题检测：风险管理、执行纪律、心理模式（标注 🔴 高危 / 🟡 警告）
- 改进建议：3-5 条按优先级排序的可执行措施
- 下月目标：可衡量的具体目标

**输出示例**：
```
🔴 仓位管理失控 - 保证金从 $8 暴增至 $600 (75倍)
🔴 FOMO 驱动的冲动交易 - 2小时内完成4笔交易
🟡 执行纪律缺失 - 75% 手动平仓率

改进建议：
1. [最关键] 固定仓位管理规则 - 每笔保证金 ≤ 账户 2%
2. [重要] 强制冷静期机制 - 交易间隔 ≥ 24 小时
3. [建议] 重建止损/止盈纪律
```

报告自动保存至 `./trading-data/reports/diagnosis-YYYY-MM.md`

#### 深度分析 - 自定义提问

你也可以直接向 Claude Code 提问：
- "分析我在 BTC/USDT 上的交易表现"
- "计算我本月的总盈亏"
- "找出我最常犯的交易错误"
- "对比黄金账户和 BTC 账户的收益差异"

Claude Code 会自动读取 `trading-data/` 中的数据并提供详细分析。

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

系统提供两种盈亏指标，并支持手动输入盈亏金额。

### 计算方式

**自动计算**（默认）：

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

**手动输入盈亏**：

对于外汇等需要复杂计算的交易（如合约规模乘数、汇率转换），平仓时可以选择手动输入实际盈亏金额：

```bash
trading-cli close
# ... 选择仓位，输入平仓价格和数量 ...
? 是否手动输入盈亏金额? Yes
? 盈亏金额 (正数为盈利，负数为亏损): 44.33
```

手动输入的盈亏会直接用于计算账户盈亏比和保证金ROI，无需进行自动计算。

**适用场景**：
- 外汇交易（需要合约规模和汇率转换）
- 复杂的衍生品交易
- 从交易平台直接获取实际盈亏数据

## 项目结构

```
trading-journal-cli/
├── .claude/
│   └── skills/            # Claude Code 技能
│       └── anasisly-trading/  # 交易日志分析技能
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
├── trading-data/          # 交易数据存储目录
│   └── reports/           # 分析报告（由 skill 生成）
├── main.go               # 程序入口
├── CLAUDE.md             # Claude Code 使用指南
└── README.md             # 本文档
```

## 支持的市场类型

- `crypto` - 加密货币
- `forex` - 外汇
- `gold` - 黄金
- `silver` - 白银
- `futures` - 期货
- `cn_stocks` - A股（中国股市）
- `us_stocks` - 美股（美国股市）

## 常见问题

### Q: 如何快速分析我的交易表现？
A: 在 Claude Code 中运行 `/anasisly-trading`，会自动生成精简诊断报告，识别关键问题并提供改进建议。报告保存在 `./trading-data/reports/` 目录。

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
