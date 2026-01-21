package validator

import (
	"testing"
	"time"
	"trading-journal-cli/internal/models"
)

func TestValidateOpenPosition_Success(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionLong,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   41000.0,
		TakeProfit: 45000.0,
		Margin:     5000.0,
		OpenTime:   time.Now(),
		Status:     models.StatusOpen,
	}

	err := validator.ValidateOpenPosition(pos)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

func TestValidateOpenPosition_MissingStopLoss(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionLong,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   0, // Missing stop loss
		TakeProfit: 45000.0,
		Margin:     5000.0,
	}

	err := validator.ValidateOpenPosition(pos)
	if err != ErrInvalidStopLoss {
		t.Errorf("Expected ErrInvalidStopLoss, got %v", err)
	}
}

func TestValidateOpenPosition_MissingTakeProfit(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionLong,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   41000.0,
		TakeProfit: 0, // Missing take profit
		Margin:     5000.0,
	}

	err := validator.ValidateOpenPosition(pos)
	if err != ErrInvalidTakeProfit {
		t.Errorf("Expected ErrInvalidTakeProfit, got %v", err)
	}
}

func TestValidateOpenPosition_InvalidStopLossRange_Long(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionLong,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   43000.0, // Stop loss above open price for long
		TakeProfit: 45000.0,
		Margin:     5000.0,
	}

	err := validator.ValidateOpenPosition(pos)
	if err == nil || err.Error() != ErrStopLossRange.Error()+": for long position, stop loss must be below open price" {
		t.Errorf("Expected stop loss range error, got %v", err)
	}
}

func TestValidateOpenPosition_InvalidTakeProfitRange_Long(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionLong,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   41000.0,
		TakeProfit: 42000.0, // Take profit below open price for long
		Margin:     5000.0,
	}

	err := validator.ValidateOpenPosition(pos)
	if err == nil || err.Error() != ErrTakeProfitRange.Error()+": for long position, take profit must be above open price" {
		t.Errorf("Expected take profit range error, got %v", err)
	}
}

func TestValidateOpenPosition_InvalidStopLossRange_Short(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Symbol:     "BTC/USDT",
		MarketType: models.MarketTypeCrypto,
		Direction:  models.DirectionShort,
		OpenPrice:  42500.0,
		Quantity:   0.5,
		StopLoss:   41000.0, // Stop loss below open price for short
		TakeProfit: 40000.0,
		Margin:     5000.0,
	}

	err := validator.ValidateOpenPosition(pos)
	if err == nil || err.Error() != ErrStopLossRange.Error()+": for short position, stop loss must be above open price" {
		t.Errorf("Expected stop loss range error, got %v", err)
	}
}

func TestValidateClosePosition_Success(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Status:   models.StatusOpen,
		Quantity: 1.0,
	}

	err := validator.ValidateClosePosition(pos, 0.5)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

func TestValidateClosePosition_AlreadyClosed(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Status:   models.StatusClosed,
		Quantity: 1.0,
	}

	err := validator.ValidateClosePosition(pos, 0.5)
	if err != ErrPositionAlreadyClosed {
		t.Errorf("Expected ErrPositionAlreadyClosed, got %v", err)
	}
}

func TestValidateClosePosition_ExceedsQuantity(t *testing.T) {
	validator := NewPositionValidator()

	pos := &models.Position{
		Status:   models.StatusOpen,
		Quantity: 1.0,
	}

	err := validator.ValidateClosePosition(pos, 1.5)
	if err == nil {
		t.Errorf("Expected error for close quantity exceeding position quantity")
	}
}
