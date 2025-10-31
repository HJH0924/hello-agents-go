package chapter04

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestHelloAgentsLLM_Think(t *testing.T) {
	tests := []struct {
		name             string
		model            string
		messages         []openai.ChatCompletionMessage
		expectedResponse string
	}{
		{
			name:  "快速排序",
			model: openai.GPT4oMini,
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant that writes Python code."},
				{Role: openai.ChatMessageRoleUser, Content: "写一个快速排序算法"},
			},
		},
		{
			name:  "不使用 plan and solve",
			model: openai.GPT3Dot5Turbo,
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "一个水果店周一卖出了15个苹果。周二卖出的苹果数量是周一的两倍。周三卖出的数量比周二少了5个。请问这三天总共卖出了多少个苹果？"},
			},
			// 15 + 15 * 2 + 15 * 2 - 5 = 70
			expectedResponse: "70",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := os.Getenv("OPENAI_API_KEY")
			baseUrl := os.Getenv("OPENAI_BASE_URL")
			timeout := 60
			llm := NewHelloAgentsLLM(tt.model, apiKey, baseUrl, timeout)
			resp, err := llm.Think(context.Background(), tt.messages, 0.5)
			assert.NoError(t, err)
			assert.NotEmpty(t, resp)
			fmt.Printf("\n\n--- 完整模型响应 ---\n\n")
			fmt.Println(resp)
		})
	}
}
