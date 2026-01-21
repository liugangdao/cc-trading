# 交易日志 CLI 系统 - 设计文档

## 概述

trading-journal-cli 是一个用 Golang 实现的命令行工具，用于记录和管理交易操作。系统采用简单的 JSONL 文件存储，无需数据库或后台服务。分析功能通过 Claude Code 的对话界面实现，而非硬编码的分析脚本。

核心设计原则：
- 简单性：使用文件系统和 JSONL 格式，无需复杂的基础设施
- 可移植性：编译为单一可执行文件，跨平台运行
- 灵活性：通过 Claude Code 进行自然语言分析，而非固定的分析脚本
- 数据完整性：强制验证关键字段（止损、止盈）

## 架构

### 系统组件

```
trading-journal-cli/
├── cmd/
│   ├── root.go          # 根命令和全局配置
│   ├── open.go          # 开仓命令
│   ├── close.go         # 平仓命令
│   └── list.go          # 查询命令
├── internal/
│   ├── models/
│   │   └── position.go  # 仓位数据模型
│   ├── storage/
│   │   └── jsonl.go     # JSONL 文件读写
│   ├── validator/
│   │   └── validate.go  # 数据验证逻辑
│   └── ui/
│       └── prompt.go    # 交互式输入界面
├── prompts/
│   ├── risk-assessment.md      # 风险评估 Prompt 模板
│   ├── trade-optimization.md   # 交易优化 Prompt 模板
│   └── README.md               # Prompt 使用说明
└── main.go
```

### 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: cobra (命令行框架)
- **交互式输入**: survey/v2 (用户友好的提示界面)
- **数据格式**: JSONL (每行一个 JSON 对象)
- **时间处理**: time 标准库
- **ID 生成**: crypto/rand + 时间戳

### 数据流

1. **开仓流程**:
   ```
   用户执行 `trading-cli open`
   → 交互式提示收集信息
   → 验证必填字段和数据格式
   → 生成唯一仓位 ID
   → 追加到当月 JSONL 文件
   → 显示确认信息
   ```

2. **平仓流程**:
   ```
   用户执行 `trading-cli close`
   → 读取所有 JSONL 文件
   → 筛选未平仓位
   → 显示列表供用户选择
   → 收集平仓信息
   → 计算盈亏和持仓时长
   → 更新原记录（追加新版本）
   → 显示确认信息
   ```

3. **查询流程**:
   ```
   用户执行 `trading-cli list [flags]`
   → 读取相关 JSONL 文件
   → 应用筛选条件
   → 格式化输出（表格或 JSON）
   ```

## 组件和接口

### Position 模型

```go
package models

import "time"

type Direction string
type MarketType string
type Status string
type CloseReason string

const (
    DirectionLong  Direction = "long"
    DirectionShort Direction = "short"
)

const (
    MarketTypeCrypto  MarketType = "crypto"
    MarketTypeForex   MarketType = "forex"
    MarketTypeGold    MarketType = "gold"
    MarketTypeSilver  MarketType = "silver"
    MarketTypeFutures MarketType = "futures"
)

const (
    StatusOpen   Status = "open"
    StatusClosed Status = "closed"
)

const (
    CloseReasonStopLoss   CloseReason = "stop_loss"
    CloseReasonTakeProfit CloseReason = "take_profit"
    CloseReasonManual     CloseReason = "manual"
)

type Position struct {
    // 开仓信息
    PositionID   string      `json:"positionId"`
    Symbol       string      `json:"symbol"`
    MarketType   MarketType  `json:"marketType"`
    OpenTime     time.Time   `json:"openTime"`
    Direction    Direction   `json:"direction"`
    OpenPrice    float64     `json:"openPrice"`
    Quantity     float64     `json:"quantity"`
    StopLoss     float64     `json:"stopLoss"`
    TakeProfit   float64     `json:"takeProfit"`
    Margin       float64     `json:"margin"`
    Reason       string      `json:"reason,omitempty"`
    Status       Status      `json:"status"`
    
    // 平仓信息（可选）
    CloseTime       *time.Time   `json:"closeTime,omitempty"`
    ClosePrice      *float64     `json:"closePrice,omitempty"`
    CloseQuantity   *float64     `json:"closeQuantity,omitempty"`
    RealizedPnL     *float64     `json:"realizedPnL,omitempty"`
    PnLPercentage   *float64     `json:"pnlPercentage,omitempty"`
    HoldingDuration *string      `json:"holdingDuration,omitempty"`
    CloseReason     *CloseReason `json:"closeReason,omitempty"`
    CloseNote       string       `json:"closeNote,omitempty"`
}
```

### Storage 接口

```go
package storage

import (
    "trading-journal-cli/internal/models"
    "time"
)

type Storage interface {
    // 追加新仓位记录
    AppendPosition(pos *models.Position) error
    
    // 读取指定月份的所有仓位
    ReadPositions(year int, month time.Month) ([]*models.Position, error)
    
    // 读取所有仓位（跨月份）
    ReadAllPositions() ([]*models.Position, error)
    
    // 读取未平仓位
    ReadOpenPositions() ([]*models.Position, error)
    
    // 更新仓位（实际是追加新版本）
    UpdatePosition(pos *models.Position) error
}

type JSONLStorage struct {
    dataDir string
}

func NewJSONLStorage(dataDir string) *JSONLStorage {
    return &JSONLStorage{dataDir: dataDir}
}
```

### Validator 接口

```go
package validator

import (
    "trading-journal-cli/internal/models"
    "errors"
)

var (
    ErrMissingField     = errors.New("required field is missing")
    ErrInvalidPrice     = errors.New("invalid price value")
    ErrInvalidQuantity  = errors.New("invalid quantity value")
    ErrInvalidStopLoss  = errors.New("stop loss must be set")
    ErrInvalidTakeProfit = errors.New("take profit must be set")
    ErrStopLossRange    = errors.New("stop loss price out of valid range")
    ErrTakeProfitRange  = errors.New("take profit price out of valid range")
)

type Validator interface {
    ValidateOpenPosition(pos *models.Position) error
    ValidateClosePosition(pos *models.Position) error
}

type PositionValidator struct{}

func NewPositionValidator() *PositionValidator {
    return &PositionValidator{}
}
```

### UI Prompt 接口

```go
package ui

import "trading-journal-cli/internal/models"

type Prompter interface {
    // 收集开仓信息
    PromptOpenPosition() (*models.Position, error)
    
    // 选择要平仓的仓位
    SelectPosition(positions []*models.Position) (*models.Position, error)
    
    // 收集平仓信息
    PromptCloseInfo(pos *models.Position) error
    
    // 确认操作
    Confirm(message string) (bool, error)
}

type InteractivePrompter struct{}

func NewInteractivePrompter() *InteractivePrompter {
    return &InteractivePrompter{}
}
```

## 数据模型

### JSONL 文件格式

每个交易记录占一行，为完整的 JSON 对象：

```jsonl
{"positionId":"20250120-143022-A7B3","symbol":"BTC/USDT","marketType":"crypto","openTime":"2025-01-20T14:30:22Z","direction":"long","openPrice":42500.00,"quantity":0.5,"stopLoss":41000.00,"takeProfit":45000.00,"margin":5000.00,"reason":"突破关键阻力位","status":"open"}
{"positionId":"20250120-143022-A7B3","symbol":"BTC/USDT","marketType":"crypto","openTime":"2025-01-20T14:30:22Z","direction":"long","openPrice":42500.00,"quantity":0.5,"stopLoss":41000.00,"takeProfit":45000.00,"margin":5000.00,"reason":"突破关键阻力位","status":"closed","closeTime":"2025-01-21T10:15:30Z","closePrice":44200.00,"closeQuantity":0.5,"realizedPnL":850.00,"pnlPercentage":17.0,"holdingDuration":"19h45m","closeReason":"take_profit"}
```

### 文件命名规则

- 格式: `trades-YYYY-MM.jsonl`
- 示例: `trades-2025-01.jsonl`, `trades-2025-02.jsonl`
- 位置: `./trading-data/` 目录

### 仓位 ID 生成规则

格式: `YYYYMMDD-HHMMSS-XXXX`
- `YYYYMMDD`: 年月日
- `HHMMSS`: 时分秒
- `XXXX`: 4位随机十六进制字符

示例: `20250120-143022-A7B3`

实现:
```go
func GeneratePositionID() string {
    now := time.Now()
    timestamp := now.Format("20060102-150405")
    
    randomBytes := make([]byte, 2)
    rand.Read(randomBytes)
    randomHex := fmt.Sprintf("%X", randomBytes)
    
    return fmt.Sprintf("%s-%s", timestamp, randomHex)
}
```

### 数据更新策略

JSONL 是追加式存储，不支持原地更新。平仓时的更新策略：
1. 读取所有记录到内存
2. 找到匹配的仓位 ID
3. 更新该记录的平仓字段
4. 将更新后的记录追加到文件末尾
5. 查询时，同一 positionId 取最后一条记录

这种设计保留了完整的历史记录，支持审计和回溯。

### 盈亏计算

**做多 (Long)**:
```
realizedPnL = (closePrice - openPrice) * closeQuantity
pnlPercentage = (realizedPnL / margin) * 100
```

**做空 (Short)**:
```
realizedPnL = (openPrice - closePrice) * closeQuantity
pnlPercentage = (realizedPnL / margin) * 100
```

### 持仓时长计算

```go
duration := closeTime.Sub(openTime)
holdingDuration := formatDuration(duration) // 如 "2d 5h 30m"
```

## CLI 命令设计

### 根命令

```bash
trading-cli [command]

Available Commands:
  open        开仓记录
  close       平仓记录
  list        查询交易记录
  help        帮助信息
  version     版本信息

Flags:
  -h, --help              帮助信息
  -d, --data-dir string   数据目录 (默认: "./trading-data")
```

### open 命令

```bash
trading-cli open

# 交互式提示：
? 交易品种 (如 BTC/USDT): BTC/USDT
? 市场类型: [crypto, forex, gold, silver, futures]
? 开仓时间 (留空使用当前时间): 
? 方向: [long, short]
? 开仓价格: 42500.00
? 仓位大小: 0.5
? 止损价格: 41000.00
? 止盈价格: 45000.00
? 保证金/成本: 5000.00
? 交易理由 (可选): 突破关键阻力位

✓ 仓位已记录: 20250120-143022-A7B3
```

### close 命令

```bash
trading-cli close

# 显示未平仓位列表：
未平仓位:
1. [20250120-143022-A7B3] BTC/USDT (long) @ 42500.00
2. [20250119-091530-B2C4] EUR/USD (short) @ 1.0850

? 选择要平仓的仓位: 1

? 平仓时间 (留空使用当前时间): 
? 平仓价格: 44200.00
? 平仓数量 (最大 0.5): 0.5
? 平仓原因: [stop_loss, take_profit, manual]
? 平仓备注 (可选): 达到止盈目标

✓ 仓位已平仓: 20250120-143022-A7B3
  盈亏: +850.00 (+17.0%)
  持仓时长: 19h 45m
```

### list 命令

```bash
trading-cli list [flags]

Flags:
  --status string      筛选状态 (open, closed, all) (默认: all)
  --symbol string      筛选交易品种
  --market string      筛选市场类型
  --from string        起始日期 (YYYY-MM-DD)
  --to string          结束日期 (YYYY-MM-DD)
  --format string      输出格式 (table, json) (默认: table)

示例:
  trading-cli list --status open
  trading-cli list --symbol BTC/USDT --format json
  trading-cli list --from 2025-01-01 --to 2025-01-31
```

输出示例（表格格式）:
```
ID                      Symbol      Market  Direction  Open Price  Status  PnL
20250120-143022-A7B3   BTC/USDT    crypto  long       42500.00    closed  +850.00 (+17.0%)
20250119-091530-B2C4   EUR/USD     forex   short      1.0850      open    -
```

## Prompt 模板设计

### 风险评估 Prompt

文件: `prompts/risk-assessment.md`

```markdown
# 交易风险评估

请分析 `trading-data/` 目录中的交易数据，重点关注当前未平仓位的风险状况。

## 分析要求

1. **总体风险敞口**
   - 计算所有未平仓位的总保证金
   - 计算最大可能损失（所有止损触发）
   - 评估风险敞口占总资金的比例

2. **单个仓位风险**
   - 列出每个未平仓位的风险回报比 (Risk/Reward Ratio)
   - 计算: RR = (TakeProfit - OpenPrice) / (OpenPrice - StopLoss)
   - 标注风险回报比 < 2 的仓位

3. **仓位集中度**
   - 按市场类型统计仓位分布
   - 按交易品种统计仓位分布
   - 标注单一品种或市场占比 > 40% 的情况

4. **风险预警**
   - 是否有仓位未设置止损或止盈（理论上不应该有）
   - 是否有仓位的止损距离过大（> 5%）
   - 是否有过度集中的风险

## 输出格式

请以清晰的结构化格式输出，包括：
- 风险总览（数字和百分比）
- 仓位明细表
- 风险预警列表
- 优化建议
```

### 交易优化 Prompt

文件: `prompts/trade-optimization.md`

```markdown
# 交易优化分析

请分析 `trading-data/` 目录中的历史交易数据，提供基于数据的优化建议。

## 分析要求

1. **整体表现**
   - 总交易次数
   - 胜率（盈利交易 / 总交易）
   - 平均盈亏比
   - 总盈亏和盈亏百分比

2. **品种表现**
   - 按交易品种统计胜率和平均盈亏
   - 识别表现最好的 3 个品种
   - 识别表现最差的 3 个品种

3. **市场类型表现**
   - 按市场类型（crypto, forex, gold 等）统计
   - 比较不同市场的胜率和盈亏

4. **止损止盈分析**
   - 统计触发止损 vs 止盈 vs 手动平仓的比例
   - 分析止损止盈设置是否合理
   - 计算平均风险回报比

5. **持仓时长分析**
   - 统计平均持仓时长
   - 分析持仓时长与盈亏的关系
   - 识别最佳持仓时长区间

6. **交易模式**
   - 识别交易频率模式
   - 分析做多 vs 做空的表现差异
   - 识别可能的过度交易或不足交易

## 输出格式

请以清晰的结构化格式输出，包括：
- 关键指标总览
- 详细分析表格
- 可视化建议（如果可能）
- 具体优化建议（至少 5 条）
```

### Prompt 使用说明

文件: `prompts/README.md`

```markdown
# Claude Code 分析 Prompt 使用指南

## 快速开始

1. 在 Claude Code 中打开项目目录
2. 复制对应的 Prompt 内容
3. 粘贴到对话框并发送
4. Claude Code 会自动读取 `trading-data/` 目录中的 JSONL 文件进行分析

## 可用 Prompts

- `risk-assessment.md`: 当前持仓风险评估
- `trade-optimization.md`: 历史交易优化分析

## 自定义分析

你也可以直接向 Claude Code 提问，例如：
- "分析我在 BTC/USDT 上的交易表现"
- "计算我本月的总盈亏"
- "找出我最常犯的交易错误"
- "比较我在加密货币和外汇市场的表现"

Claude Code 会自动读取相关数据并提供分析。
```


## 正确性属性

正确性属性是关于系统行为的形式化陈述，应该在所有有效执行中保持为真。这些属性作为人类可读规范和机器可验证正确性保证之间的桥梁。

基于需求分析，我们识别出以下需要通过属性测试验证的核心正确性属性：

### Property 1: 仓位 ID 唯一性

*对于任意* 生成的仓位 ID 集合，所有 ID 必须是唯一的，不存在重复。

**验证需求: 2.1.2**

### Property 2: 止损止盈必填验证

*对于任意* 要保存的开仓记录，必须包含有效的止损价格和止盈价格（非零、非空），否则验证应该失败。

**验证需求: 2.1.3**

### Property 3: 市场类型支持

*对于任意* 有效的市场类型（crypto, forex, gold, silver, futures），系统应该能够正确接受、存储和检索该类型的仓位。

**验证需求: 2.1.4**

### Property 4: 按月份存储

*对于任意* 仓位记录，如果其开仓时间为 YYYY-MM，则该记录应该被保存到 `trades-YYYY-MM.jsonl` 文件中。

**验证需求: 2.1.5**

### Property 5: 仓位 ID 查找

*对于任意* 已保存的仓位，通过其仓位 ID 应该能够准确检索到该仓位的完整信息。

**验证需求: 2.2.2**

### Property 6: 平仓完整性

*对于任意* 仓位，当执行平仓操作后，该仓位应该包含所有必需的平仓字段（closeTime, closePrice, closeQuantity, realizedPnL, pnlPercentage, holdingDuration），并且状态应该从 "open" 变为 "closed"。

**验证需求: 2.2.3, 2.2.4**

### Property 7: 部分平仓数量

*对于任意* 仓位，如果执行部分平仓操作，平仓数量 closeQuantity 应该小于或等于原始数量 quantity，并且 closeQuantity 应该等于用户指定的平仓数量。

**验证需求: 2.2.5**

### Property 8: 筛选准确性

*对于任意* 筛选条件（状态、品种、市场类型、日期范围），查询结果中的所有仓位都应该满足该筛选条件，不应该包含不匹配的记录。

**验证需求: 2.2.1, 2.3.2, 2.3.3**

### Property 9: 序列化往返一致性

*对于任意* 有效的仓位对象，将其序列化为 JSON 并保存到 JSONL 文件，然后读取并反序列化，应该得到等价的仓位对象（所有字段值相同）。

**验证需求: 3.3**

### Property 10: 盈亏计算正确性

*对于任意* 平仓的仓位，其计算的 realizedPnL 应该等于：
- 做多时：`(closePrice - openPrice) * closeQuantity`
- 做空时：`(openPrice - closePrice) * closeQuantity`

并且 pnlPercentage 应该等于 `(realizedPnL / margin) * 100`。

**验证需求: 2.2.3**

### Property 11: 止损止盈范围验证

*对于任意* 做多仓位，止损价格应该小于开仓价格，止盈价格应该大于开仓价格。
*对于任意* 做空仓位，止损价格应该大于开仓价格，止盈价格应该小于开仓价格。

**验证需求: 2.1.3**

## 错误处理

### 验证错误

**开仓验证错误**:
- 缺少必填字段 → 返回 `ErrMissingField` 并指明具体字段
- 价格为负数或零 → 返回 `ErrInvalidPrice`
- 数量为负数或零 → 返回 `ErrInvalidQuantity`
- 未设置止损 → 返回 `ErrInvalidStopLoss`
- 未设置止盈 → 返回 `ErrInvalidTakeProfit`
- 止损价格范围错误 → 返回 `ErrStopLossRange`（做多时止损 >= 开仓价，做空时止损 <= 开仓价）
- 止盈价格范围错误 → 返回 `ErrTakeProfitRange`（做多时止盈 <= 开仓价，做空时止盈 >= 开仓价）

**平仓验证错误**:
- 仓位 ID 不存在 → 返回 `ErrPositionNotFound`
- 仓位已平仓 → 返回 `ErrPositionAlreadyClosed`
- 平仓数量超过持仓数量 → 返回 `ErrInvalidCloseQuantity`
- 平仓价格无效 → 返回 `ErrInvalidPrice`

### 文件 I/O 错误

**读取错误**:
- 数据目录不存在 → 自动创建目录
- JSONL 文件不存在 → 返回空列表（首次使用）
- JSON 解析失败 → 记录错误行号，跳过该行，继续处理其他行
- 文件权限错误 → 返回明确的权限错误信息

**写入错误**:
- 磁盘空间不足 → 返回 `ErrDiskFull`
- 文件权限错误 → 返回 `ErrPermissionDenied`
- JSON 序列化失败 → 返回 `ErrSerializationFailed`

### 错误处理策略

1. **优雅降级**: JSON 解析错误时跳过损坏的行，继续处理其他记录
2. **明确错误信息**: 所有错误都包含清晰的描述和建议的解决方案
3. **数据完整性优先**: 验证失败时拒绝操作，不保存不完整的数据
4. **自动恢复**: 缺少目录时自动创建，首次使用时初始化数据结构

### 错误日志

使用标准错误输出 (stderr) 记录错误：
```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
```

关键操作（保存、更新）失败时，提供详细的错误上下文：
```go
return fmt.Errorf("failed to save position %s: %w", pos.PositionID, err)
```

## 测试策略

### 双重测试方法

系统采用单元测试和基于属性的测试相结合的方法，以确保全面覆盖：

**单元测试**:
- 验证特定示例和边界情况
- 测试错误条件和异常处理
- 测试组件之间的集成点
- 关注具体的、可预测的场景

**基于属性的测试**:
- 验证跨所有输入的通用属性
- 通过随机化实现全面的输入覆盖
- 每个测试最少运行 100 次迭代
- 关注不变量和通用规则

两种方法是互补的：单元测试捕获具体的错误，属性测试验证通用正确性。

### 属性测试配置

**测试库选择**: 使用 Go 的 `testing/quick` 包或 `gopter` 库进行基于属性的测试

**测试配置**:
- 每个属性测试最少 100 次迭代
- 使用随机生成器创建测试数据
- 每个测试必须引用设计文档中的属性

**标签格式**:
```go
// Feature: trading-journal-cli, Property 1: 仓位 ID 唯一性
func TestProperty_PositionIDUniqueness(t *testing.T) {
    // ...
}
```

### 测试覆盖范围

**模型层测试**:
- Position 结构的验证逻辑
- 盈亏计算函数
- ID 生成函数
- 属性: 1, 2, 7, 10, 11

**存储层测试**:
- JSONL 读写操作
- 文件命名和路径处理
- 记录更新（追加）逻辑
- 跨月份查询
- 属性: 4, 5, 8, 9

**验证器测试**:
- 开仓验证规则
- 平仓验证规则
- 边界条件和错误情况
- 属性: 2, 11

**集成测试**:
- 完整的开仓-平仓流程
- 多个仓位的并发操作
- 跨月份的数据一致性
- 属性: 6, 9

### 单元测试示例

```go
func TestOpenPosition_MissingStopLoss(t *testing.T) {
    validator := NewPositionValidator()
    pos := &models.Position{
        Symbol:     "BTC/USDT",
        OpenPrice:  42500.0,
        Quantity:   0.5,
        TakeProfit: 45000.0,
        // StopLoss 缺失
    }
    
    err := validator.ValidateOpenPosition(pos)
    if err != ErrInvalidStopLoss {
        t.Errorf("Expected ErrInvalidStopLoss, got %v", err)
    }
}
```

### 属性测试示例

```go
// Feature: trading-journal-cli, Property 1: 仓位 ID 唯一性
func TestProperty_PositionIDUniqueness(t *testing.T) {
    config := &quick.Config{MaxCount: 100}
    
    property := func(count uint8) bool {
        if count == 0 {
            return true
        }
        
        ids := make(map[string]bool)
        for i := 0; i < int(count); i++ {
            id := GeneratePositionID()
            if ids[id] {
                return false // 发现重复
            }
            ids[id] = true
            time.Sleep(time.Millisecond) // 避免时间戳冲突
        }
        return true
    }
    
    if err := quick.Check(property, config); err != nil {
        t.Error(err)
    }
}

// Feature: trading-journal-cli, Property 9: 序列化往返一致性
func TestProperty_SerializationRoundTrip(t *testing.T) {
    config := &quick.Config{MaxCount: 100}
    
    property := func(pos *models.Position) bool {
        // 序列化
        data, err := json.Marshal(pos)
        if err != nil {
            return false
        }
        
        // 反序列化
        var loaded models.Position
        err = json.Unmarshal(data, &loaded)
        if err != nil {
            return false
        }
        
        // 比较
        return reflect.DeepEqual(pos, &loaded)
    }
    
    if err := quick.Check(property, config); err != nil {
        t.Error(err)
    }
}
```

### 测试数据生成

为属性测试实现自定义生成器：

```go
func (Position) Generate(rand *rand.Rand, size int) reflect.Value {
    directions := []Direction{DirectionLong, DirectionShort}
    markets := []MarketType{MarketTypeCrypto, MarketTypeForex, MarketTypeGold}
    
    direction := directions[rand.Intn(len(directions))]
    openPrice := 1000.0 + rand.Float64()*9000.0
    
    var stopLoss, takeProfit float64
    if direction == DirectionLong {
        stopLoss = openPrice * (0.95 - rand.Float64()*0.05)
        takeProfit = openPrice * (1.05 + rand.Float64()*0.10)
    } else {
        stopLoss = openPrice * (1.05 + rand.Float64()*0.05)
        takeProfit = openPrice * (0.95 - rand.Float64()*0.10)
    }
    
    pos := &Position{
        PositionID: GeneratePositionID(),
        Symbol:     "TEST/USD",
        MarketType: markets[rand.Intn(len(markets))],
        OpenTime:   time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
        Direction:  direction,
        OpenPrice:  openPrice,
        Quantity:   rand.Float64() * 10,
        StopLoss:   stopLoss,
        TakeProfit: takeProfit,
        Margin:     1000.0 + rand.Float64()*9000.0,
        Status:     StatusOpen,
    }
    
    return reflect.ValueOf(pos)
}
```

### 持续集成

测试应该在以下情况下自动运行：
- 每次代码提交
- Pull Request 创建时
- 发布前的最终验证

CI 配置应该：
- 运行所有单元测试
- 运行所有属性测试（100+ 迭代）
- 生成覆盖率报告（目标 > 80%）
- 在测试失败时阻止合并
