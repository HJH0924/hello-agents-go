package chapter04

import (
	"context"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestPlanAndSolveAgent_Run(t *testing.T) {
	tests := []struct {
		name           string
		question       string
		expectedAnswer string
	}{
		{
			name:           "水果店问题",
			question:       "一个水果店周一卖出了15个苹果。周二卖出的苹果数量是周一的两倍。周三卖出的数量比周二少了5个。请问这三天总共卖出了多少个苹果？",
			expectedAnswer: "70",
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

			// 创建 Plan and Solve Agent
			agent := NewPlanAndSolveAgent(llmClient)

			// 运行 agent
			ctx := context.Background()
			answer, err := agent.Run(ctx, tt.question)
			assert.NoError(t, err)
			assert.NotEmpty(t, answer)

			t.Logf("期望答案：%s", tt.expectedAnswer)
			t.Logf("最终答案: %s", answer)
		})
	}

}
