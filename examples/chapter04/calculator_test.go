package chapter04

import (
	"context"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestCalculatorTool(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       string
		wantErr    bool
	}{
		{
			name:       "简单加法",
			expression: "123 + 456",
			want:       "579",
			wantErr:    false,
		},
		{
			name:       "复杂表达式",
			expression: "(123 + 456) * 789 / 12",
			want:       "38069.25",
			wantErr:    false,
		},
		{
			name:       "幂运算",
			expression: "2 ** 10",
			want:       "1024",
			wantErr:    false,
		},
		{
			name:       "错误表达式",
			expression: "123 + + 456",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculatorTool(tt.expression)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestReactAgentWithCalculator(t *testing.T) {
	// 初始化工具执行器
	toolExecutor := NewToolExecutor()

	// 注册计算器工具
	RegisterCalculatorTool(toolExecutor)

	// 注册搜索工具（可选）
	RegisterGoogleSearchTool(toolExecutor)

	// 初始化LLM客户端
	llmClient := NewHelloAgentsLLM(
		openai.GPT4oMini,
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("OPENAI_BASE_URL"),
		60,
	)

	// 创建 ReAct Agent
	agent := NewReactAgent(llmClient, toolExecutor, 5)

	// 测试问题
	question := "计算 (123 + 456) × 789 / 12 的结果是多少？"

	// 运行
	ctx := context.Background()
	answer, err := agent.Run(ctx, question)

	assert.NoError(t, err)
	assert.NotEmpty(t, answer)
	assert.Contains(t, answer, "38069") // 验证答案中包含正确结果

	t.Logf("问题: %s", question)
	t.Logf("答案: %s", answer)
}
