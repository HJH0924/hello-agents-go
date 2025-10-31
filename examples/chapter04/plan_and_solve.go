package chapter04

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// PlannerPromptTemplate 规划器提示词模板
const PlannerPromptTemplate = `
你是一个顶级的AI规划专家。你的任务是将用户提出的复杂问题分解成一个由多个简单步骤组成的行动计划。
请确保计划中的每个步骤都是一个独立的、可执行的子任务，并且严格按照逻辑顺序排列。
你的输出必须是一个JSON数组，其中每个元素都是一个描述子任务的字符串。

问题: %s

请严格按照以下格式输出你的计划，` + "```json与```" + `作为前后缀是必要的:
` + "```json" + `
["步骤1", "步骤2", "步骤3", ...]
` + "```" + `
`

// ExecutorPromptTemplate 执行器提示词模板
const ExecutorPromptTemplate = `
你是一位顶级的AI执行专家。你的任务是严格按照给定的计划，一步步地解决问题。
你将收到原始问题、完整的计划、以及到目前为止已经完成的步骤和结果。
请你专注于解决"当前步骤"，并仅输出该步骤的最终答案，不要输出任何额外的解释或对话。

# 原始问题:
%s

# 完整计划:
%v

# 历史步骤与结果:
%s

# 当前步骤:
%s

请仅输出针对"当前步骤"的回答:
`

// Planner 规划器，负责将复杂问题分解成多个步骤
type Planner struct {
	llmClient *HelloAgentsLLM
}

func NewPlanner(llmClient *HelloAgentsLLM) *Planner {
	return &Planner{llmClient: llmClient}
}

// Plan 生成问题的行动计划
func (p *Planner) Plan(ctx context.Context, question string) ([]string, error) {
	prompt := fmt.Sprintf(PlannerPromptTemplate, question)
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}

	fmt.Println("\n--- 正在生成计划 ---")
	response, err := p.llmClient.Think(ctx, messages, 0.7)
	if err != nil {
		return nil, fmt.Errorf("生成计划失败: %w", err)
	}

	fmt.Println("✅ 计划已生成")

	// 解析计划
	plan, err := p.parsePlan(response)
	if err != nil {
		fmt.Printf("❌ 解析计划时出错: %v\n", err)
		fmt.Printf("原始响应: %s\n", response)
		return nil, err
	}

	return plan, nil
}

// parsePlan 从 LLM 响应中解析计划列表
func (p *Planner) parsePlan(response string) ([]string, error) {
	// 提取 ```json ... ``` 之间的内容
	jsonRegex := regexp.MustCompile("```(?:json)?\\s*\\n?([\\s\\S]*?)```")
	matches := jsonRegex.FindStringSubmatch(response)

	var planStr string
	if len(matches) > 1 {
		planStr = strings.TrimSpace(matches[1])
	} else {
		// 如果没有代码块标记，尝试直接解析
		planStr = strings.TrimSpace(response)
	}

	// 解析 JSON 数组
	var plan []string
	err := json.Unmarshal([]byte(planStr), &plan)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if len(plan) == 0 {
		return nil, fmt.Errorf("解析出的计划为空")
	}

	return plan, nil
}

// Executor 执行器，负责执行计划中的每个步骤
type Executor struct {
	llmClient *HelloAgentsLLM
}

func NewExecutor(llmClient *HelloAgentsLLM) *Executor {
	return &Executor{llmClient: llmClient}
}

// Execute 执行完整的计划并返回最终答案
func (e *Executor) Execute(ctx context.Context, question string, plan []string) (string, error) {
	history := ""
	finalAnswer := ""

	fmt.Println("\n--- 正在执行计划 ---")
	for i, step := range plan {
		fmt.Printf("\n-> 正在执行步骤 %d/%d: %s\n", i+1, len(plan), step)

		historyDisplay := history
		if historyDisplay == "" {
			historyDisplay = "无"
		}

		prompt := fmt.Sprintf(ExecutorPromptTemplate, question, plan, historyDisplay, step)
		fmt.Printf("\n\nprompt: %s\n\n", prompt)
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		}

		response, err := e.llmClient.Think(ctx, messages, 0.5)
		if err != nil {
			return "", fmt.Errorf("执行步骤 %d 失败: %w", i+1, err)
		}

		history += fmt.Sprintf("步骤 %d: %s\n结果: %s\n\n", i+1, step, response)
		finalAnswer = response
		fmt.Printf("✅ 步骤 %d 已完成，结果: %s\n", i+1, finalAnswer)
	}

	return finalAnswer, nil
}

// PlanAndSolveAgent Plan and Solve 智能体
type PlanAndSolveAgent struct {
	llmClient *HelloAgentsLLM
	planner   *Planner
	executor  *Executor
}

func NewPlanAndSolveAgent(llmClient *HelloAgentsLLM) *PlanAndSolveAgent {
	return &PlanAndSolveAgent{
		llmClient: llmClient,
		planner:   NewPlanner(llmClient),
		executor:  NewExecutor(llmClient),
	}
}

// Run 运行 Plan and Solve 流程
func (a *PlanAndSolveAgent) Run(ctx context.Context, question string) (string, error) {
	fmt.Printf("\n--- 开始处理问题 ---\n问题: %s\n", question)

	// 生成计划
	plan, err := a.planner.Plan(ctx, question)
	if err != nil {
		fmt.Println("\n--- 任务终止 ---\n无法生成有效的行动计划。")
		return "", err
	}

	// 执行计划
	finalAnswer, err := a.executor.Execute(ctx, question, plan)
	if err != nil {
		return "", err
	}

	fmt.Printf("\n--- 任务完成 ---\n最终答案: %s\n", finalAnswer)
	return finalAnswer, nil
}
