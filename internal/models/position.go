package models

import (
	"crypto/rand"
	"fmt"
	"time"
)

// Direction 交易方向
type Direction string

const (
	DirectionLong  Direction = "long"
	DirectionShort Direction = "short"
)

// MarketType 市场类型
type MarketType string

const (
	MarketTypeCrypto  MarketType = "crypto"
	MarketTypeForex   MarketType = "forex"
	MarketTypeGold    MarketType = "gold"
	MarketTypeSilver  MarketType = "silver"
	MarketTypeFutures MarketType = "futures"
)

// Status 仓位状态
type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

// CloseReason 平仓原因
type CloseReason string

const (
	CloseReasonStopLoss   CloseReason = "stop_loss"
	CloseReasonTakeProfit CloseReason = "take_profit"
	CloseReasonManual     CloseReason = "manual"
)

// Position 仓位信息
type Position struct {
	// 开仓信息
	PositionID     string     `json:"positionId"`
	AccountName    string     `json:"accountName"`              // 账户名称
	AccountBalance float64    `json:"accountBalance,omitempty"` // 开仓时的账户余额
	Symbol         string     `json:"symbol"`
	MarketType     MarketType `json:"marketType"`
	OpenTime       time.Time  `json:"openTime"`
	Direction      Direction  `json:"direction"`
	OpenPrice      float64    `json:"openPrice"`
	Quantity       float64    `json:"quantity"`
	StopLoss       float64    `json:"stopLoss"`
	TakeProfit     float64    `json:"takeProfit"`
	Margin         float64    `json:"margin"`
	Reason         string     `json:"reason,omitempty"`
	Status         Status     `json:"status"`

	// 平仓信息（可选）
	CloseTime       *time.Time   `json:"closeTime,omitempty"`
	ClosePrice      *float64     `json:"closePrice,omitempty"`
	CloseQuantity   *float64     `json:"closeQuantity,omitempty"`
	RealizedPnL     *float64     `json:"realizedPnL,omitempty"`
	PnLPercentage   *float64     `json:"pnlPercentage,omitempty"`   // 占账户余额的百分比
	MarginROI       *float64     `json:"marginROI,omitempty"`       // 保证金回报率
	HoldingDuration *string      `json:"holdingDuration,omitempty"`
	CloseReason     *CloseReason `json:"closeReason,omitempty"`
	CloseNote       string       `json:"closeNote,omitempty"`
}

// GeneratePositionID 生成唯一的仓位ID
// 格式: YYYYMMDD-HHMMSS-XXXX
func GeneratePositionID() string {
	now := time.Now()
	timestamp := now.Format("20060102-150405")

	randomBytes := make([]byte, 2)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// 如果随机数生成失败，使用纳秒作为后备
		randomHex := fmt.Sprintf("%04X", now.Nanosecond()%65536)
		return fmt.Sprintf("%s-%s", timestamp, randomHex)
	}
	randomHex := fmt.Sprintf("%04X", uint16(randomBytes[0])<<8|uint16(randomBytes[1]))

	return fmt.Sprintf("%s-%s", timestamp, randomHex)
}

// CalculateRealizedPnL 计算实际盈亏
func CalculateRealizedPnL(direction Direction, openPrice, closePrice, quantity float64) float64 {
	if direction == DirectionLong {
		return (closePrice - openPrice) * quantity
	}
	// DirectionShort
	return (openPrice - closePrice) * quantity
}

// CalculatePnLPercentage 计算盈亏占账户余额的百分比
func CalculatePnLPercentage(realizedPnL, accountBalance float64) float64 {
	if accountBalance == 0 {
		return 0
	}
	return (realizedPnL / accountBalance) * 100
}

// CalculateMarginROI 计算保证金回报率
func CalculateMarginROI(realizedPnL, margin float64) float64 {
	if margin == 0 {
		return 0
	}
	return (realizedPnL / margin) * 100
}

// FormatHoldingDuration 格式化持仓时长
func FormatHoldingDuration(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
