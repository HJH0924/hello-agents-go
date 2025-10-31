package chapter04

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolExecutor(t *testing.T) {
	tests := []struct {
		name            string
		toolName        string
		toolDescription string
		toolInput       string
		toolCall        func(string) (string, error)
	}{
		{
			name:            "GoogleSearch",
			toolName:        "GoogleSearch",
			toolDescription: "一个网页搜索引擎。当你需要回答关于时事、事实以及在你的知识库中找不到的信息时，应使用此工具。",
			toolInput:       "英伟达最新的GPU型号是什么",
			toolCall:        GoogleSearch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolExecutor := NewToolExecutor()
			toolExecutor.RegisterTool(tt.toolName, tt.toolDescription, tt.toolCall)

			fmt.Println("\n--- 可用的工具 ---")
			fmt.Println(toolExecutor.GetAvailableTools())

			fmt.Println("\n--- 调用工具 ---")
			fmt.Printf("工具名: %s\n", tt.toolName)
			fmt.Printf("输入: %s\n\n", tt.toolInput)

			toolCall, err := toolExecutor.GetToolCall(tt.toolName)
			assert.NoError(t, err)
			observation, err := toolCall(tt.toolInput)
			assert.NoError(t, err)

			fmt.Println("\n--- 观察 (Observation) ---")
			fmt.Println(observation)
		})
	}
}
