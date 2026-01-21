package validator

import (
	"errors"
	"fmt"
	"trading-journal-cli/internal/models"
)

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

// Validator 验证器接口
type Validator interface {
	ValidateOpenPosition(pos *models.Position) error
	ValidateClosePosition(pos *models.Position, closeQuantity float64) error
}

// PositionValidator 仓位验证器
type PositionValidator struct{}

// NewPositionValidator 创建新的验证器
func NewPositionValidator() *PositionValidator {
	return &PositionValidator{}
}

// ValidateOpenPosition 验证开仓数据
func (v *PositionValidator) ValidateOpenPosition(pos *models.Position) error {
	// 验证必填字段
	if pos.Symbol == "" {
		return fmt.Errorf("%w: symbol", ErrMissingField)
	}
	if pos.MarketType == "" {
		return fmt.Errorf("%w: marketType", ErrMissingField)
	}
	if pos.Direction == "" {
		return fmt.Errorf("%w: direction", ErrMissingField)
	}

	// 验证价格和数量为正数
	if pos.OpenPrice <= 0 {
		return fmt.Errorf("%w: openPrice must be positive", ErrInvalidPrice)
	}
	if pos.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrInvalidQuantity)
	}
	if pos.Margin <= 0 {
		return fmt.Errorf("%w: margin must be positive", ErrInvalidPrice)
	}

	// 验证止损止盈必填
	if pos.StopLoss <= 0 {
		return ErrInvalidStopLoss
	}
	if pos.TakeProfit <= 0 {
		return ErrInvalidTakeProfit
	}

	// 验证止损止盈范围
	if pos.Direction == models.DirectionLong {
		// 做多: 止损 < 开仓价 < 止盈
		if pos.StopLoss >= pos.OpenPrice {
			return fmt.Errorf("%w: for long position, stop loss must be below open price", ErrStopLossRange)
		}
		if pos.TakeProfit <= pos.OpenPrice {
			return fmt.Errorf("%w: for long position, take profit must be above open price", ErrTakeProfitRange)
		}
	} else if pos.Direction == models.DirectionShort {
		// 做空: 止盈 < 开仓价 < 止损
		if pos.StopLoss <= pos.OpenPrice {
			return fmt.Errorf("%w: for short position, stop loss must be above open price", ErrStopLossRange)
		}
		if pos.TakeProfit >= pos.OpenPrice {
			return fmt.Errorf("%w: for short position, take profit must be below open price", ErrTakeProfitRange)
		}
	}

	return nil
}

// ValidateClosePosition 验证平仓数据
func (v *PositionValidator) ValidateClosePosition(pos *models.Position, closeQuantity float64) error {
	// 验证仓位状态
	if pos.Status == models.StatusClosed {
		return ErrPositionAlreadyClosed
	}

	// 验证平仓数量
	if closeQuantity <= 0 {
		return fmt.Errorf("%w: close quantity must be positive", ErrInvalidQuantity)
	}
	if closeQuantity > pos.Quantity {
		return fmt.Errorf("%w: close quantity %.4f exceeds position quantity %.4f",
			ErrInvalidCloseQuantity, closeQuantity, pos.Quantity)
	}

	return nil
}
