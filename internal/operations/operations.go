package operations

import (
	"fmt"
	"time"
	"trading-journal-cli/internal/models"
	"trading-journal-cli/internal/storage"
	"trading-journal-cli/internal/validator"
)

// OpenParams 开仓参数
type OpenParams struct {
	Symbol     string
	MarketType models.MarketType
	Direction  models.Direction
	OpenPrice  float64
	Quantity   float64
	StopLoss   float64
	TakeProfit float64
	Margin     float64
	Reason     string
	OpenTime   *time.Time // 可选，为空时使用当前时间
}

// CloseParams 平仓参数
type CloseParams struct {
	ClosePrice    float64
	CloseQuantity float64
	CloseReason   models.CloseReason
	CloseNote     string
	CloseTime     *time.Time // 可选，为空时使用当前时间
}

// FilterParams 筛选参数
type FilterParams struct {
	Status     string    // "open", "closed", "all"
	Symbol     string    // 为空则不筛选
	MarketType string    // 为空则不筛选
	FromDate   time.Time // 零值则不筛选
	ToDate     time.Time // 零值则不筛选
}

// Operations 操作接口
type Operations struct {
	storage   storage.Storage
	validator validator.Validator
}

// NewOperations 创建新的操作实例
func NewOperations(store storage.Storage, valid validator.Validator) *Operations {
	return &Operations{
		storage:   store,
		validator: valid,
	}
}

// OpenPosition 开仓操作
func (o *Operations) OpenPosition(params OpenParams) (*models.Position, error) {
	// 设置开仓时间
	openTime := time.Now()
	if params.OpenTime != nil {
		openTime = *params.OpenTime
	}

	// 创建仓位对象
	pos := &models.Position{
		PositionID: models.GeneratePositionID(),
		Symbol:     params.Symbol,
		MarketType: params.MarketType,
		OpenTime:   openTime,
		Direction:  params.Direction,
		OpenPrice:  params.OpenPrice,
		Quantity:   params.Quantity,
		StopLoss:   params.StopLoss,
		TakeProfit: params.TakeProfit,
		Margin:     params.Margin,
		Reason:     params.Reason,
		Status:     models.StatusOpen,
	}

	// 验证数据
	if err := o.validator.ValidateOpenPosition(pos); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 保存到存储
	if err := o.storage.AppendPosition(pos); err != nil {
		return nil, fmt.Errorf("failed to save position: %w", err)
	}

	return pos, nil
}

// ClosePosition 平仓操作
func (o *Operations) ClosePosition(positionID string, params CloseParams) (*models.Position, error) {
	// 查找仓位
	pos, err := o.storage.FindPositionByID(positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find position: %w", err)
	}

	// 验证平仓数据
	if err := o.validator.ValidateClosePosition(pos, params.CloseQuantity); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 设置平仓时间
	closeTime := time.Now()
	if params.CloseTime != nil {
		closeTime = *params.CloseTime
	}

	// 计算盈亏
	realizedPnL := models.CalculateRealizedPnL(pos.Direction, pos.OpenPrice, params.ClosePrice, params.CloseQuantity)
	pnlPercentage := models.CalculatePnLPercentage(realizedPnL, pos.Margin)
	holdingDuration := models.FormatHoldingDuration(closeTime.Sub(pos.OpenTime))

	// 更新仓位信息
	pos.Status = models.StatusClosed
	pos.CloseTime = &closeTime
	pos.ClosePrice = &params.ClosePrice
	pos.CloseQuantity = &params.CloseQuantity
	pos.RealizedPnL = &realizedPnL
	pos.PnLPercentage = &pnlPercentage
	pos.HoldingDuration = &holdingDuration
	pos.CloseReason = &params.CloseReason
	pos.CloseNote = params.CloseNote

	// 保存更新后的记录
	if err := o.storage.UpdatePosition(pos); err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return pos, nil
}

// ListPositions 列出仓位
func (o *Operations) ListPositions(filter FilterParams) ([]*models.Position, error) {
	// 读取所有仓位
	allPositions, err := o.storage.ReadAllPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to read positions: %w", err)
	}

	// 应用筛选条件
	result := make([]*models.Position, 0)
	for _, pos := range allPositions {
		// 状态筛选
		if filter.Status != "" && filter.Status != "all" {
			if filter.Status == "open" && pos.Status != models.StatusOpen {
				continue
			}
			if filter.Status == "closed" && pos.Status != models.StatusClosed {
				continue
			}
		}

		// 品种筛选
		if filter.Symbol != "" && pos.Symbol != filter.Symbol {
			continue
		}

		// 市场类型筛选
		if filter.MarketType != "" && string(pos.MarketType) != filter.MarketType {
			continue
		}

		// 日期范围筛选
		if !filter.FromDate.IsZero() && pos.OpenTime.Before(filter.FromDate) {
			continue
		}
		if !filter.ToDate.IsZero() && pos.OpenTime.After(filter.ToDate) {
			continue
		}

		result = append(result, pos)
	}

	return result, nil
}

// GetOpenPositions 获取所有未平仓位
func (o *Operations) GetOpenPositions() ([]*models.Position, error) {
	return o.storage.ReadOpenPositions()
}
