package chapter04

import (
	"fmt"
	"strings"
)

type RecordType string

const (
	RecordTypeExecution  RecordType = "execution"
	RecordTypeReflection RecordType = "reflection"
)

// Record 表示一条记忆记录
type Record struct {
	Type    RecordType
	Content string // 记录的内容
}

// Memory 短期记忆模块，用于存储智能体的行动与反思轨迹
type Memory struct {
	records []Record
}

// NewMemory 创建一个新的记忆实例
func NewMemory() *Memory {
	return &Memory{
		records: []Record{},
	}
}

// AddRecord 向记忆中添加一条新记录
func (m *Memory) AddRecord(recordType RecordType, content string) {
	m.records = append(m.records, Record{
		Type:    recordType,
		Content: content,
	})
	fmt.Printf("📝 记忆已更新，新增一条 '%s' 记录。\n", recordType)
}

// GetTrajectory 将所有记忆记录格式化为一个连贯的字符串文本
func (m *Memory) GetTrajectory() string {
	var trajectory strings.Builder

	for _, record := range m.records {
		switch record.Type {
		case RecordTypeExecution:
			trajectory.WriteString(fmt.Sprintf("--- 上一轮尝试 (代码) ---\n%s\n\n", record.Content))
		case RecordTypeReflection:
			trajectory.WriteString(fmt.Sprintf("--- 评审员反馈 ---\n%s\n\n", record.Content))
		}
	}

	return strings.TrimSpace(trajectory.String())
}

// GetLastExecution 获取最近一次的执行结果
func (m *Memory) GetLastExecution() string {
	// 从后往前遍历，找到最近的 execution 记录
	for i := len(m.records) - 1; i >= 0; i-- {
		if m.records[i].Type == RecordTypeExecution {
			return m.records[i].Content
		}
	}
	return ""
}
