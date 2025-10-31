package chapter04

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// 初始执行提示词模板
const InitialPromptTemplate = `
你是一位资深的Python程序员。请根据以下要求，编写一个Python函数。
你的代码必须包含完整的函数签名、文档字符串，并遵循PEP 8编码规范。

要求: %s

请直接输出代码，不要包含任何额外的解释。
`

// 反思提示词模板
const ReflectPromptTemplate = `
你是一位极其严格的代码评审专家和资深算法工程师，对代码的性能有极致的要求。
你的任务是审查以下Python代码，并专注于找出其在**算法效率**上的主要瓶颈。

# 原始任务:
%s

# 待审查的代码:
%s

请分析该代码的时间复杂度，并思考是否存在一种**算法上更优**的解决方案来显著提升性能。
如果存在，请清晰地指出当前算法的不足，并提出具体的、可行的改进算法建议（例如，使用筛法替代试除法）。
如果代码在算法层面已经达到最优，才能回答"无需改进"。

请直接输出你的反馈，不要包含任何额外的解释。
`

// 优化提示词模板
const RefinePromptTemplate = `
你是一位资深的Python程序员。你正在根据一位代码评审专家的反馈来优化你的代码。

# 原始任务:
%s

# 你上一轮尝试的代码:
%s

# 评审员的反馈:
%s

请根据评审员的反馈，生成一个优化后的新版本代码。
你的代码必须包含完整的函数签名、文档字符串，并遵循PEP 8编码规范。
请直接输出优化后的代码，不要包含任何额外的解释。
`

// ReflectionAgent Reflection 智能体，通过反思和优化迭代改进代码
type ReflectionAgent struct {
	llmClient     *HelloAgentsLLM
	memory        *Memory
	maxIterations int
}

// NewReflectionAgent 创建一个新的 Reflection Agent
func NewReflectionAgent(llmClient *HelloAgentsLLM, maxIterations int) *ReflectionAgent {
	return &ReflectionAgent{
		llmClient:     llmClient,
		memory:        NewMemory(),
		maxIterations: maxIterations,
	}
}

// Run 运行 Reflection 流程
func (a *ReflectionAgent) Run(ctx context.Context, task string) (string, error) {
	fmt.Printf("\n--- 开始处理任务 ---\n任务: %s\n", task)

	// 1. 初始执行
	fmt.Println("\n--- 正在进行初始尝试 ---")
	initialPrompt := fmt.Sprintf(InitialPromptTemplate, task)
	initialCode, err := a.getLLMResponse(ctx, initialPrompt)
	if err != nil {
		return "", fmt.Errorf("初始执行失败: %w", err)
	}
	a.memory.AddRecord(RecordTypeExecution, initialCode)

	// 2. 迭代循环：反思与优化
	for i := 0; i < a.maxIterations; i++ {
		fmt.Printf("\n--- 第 %d/%d 轮迭代 ---\n", i+1, a.maxIterations)

		// a. 反思
		fmt.Println("\n-> 正在进行反思...")
		lastCode := a.memory.GetLastExecution()
		reflectPrompt := fmt.Sprintf(ReflectPromptTemplate, task, lastCode)
		feedback, err := a.getLLMResponse(ctx, reflectPrompt)
		if err != nil {
			return "", fmt.Errorf("反思失败: %w", err)
		}
		a.memory.AddRecord(RecordTypeReflection, feedback)

		// b. 检查是否需要停止
		if strings.Contains(feedback, "无需改进") || strings.Contains(strings.ToLower(feedback), "no need for improvement") {
			fmt.Println("\n✅ 反思认为代码已无需改进，任务完成。")
			break
		}

		// c. 优化
		fmt.Println("\n-> 正在进行优化...")
		refinePrompt := fmt.Sprintf(RefinePromptTemplate, task, lastCode, feedback)
		refinedCode, err := a.getLLMResponse(ctx, refinePrompt)
		if err != nil {
			return "", fmt.Errorf("优化失败: %w", err)
		}
		a.memory.AddRecord(RecordTypeExecution, refinedCode)
	}

	finalCode := a.memory.GetLastExecution()
	fmt.Printf("\n--- 任务完成 ---\n最终生成的代码:\n```python\n%s\n```\n", finalCode)
	return finalCode, nil
}

// getLLMResponse 辅助方法，用于调用 LLM 并获取完整响应
func (a *ReflectionAgent) getLLMResponse(ctx context.Context, prompt string) (string, error) {
	fmt.Printf("\n\nprompt: %s\n\n", prompt)
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}
	return a.llmClient.Think(ctx, messages, 0.7)
}
