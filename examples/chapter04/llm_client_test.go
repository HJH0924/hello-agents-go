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
		name string
	}{
		{
			name: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := os.Getenv("LLM_MODEL_ID")
			if model == "" {
				model = "gpt-4o-mini"
			}
			apiKey := os.Getenv("OPENAI_API_KEY")
			baseUrl := os.Getenv("OPENAI_BASE_URL")
			timeout := 60
			llm := NewHelloAgentsLLM(model, apiKey, baseUrl, timeout)
			resp, err := llm.Think(context.Background(), []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant that writes Python code."},
				{Role: openai.ChatMessageRoleUser, Content: "写一个快速排序算法"},
			}, 0.5)
			assert.NoError(t, err)
			assert.NotEmpty(t, resp)
			fmt.Printf("\n\n--- 完整模型响应 ---\n\n")
			fmt.Println(resp)
		})
	}
}
