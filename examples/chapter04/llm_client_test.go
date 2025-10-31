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

const ReflectPromptTemplate_Strict = `
你是一位极其严格的代码评审专家和资深算法工程师，对代码的性能有极致的要求。
你的任务是审查以下Go代码，并专注于找出其在**算法效率**上的主要瓶颈。

# 原始任务:
%s

# 待审查的代码:
%s

请分析该代码的时间复杂度，并思考是否存在一种**算法上更优**的解决方案来显著提升性能。
如果存在，请清晰地指出当前算法的不足，并提出具体的、可行的改进算法建议。
如果代码在算法层面已经达到最优，才能回答"无需改进"。

请直接输出你的反馈，不要包含任何额外的解释。
`

const ReflectPromptTemplate_Readable = `
你是一位注重代码可读性的开源项目维护者，你的主要目标是让代码易于理解和维护。
你的任务是审查以下Go代码，并专注于找出其在**可读性和可维护性**上的改进空间。

# 原始任务:
%s

# 待审查的代码:
%s

请从以下角度评估代码：
1. 变量和函数命名是否清晰？
2. 代码逻辑是否容易理解？
3. 是否需要添加注释来解释复杂逻辑？
4. 代码结构是否清晰？是否可以进一步模块化？
5. 是否遵循Go的最佳实践和风格指南（Effective Go）？

如果代码已经足够清晰易读，才能回答"无需改进"。

请直接输出你的反馈。
`

func TestHelloAgentsLLM_Think_Reflect(t *testing.T) {
	tests := []struct {
		name           string
		promptTemplate string
	}{
		{
			name:           "严格的代码评审专家",
			promptTemplate: ReflectPromptTemplate_Strict,
		},
		{
			name:           "注重可读性的开源项目维护者",
			promptTemplate: ReflectPromptTemplate_Readable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := os.Getenv("OPENAI_API_KEY")
			baseUrl := os.Getenv("OPENAI_BASE_URL")
			timeout := 60
			llm := NewHelloAgentsLLM(openai.GPT4oMini, apiKey, baseUrl, timeout)
			task := "评审一个计算斐波那契数列的函数"
			code := `func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}`
			content := fmt.Sprintf(tt.promptTemplate, task, code)
			messages := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: content},
			}
			resp, err := llm.Think(context.Background(), messages, 0.5)
			assert.NoError(t, err)
			assert.NotEmpty(t, resp)
		})
	}
}
