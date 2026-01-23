# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

trading-journal-cli is a Go-based command-line tool for recording and managing trading operations (opening/closing positions) with risk assessment and trading optimization analysis through natural language interaction with Claude Code. The system uses JSONL file storage without databases or backend services.

**Core Design Principles:**
- Simplicity: File system + JSONL format, no complex infrastructure
- Portability: Single compiled executable, cross-platform
- Flexibility: Natural language analysis via Claude Code, not hardcoded scripts
- Data Integrity: Mandatory validation of critical fields (stop-loss, take-profit)

## Architecture

### Directory Structure

```
trading-journal-cli/
├── cmd/
│   ├── root.go          # Root command and global config
│   ├── open.go          # Open position command
│   ├── close.go         # Close position command
│   └── list.go          # Query command
├── internal/
│   ├── models/
│   │   └── position.go  # Position data model
│   ├── storage/
│   │   └── jsonl.go     # JSONL file read/write
│   ├── validator/
│   │   └── validate.go  # Data validation logic
│   └── operations/
│       └── ops.go       # Open/close/query operations
├── prompts/
│   ├── risk-assessment.md      # Risk assessment prompt template
│   ├── trade-optimization.md   # Trade optimization prompt template
│   └── README.md               # Prompt usage guide
└── main.go
```

### Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: cobra (command line framework)
- **Interactive Input**: survey/v2 (user-friendly prompts)
- **Data Format**: JSONL (one JSON object per line)
- **Time Handling**: time standard library
- **ID Generation**: crypto/rand + timestamp

### Data Storage

**File Format**: JSONL (JSON Lines)
- File naming: `trades-YYYY-MM.jsonl` (e.g., `trades-2025-01.jsonl`)
- Location: `./trading-data/` directory
- Each trade record occupies one line as a complete JSON object
- For updates: append new version to file, query takes last record by positionId (preserves audit trail)

**Position ID Format**: `YYYYMMDD-HHMMSS-XXXX`
- `YYYYMMDD`: Year-Month-Day
- `HHMMSS`: Hour-Minute-Second
- `XXXX`: 4-digit random hexadecimal characters
- Example: `20250120-143022-A7B3`

## Position Data Model

### Core Structure

```go
type Position struct {
    // Open position info
    PositionID     string      `json:"positionId"`
    AccountName    string      `json:"accountName"`         // Account name
    AccountBalance float64     `json:"accountBalance"`      // Account balance at open time
    Symbol         string      `json:"symbol"`              // e.g., BTC/USDT, EUR/USD
    MarketType     MarketType  `json:"marketType"`          // crypto, forex, gold, silver, futures, cn_stocks, us_stocks
    OpenTime       time.Time   `json:"openTime"`
    Direction      Direction   `json:"direction"`           // long, short
    OpenPrice      float64     `json:"openPrice"`
    Quantity       float64     `json:"quantity"`
    StopLoss       float64     `json:"stopLoss"`            // REQUIRED
    TakeProfit     float64     `json:"takeProfit"`          // REQUIRED
    Margin         float64     `json:"margin"`
    Reason         string      `json:"reason,omitempty"`
    Status         Status      `json:"status"`              // open, closed

    // Close position info (optional)
    CloseTime       *time.Time   `json:"closeTime,omitempty"`
    ClosePrice      *float64     `json:"closePrice,omitempty"`
    CloseQuantity   *float64     `json:"closeQuantity,omitempty"`
    RealizedPnL     *float64     `json:"realizedPnL,omitempty"`
    PnLPercentage   *float64     `json:"pnlPercentage,omitempty"`    // Percentage of account balance
    MarginROI       *float64     `json:"marginROI,omitempty"`        // Margin return on investment
    HoldingDuration *string      `json:"holdingDuration,omitempty"`
    CloseReason     *CloseReason `json:"closeReason,omitempty"`      // stop_loss, take_profit, manual
    CloseNote       string       `json:"closeNote,omitempty"`
}
```

### Enums

```go
type Direction string
const (
    DirectionLong  Direction = "long"
    DirectionShort Direction = "short"
)

type MarketType string
const (
    MarketTypeCrypto   MarketType = "crypto"
    MarketTypeForex    MarketType = "forex"
    MarketTypeGold     MarketType = "gold"
    MarketTypeSilver   MarketType = "silver"
    MarketTypeFutures  MarketType = "futures"
    MarketTypeCNStocks MarketType = "cn_stocks"
    MarketTypeUSStocks MarketType = "us_stocks"
)

type Status string
const (
    StatusOpen   Status = "open"
    StatusClosed Status = "closed"
)

type CloseReason string
const (
    CloseReasonStopLoss   CloseReason = "stop_loss"
    CloseReasonTakeProfit CloseReason = "take_profit"
    CloseReasonManual     CloseReason = "manual"
)
```

## Critical Validation Rules

### Open Position Validation

**Required Fields**: symbol, openPrice, quantity, stopLoss, takeProfit, margin

**Price/Quantity Rules**:
- All prices and quantities must be positive (> 0)
- Stop-loss and take-profit are MANDATORY

**Stop-Loss/Take-Profit Range Rules**:
- **Long positions**: stopLoss < openPrice < takeProfit
- **Short positions**: takeProfit < openPrice < stopLoss

### Close Position Validation

- Position must exist and status must be "open"
- Close price must be positive
- Close quantity must not exceed position quantity (supports partial close)

### Validation Error Types

```go
var (
    ErrMissingField          = errors.New("required field is missing")
    ErrInvalidPrice          = errors.New("invalid price value")
    ErrInvalidQuantity       = errors.New("invalid quantity value")
    ErrInvalidStopLoss       = errors.New("stop loss must be set")
    ErrInvalidTakeProfit     = errors.New("take profit must be set")
    ErrStopLossRange         = errors.New("stop loss price out of valid range")
    ErrTakeProfitRange       = errors.New("take profit price out of valid range")
    ErrPositionNotFound      = errors.New("position not found")
    ErrPositionAlreadyClosed = errors.New("position already closed")
    ErrInvalidCloseQuantity  = errors.New("close quantity exceeds position quantity")
)
```

## PnL Calculation

**Manual PnL Input (Priority)**:
Users can manually input the realized PnL during position closing. This is useful for:
- Forex trading (requires contract size multiplication and exchange rate conversion)
- Complex derivative products
- Direct import of actual PnL from trading platforms

When `ManualPnL` is provided in `CloseParams`, it will be used directly instead of automatic calculation.

**Automatic Calculation (Default)**:

**Long Positions**:
```
realizedPnL = (closePrice - openPrice) * closeQuantity
pnlPercentage = (realizedPnL / accountBalance) * 100
marginROI = (realizedPnL / margin) * 100
```

**Short Positions**:
```
realizedPnL = (openPrice - closePrice) * closeQuantity
pnlPercentage = (realizedPnL / accountBalance) * 100
marginROI = (realizedPnL / margin) * 100
```

**Note**: `pnlPercentage` is calculated against `accountBalance` (real return), while `marginROI` is calculated against `margin` (capital efficiency).

**Holding Duration**:
```go
duration := closeTime.Sub(openTime)
holdingDuration := formatDuration(duration) // e.g., "2d 5h 30m"
```

## CLI Commands

### Open Position
```bash
trading-cli open

# Interactive prompts guide user through:
# - Symbol (e.g., BTC/USDT)
# - Market type selection
# - Direction (long/short)
# - Open price, quantity
# - Stop-loss price (REQUIRED)
# - Take-profit price (REQUIRED)
# - Margin/cost
# - Trading reason (optional)
```

### Close Position
```bash
trading-cli close

# Displays list of open positions
# User selects position by ID
# Interactive prompts for:
# - Close price
# - Close quantity (supports partial close)
# - Manual PnL input (optional) - user can choose to manually input realized PnL
# - Close reason (stop_loss/take_profit/manual)
# - Close note (optional)
# Automatically calculates PnL (if not manually provided), percentage, holding duration
# Automatically updates account balance (account balance += realized PnL)
```

### List/Query Positions
```bash
trading-cli list [flags]

Flags:
  --status string      Filter by status (open, closed, all) [default: all]
  --symbol string      Filter by trading symbol
  --market string      Filter by market type
  --account string     Filter by account name
  --from string        Start date (YYYY-MM-DD)
  --to string          End date (YYYY-MM-DD)
  --format string      Output format (table, json) [default: table]

Examples:
  trading-cli list --status open
  trading-cli list --symbol BTC/USDT --format json
  trading-cli list --from 2025-01-01 --to 2025-01-31
  trading-cli list --account "黄金账户" --status closed

Table Output Features:
  - Displays all key information including position ID, symbol, direction, prices, quantity, status, PnL
  - For closed positions, shows "Balance After Close" column that displays cumulative account balance
  - Balance is calculated chronologically (sorted by close time) to show actual balance progression
  - Uses color coding: green for profit, red for loss
```

## Property-Based Testing

The system implements 11 core correctness properties that should be verified through property-based testing (using `testing/quick` or `gopter`):

1. **Position ID Uniqueness**: All generated position IDs must be unique
2. **Stop-Loss/Take-Profit Required**: Open positions must have valid stop-loss and take-profit
3. **Market Type Support**: All market types (crypto, forex, gold, silver, futures, cn_stocks, us_stocks) supported
4. **Monthly Storage**: Positions stored in correct monthly JSONL file based on openTime
5. **Position ID Lookup**: Any saved position retrievable by its position ID
6. **Close Completeness**: Closed positions must have all close fields populated and status = "closed"
7. **Partial Close Quantity**: Close quantity must be <= position quantity
8. **Filter Accuracy**: Query results must match filter criteria exactly
9. **Serialization Round-Trip**: Position object → JSON → Position object must preserve all fields
10. **PnL Calculation Correctness**: PnL must match formula based on direction
11. **Stop-Loss/Take-Profit Range**: Stop-loss and take-profit must be in valid range for direction

**Property Test Configuration**:
- Minimum 100 iterations per property test
- Tag format: `// Feature: trading-journal-cli, Property N: <description>`
- Custom generators for Position objects that respect validation rules

## Error Handling Strategy

1. **Graceful Degradation**: Skip corrupted lines in JSONL, continue processing other records
2. **Clear Error Messages**: All errors include clear description and suggested solutions
3. **Data Integrity First**: Reject operations that fail validation, never save incomplete data
4. **Auto-Recovery**: Auto-create missing directories, initialize data structures on first use

## Analysis Features (via Claude Code)

The system provides prompt templates for Claude Code analysis rather than hardcoded analytics:

### Risk Assessment Prompt
- Analyze current open positions
- Calculate total margin and maximum possible loss
- Compute risk/reward ratio per position
- Evaluate position concentration by market type and symbol
- Provide risk warnings (e.g., single symbol > 40%)

### Trade Optimization Prompt
- Analyze historical performance (win rate, avg PnL ratio)
- Identify best/worst performing symbols and market types
- Analyze stop-loss/take-profit effectiveness
- Examine holding duration vs PnL relationship
- Detect trading patterns and habits

### Custom Analysis
Users can directly ask Claude Code any question about trading data. Claude Code reads JSONL files and provides flexible analysis.

## Implementation Notes

**Storage Strategy**:
- JSONL is append-only
- Updates append new version to file (preserves history)
- Queries take last record with matching positionId for current state
- Enables full audit trail

**File I/O Error Handling**:
- Missing directory → auto-create
- Missing JSONL file → return empty list (first use)
- JSON parse error → log error line, skip, continue
- Permission errors → return clear error message

**Minimal Dependencies**:
- Use Go standard library wherever possible
- Only external deps: cobra (CLI), survey (interactive prompts)
- No database, no web server, no complex infrastructure
