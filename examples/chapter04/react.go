package chapter04

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const ReactPromptTemplate = `
请注意，你是一个有能力调用外部工具的智能助手。

可用工具如下：
%s

请严格按照以下格式进行回应：

Thought: 你的思考过程，用于分析问题、拆解任务和规划下一步行动。
Action: 你决定采取的行动，必须是以下格式之一：
- '{tool_name}[{tool_input}]'：调用一个可用工具。
- 'Finish[最终答案]'：当你认为已经获得最终答案时。

现在，请开始解决以下问题：
Question: {%s}
History: {%s}
`

type ReactAgent struct {
	llmClient    *HelloAgentsLLM
	toolExecutor *ToolExecutor
	maxSteps     int
	history      []string
}

func NewReactAgent(llmClient *HelloAgentsLLM, toolExecutor *ToolExecutor, maxSteps int) *ReactAgent {
	return &ReactAgent{
		llmClient:    llmClient,
		toolExecutor: toolExecutor,
		maxSteps:     maxSteps,
		history:      []string{},
	}
}

func (a *ReactAgent) Run(ctx context.Context, question string) (string, error) {
	a.history = []string{} // 每次运行时重置历史记录
	currentStep := 0

	for currentStep < a.maxSteps {
		currentStep++
		fmt.Printf("\n--- 第 %d 步 ---\n", currentStep)

		// 格式化提示词
		toolsDesc := a.toolExecutor.GetAvailableTools()
		historyStr := strings.Join(a.history, "\n")
		prompt := fmt.Sprintf(ReactPromptTemplate, toolsDesc, question, historyStr)
		fmt.Printf("prompt: %s\n", prompt)

		// 调用 LLM 进行思考
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		}
		response, err := a.llmClient.Think(ctx, messages, 0.5)
		if err != nil {
			return "", err
		}

		if response == "" {
			fmt.Println("错误：LLM未能返回有效响应。")
			break
		}

		thought, action := a.parseOutput(response)
		if thought != "" {
			fmt.Printf("🤔 思考: %s\n", thought)
		}
		if action == "" {
			fmt.Println("警告：未能解析出有效的Action，流程终止。")
			break
		}

		// 检查是否是 Finish 动作
		if strings.HasPrefix(action, "Finish") {
			finalAnswer := a.parseActionInput(action)
			fmt.Printf("🎉 最终答案: %s\n", finalAnswer)
			return finalAnswer, nil
		}

		// 解析工具调用
		toolName, toolInput := a.parseAction(action)
		if toolName == "" || toolInput == "" {
			a.history = append(a.history, "Observation: 无效的Action格式，请检查。")
			continue
		}

		fmt.Printf("🎬 行动: %s[%s]\n", toolName, toolInput)
		toolCall, err := a.toolExecutor.GetToolCall(toolName)
		var observation string
		if err != nil {
			observation = fmt.Sprintf("错误：未找到名为 '%s' 的工具。", toolName)
		} else {
			result, err := toolCall(toolInput)
			if err != nil {
				observation = fmt.Sprintf("工具执行错误: %v", err)
			} else {
				observation = result
			}
		}

		fmt.Printf("👀 观察: %s\n", observation)
		a.history = append(a.history, fmt.Sprintf("Action: %s", action))
		a.history = append(a.history, fmt.Sprintf("Observation: %s", observation))
	}

	fmt.Println("已达到最大步数，流程终止。")
	return "", nil
}

// parseOutput 从 LLM 响应中解析 Thought 和 Action
func (a *ReactAgent) parseOutput(text string) (string, string) {
	thoughtRegex := regexp.MustCompile(`Thought: (.*)`)
	actionRegex := regexp.MustCompile(`Action: (.*)`)

	var thought, action string

	if thoughtMatch := thoughtRegex.FindStringSubmatch(text); len(thoughtMatch) > 1 {
		thought = strings.TrimSpace(thoughtMatch[1])
	}

	if actionMatch := actionRegex.FindStringSubmatch(text); len(actionMatch) > 1 {
		action = strings.TrimSpace(actionMatch[1])
	}

	return thought, action
}

// parseAction 从 Action 文本中解析工具名称和输入
func (a *ReactAgent) parseAction(actionText string) (string, string) {
	actionRegex := regexp.MustCompile(`(\w+)\[(.*)\]`)
	if match := actionRegex.FindStringSubmatch(actionText); len(match) > 2 {
		return match[1], match[2]
	}
	return "", ""
}

// parseActionInput 从 Action 文本中解析输入内容
func (a *ReactAgent) parseActionInput(actionText string) string {
	actionInputRegex := regexp.MustCompile(`\w+\[(.*)\]`)
	if match := actionInputRegex.FindStringSubmatch(actionText); len(match) > 1 {
		return match[1]
	}
	return ""
}
