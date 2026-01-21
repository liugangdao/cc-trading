package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"trading-journal-cli/internal/models"
)

// Storage 存储接口
type Storage interface {
	AppendPosition(pos *models.Position) error
	ReadPositions(year int, month time.Month) ([]*models.Position, error)
	ReadAllPositions() ([]*models.Position, error)
	ReadOpenPositions() ([]*models.Position, error)
	UpdatePosition(pos *models.Position) error
	FindPositionByID(positionID string) (*models.Position, error)
}

// JSONLStorage JSONL文件存储
type JSONLStorage struct {
	dataDir string
}

// NewJSONLStorage 创建新的JSONL存储
func NewJSONLStorage(dataDir string) *JSONLStorage {
	return &JSONLStorage{dataDir: dataDir}
}

// ensureDataDir 确保数据目录存在
func (s *JSONLStorage) ensureDataDir() error {
	if _, err := os.Stat(s.dataDir); os.IsNotExist(err) {
		return os.MkdirAll(s.dataDir, 0755)
	}
	return nil
}

// getFilePath 获取指定月份的文件路径
func (s *JSONLStorage) getFilePath(year int, month time.Month) string {
	filename := fmt.Sprintf("trades-%04d-%02d.jsonl", year, month)
	return filepath.Join(s.dataDir, filename)
}

// AppendPosition 追加仓位记录
func (s *JSONLStorage) AppendPosition(pos *models.Position) error {
	if err := s.ensureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := s.getFilePath(pos.OpenTime.Year(), pos.OpenTime.Month())
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(pos)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write position: %w", err)
	}

	return nil
}

// ReadPositions 读取指定月份的所有仓位
func (s *JSONLStorage) ReadPositions(year int, month time.Month) ([]*models.Position, error) {
	filePath := s.getFilePath(year, month)

	// 文件不存在时返回空列表
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []*models.Position{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	positions := make(map[string]*models.Position)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var pos models.Position
		if err := json.Unmarshal(scanner.Bytes(), &pos); err != nil {
			// 跳过损坏的行，继续处理
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid JSON at line %d in %s: %v\n",
				lineNum, filePath, err)
			continue
		}
		// 同一 positionId 取最后一条记录
		positions[pos.PositionID] = &pos
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// 转换为切片
	result := make([]*models.Position, 0, len(positions))
	for _, pos := range positions {
		result = append(result, pos)
	}

	// 按开仓时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].OpenTime.Before(result[j].OpenTime)
	})

	return result, nil
}

// ReadAllPositions 读取所有月份的仓位
func (s *JSONLStorage) ReadAllPositions() ([]*models.Position, error) {
	if err := s.ensureDataDir(); err != nil {
		return nil, fmt.Errorf("failed to ensure data directory: %w", err)
	}

	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	allPositions := make(map[string]*models.Position)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		filePath := filepath.Join(s.dataDir, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open file %s: %v\n", entry.Name(), err)
			continue
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			var pos models.Position
			if err := json.Unmarshal(scanner.Bytes(), &pos); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: skipping invalid JSON at line %d in %s: %v\n",
					lineNum, entry.Name(), err)
				continue
			}
			allPositions[pos.PositionID] = &pos
		}

		file.Close()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error reading file %s: %v\n", entry.Name(), err)
		}
	}

	// 转换为切片
	result := make([]*models.Position, 0, len(allPositions))
	for _, pos := range allPositions {
		result = append(result, pos)
	}

	// 按开仓时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].OpenTime.Before(result[j].OpenTime)
	})

	return result, nil
}

// ReadOpenPositions 读取所有未平仓位
func (s *JSONLStorage) ReadOpenPositions() ([]*models.Position, error) {
	allPositions, err := s.ReadAllPositions()
	if err != nil {
		return nil, err
	}

	openPositions := make([]*models.Position, 0)
	for _, pos := range allPositions {
		if pos.Status == models.StatusOpen {
			openPositions = append(openPositions, pos)
		}
	}

	return openPositions, nil
}

// UpdatePosition 更新仓位（追加新版本）
func (s *JSONLStorage) UpdatePosition(pos *models.Position) error {
	return s.AppendPosition(pos)
}

// FindPositionByID 根据ID查找仓位
func (s *JSONLStorage) FindPositionByID(positionID string) (*models.Position, error) {
	allPositions, err := s.ReadAllPositions()
	if err != nil {
		return nil, err
	}

	for _, pos := range allPositions {
		if pos.PositionID == positionID {
			return pos, nil
		}
	}

	return nil, fmt.Errorf("position not found: %s", positionID)
}
