package models

import (
	"testing"
	"time"
)

func TestGeneratePositionID(t *testing.T) {
	// 测试生成的ID格式
	id := GeneratePositionID()
	if len(id) != 20 { // YYYYMMDD-HHMMSS-XXXX = 8+1+6+1+4 = 20
		t.Errorf("Position ID length should be 20, got %d: %s", len(id), id)
	}

	// 测试唯一性 - 生成较少的ID并增加延迟
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := GeneratePositionID()
		if ids[id] {
			t.Errorf("Duplicate position ID generated: %s", id)
		}
		ids[id] = true
		time.Sleep(10 * time.Millisecond) // 增加延迟避免秒级冲突
	}
}

func TestCalculateRealizedPnL(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		openPrice float64
		closePrice float64
		quantity  float64
		expected  float64
	}{
		{
			name:       "Long position profit",
			direction:  DirectionLong,
			openPrice:  100.0,
			closePrice: 110.0,
			quantity:   1.0,
			expected:   10.0,
		},
		{
			name:       "Long position loss",
			direction:  DirectionLong,
			openPrice:  100.0,
			closePrice: 95.0,
			quantity:   1.0,
			expected:   -5.0,
		},
		{
			name:       "Short position profit",
			direction:  DirectionShort,
			openPrice:  100.0,
			closePrice: 90.0,
			quantity:   1.0,
			expected:   10.0,
		},
		{
			name:       "Short position loss",
			direction:  DirectionShort,
			openPrice:  100.0,
			closePrice: 105.0,
			quantity:   1.0,
			expected:   -5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRealizedPnL(tt.direction, tt.openPrice, tt.closePrice, tt.quantity)
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestCalculatePnLPercentage(t *testing.T) {
	tests := []struct {
		name        string
		realizedPnL float64
		margin      float64
		expected    float64
	}{
		{
			name:        "Positive PnL",
			realizedPnL: 100.0,
			margin:      1000.0,
			expected:    10.0,
		},
		{
			name:        "Negative PnL",
			realizedPnL: -50.0,
			margin:      1000.0,
			expected:    -5.0,
		},
		{
			name:        "Zero margin",
			realizedPnL: 100.0,
			margin:      0.0,
			expected:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePnLPercentage(tt.realizedPnL, tt.margin)
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestFormatHoldingDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Days, hours, minutes",
			duration: 2*24*time.Hour + 5*time.Hour + 30*time.Minute,
			expected: "2d 5h 30m",
		},
		{
			name:     "Hours and minutes",
			duration: 5*time.Hour + 30*time.Minute,
			expected: "5h 30m",
		},
		{
			name:     "Only minutes",
			duration: 30 * time.Minute,
			expected: "30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHoldingDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
