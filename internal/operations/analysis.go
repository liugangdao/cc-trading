package operations

import (
	"fmt"
	"time"
	"trading-journal-cli/internal/models"
)

// RiskReport 风险报告
type RiskReport struct {
	TotalMargin          float64
	MaxPossibleLoss      float64
	RiskExposurePercent  float64
	PositionCount        int
	PositionRisks        []PositionRisk
	ConcentrationByType  map[models.MarketType]float64
	ConcentrationBySymbol map[string]float64
	Warnings             []string
}

// PositionRisk 单个仓位风险
type PositionRisk struct {
	PositionID      string
	Symbol          string
	Direction       models.Direction
	Margin          float64
	PossibleLoss    float64
	RiskRewardRatio float64
}

// PerformanceReport 表现报告
type PerformanceReport struct {
	TotalTrades        int
	WinningTrades      int
	LosingTrades       int
	WinRate            float64
	TotalPnL           float64
	TotalPnLPercentage float64
	AveragePnL         float64
	BestTrade          *models.Position
	WorstTrade         *models.Position
	BySymbol           map[string]*SymbolStats
	ByMarketType       map[models.MarketType]*MarketTypeStats
	ByCloseReason      map[models.CloseReason]int
	AverageHoldingTime time.Duration
}

// SymbolStats 品种统计
type SymbolStats struct {
	Symbol        string
	TotalTrades   int
	WinningTrades int
	WinRate       float64
	TotalPnL      float64
	AveragePnL    float64
}

// MarketTypeStats 市场类型统计
type MarketTypeStats struct {
	MarketType    models.MarketType
	TotalTrades   int
	WinningTrades int
	WinRate       float64
	TotalPnL      float64
	AveragePnL    float64
}

// AnalyzeRisk 分析风险
func (o *Operations) AnalyzeRisk() (*RiskReport, error) {
	openPositions, err := o.storage.ReadOpenPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to read open positions: %w", err)
	}

	report := &RiskReport{
		PositionRisks:         make([]PositionRisk, 0),
		ConcentrationByType:   make(map[models.MarketType]float64),
		ConcentrationBySymbol: make(map[string]float64),
		Warnings:              make([]string, 0),
	}

	// 计算总保证金和最大可能损失
	for _, pos := range openPositions {
		report.TotalMargin += pos.Margin
		report.PositionCount++

		// 计算单个仓位的可能损失
		var possibleLoss float64
		if pos.Direction == models.DirectionLong {
			possibleLoss = (pos.OpenPrice - pos.StopLoss) * pos.Quantity
		} else {
			possibleLoss = (pos.StopLoss - pos.OpenPrice) * pos.Quantity
		}
		report.MaxPossibleLoss += possibleLoss

		// 计算风险回报比
		var riskRewardRatio float64
		if possibleLoss > 0 {
			var potentialProfit float64
			if pos.Direction == models.DirectionLong {
				potentialProfit = (pos.TakeProfit - pos.OpenPrice) * pos.Quantity
			} else {
				potentialProfit = (pos.OpenPrice - pos.TakeProfit) * pos.Quantity
			}
			riskRewardRatio = potentialProfit / possibleLoss
		}

		report.PositionRisks = append(report.PositionRisks, PositionRisk{
			PositionID:      pos.PositionID,
			Symbol:          pos.Symbol,
			Direction:       pos.Direction,
			Margin:          pos.Margin,
			PossibleLoss:    possibleLoss,
			RiskRewardRatio: riskRewardRatio,
		})

		// 统计仓位集中度
		report.ConcentrationByType[pos.MarketType] += pos.Margin
		report.ConcentrationBySymbol[pos.Symbol] += pos.Margin
	}

	// 生成风险预警
	if report.TotalMargin > 0 {
		report.RiskExposurePercent = (report.MaxPossibleLoss / report.TotalMargin) * 100

		// 检查单一品种集中度
		for symbol, margin := range report.ConcentrationBySymbol {
			concentration := (margin / report.TotalMargin) * 100
			if concentration > 40 {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("品种 %s 占比过高: %.2f%%", symbol, concentration))
			}
		}

		// 检查单一市场类型集中度
		for marketType, margin := range report.ConcentrationByType {
			concentration := (margin / report.TotalMargin) * 100
			if concentration > 40 {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("市场类型 %s 占比过高: %.2f%%", marketType, concentration))
			}
		}

		// 检查低风险回报比
		for _, pr := range report.PositionRisks {
			if pr.RiskRewardRatio < 2 {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("仓位 %s 风险回报比偏低: %.2f", pr.PositionID, pr.RiskRewardRatio))
			}
		}
	}

	return report, nil
}

// AnalyzePerformance 分析表现
func (o *Operations) AnalyzePerformance(fromDate, toDate time.Time) (*PerformanceReport, error) {
	// 读取所有已平仓位
	allPositions, err := o.storage.ReadAllPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to read positions: %w", err)
	}

	report := &PerformanceReport{
		BySymbol:      make(map[string]*SymbolStats),
		ByMarketType:  make(map[models.MarketType]*MarketTypeStats),
		ByCloseReason: make(map[models.CloseReason]int),
	}

	var totalHoldingSeconds int64

	for _, pos := range allPositions {
		// 只统计已平仓位
		if pos.Status != models.StatusClosed {
			continue
		}

		// 日期范围筛选
		if !fromDate.IsZero() && pos.OpenTime.Before(fromDate) {
			continue
		}
		if !toDate.IsZero() && pos.OpenTime.After(toDate) {
			continue
		}

		report.TotalTrades++
		pnl := *pos.RealizedPnL
		report.TotalPnL += pnl

		// 统计盈亏
		if pnl > 0 {
			report.WinningTrades++
		} else if pnl < 0 {
			report.LosingTrades++
		}

		// 记录最佳和最差交易
		if report.BestTrade == nil || pnl > *report.BestTrade.RealizedPnL {
			report.BestTrade = pos
		}
		if report.WorstTrade == nil || pnl < *report.WorstTrade.RealizedPnL {
			report.WorstTrade = pos
		}

		// 按品种统计
		if _, exists := report.BySymbol[pos.Symbol]; !exists {
			report.BySymbol[pos.Symbol] = &SymbolStats{Symbol: pos.Symbol}
		}
		stats := report.BySymbol[pos.Symbol]
		stats.TotalTrades++
		if pnl > 0 {
			stats.WinningTrades++
		}
		stats.TotalPnL += pnl

		// 按市场类型统计
		if _, exists := report.ByMarketType[pos.MarketType]; !exists {
			report.ByMarketType[pos.MarketType] = &MarketTypeStats{MarketType: pos.MarketType}
		}
		mtStats := report.ByMarketType[pos.MarketType]
		mtStats.TotalTrades++
		if pnl > 0 {
			mtStats.WinningTrades++
		}
		mtStats.TotalPnL += pnl

		// 按平仓原因统计
		if pos.CloseReason != nil {
			report.ByCloseReason[*pos.CloseReason]++
		}

		// 统计持仓时长
		if pos.CloseTime != nil {
			duration := pos.CloseTime.Sub(pos.OpenTime)
			totalHoldingSeconds += int64(duration.Seconds())
		}
	}

	// 计算平均值和比率
	if report.TotalTrades > 0 {
		report.WinRate = float64(report.WinningTrades) / float64(report.TotalTrades) * 100
		report.AveragePnL = report.TotalPnL / float64(report.TotalTrades)
		report.AverageHoldingTime = time.Duration(totalHoldingSeconds/int64(report.TotalTrades)) * time.Second

		for _, stats := range report.BySymbol {
			if stats.TotalTrades > 0 {
				stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
				stats.AveragePnL = stats.TotalPnL / float64(stats.TotalTrades)
			}
		}

		for _, stats := range report.ByMarketType {
			if stats.TotalTrades > 0 {
				stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
				stats.AveragePnL = stats.TotalPnL / float64(stats.TotalTrades)
			}
		}
	}

	return report, nil
}
