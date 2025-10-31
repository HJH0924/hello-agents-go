package chapter04

import (
	"context"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestReflectionAgent_Run(t *testing.T) {
	tests := []struct {
		name string
		task string
	}{
		{
			name: "素数问题",
			task: "编写一个Python函数，找出1到n之间所有的素数 (prime numbers)。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 初始化 LLM 客户端
			llmClient := NewHelloAgentsLLM(
				openai.GPT4oMini,
				os.Getenv("OPENAI_API_KEY"),
				os.Getenv("OPENAI_BASE_URL"),
				60,
			)

			// 创建 Reflection Agent，设置最多迭代2轮
			agent := NewReflectionAgent(llmClient, 2)

			// 运行 agent
			ctx := context.Background()
			finalCode, err := agent.Run(ctx, tt.task)
			assert.NoError(t, err)
			assert.NotEmpty(t, finalCode)
		})
	}

}
