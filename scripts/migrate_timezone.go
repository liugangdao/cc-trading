package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	dataDir := "./trading-data"

	// 读取所有 JSONL 文件
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		filePath := filepath.Join(dataDir, entry.Name())
		fmt.Printf("Processing: %s\n", entry.Name())

		if err := migrateFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", entry.Name(), err)
			continue
		}

		fmt.Printf("✓ Migrated: %s\n", entry.Name())
	}

	fmt.Println("\n✓ Migration completed!")
}

func migrateFile(filePath string) error {
	// 读取所有行到内存
	lines, err := readLines(filePath)
	if err != nil {
		return err
	}

	// 创建备份
	backupPath := filePath + ".backup"
	if err := copyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("  Created backup: %s\n", backupPath)

	// 写入新文件
	if err := writeLines(filePath, lines); err != nil {
		return fmt.Errorf("failed to write new file: %w", err)
	}

	return nil
}

func readLines(filePath string) ([]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []map[string]interface{}
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var data map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid JSON at line %d: %v\n", lineNum, err)
			continue
		}

		// 转换时间字段
		convertTimeField(data, "openTime")
		convertTimeField(data, "closeTime")

		lines = append(lines, data)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func writeLines(filePath string, lines []map[string]interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		data, err := json.Marshal(line)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func convertTimeField(data map[string]interface{}, fieldName string) {
	if timeStr, ok := data[fieldName].(string); ok && timeStr != "" {
		// 解析时间
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, timeStr)
			if err != nil {
				return
			}
		}

		// 转换为本地时区
		localTime := t.Local()

		// 重新格式化为 RFC3339Nano（保持本地时区）
		data[fieldName] = localTime.Format(time.RFC3339Nano)
	}
}
