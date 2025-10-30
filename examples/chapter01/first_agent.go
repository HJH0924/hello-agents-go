package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/iamwavecut/go-tavily"
	openai "github.com/sashabaranov/go-openai"
)

const agentSystemPrompt = `你是一个智能旅行助手。你的任务是分析用户的请求，并使用可用工具一步步地解决问题。

# 可用工具:
- 'get_weather(city: str)': 查询指定城市的实时天气。
- 'get_attraction(city: str, weather: str)': 根据城市和天气搜索推荐的旅游景点。

# 行动格式:
你的回答必须严格遵循以下格式。首先是你的思考过程，然后是你要执行的具体行动。
Thought: [这里是你的思考过程和下一步计划]
Action: [这里是你要调用的工具，格式为 function_name(arg_name="arg_value")]

# 任务完成:
当你收集到足够的信息，能够回答用户的最终问题时，你必须使用 finish(answer="...") 来输出最终答案。

请开始吧！
`

// WeatherResponse represents the weather API response structure
type WeatherResponse struct {
	CurrentCondition []struct {
		TempC       string `json:"temp_C"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
}

// OpenAICompatibleClient is a client for calling OpenAI-compatible LLM services
type OpenAICompatibleClient struct {
	model  string
	client *openai.Client
}

// NewOpenAICompatibleClient creates a new LLM client
func NewOpenAICompatibleClient(model, apiKey, baseURL string) *OpenAICompatibleClient {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	return &OpenAICompatibleClient{
		model:  model,
		client: openai.NewClientWithConfig(config),
	}
}

// Generate calls the LLM API to generate a response
func (c *OpenAICompatibleClient) Generate(ctx context.Context, prompt, systemPrompt string) (string, error) {
	fmt.Println("正在调用大语言模型...")

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	})

	if err != nil {
		return "", fmt.Errorf("调用LLM API时发生错误: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM返回了空响应")
	}

	fmt.Println("大语言模型响应成功。")
	return resp.Choices[0].Message.Content, nil
}

// getWeather 通过调用 wttr.in API 查询真实的天气信息。
func getWeather(ctx context.Context, city string) string {
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)

	client := resty.New()
	var weatherData WeatherResponse

	resp, err := client.R().
		SetContext(ctx).
		SetResult(&weatherData).
		Get(url)

	if err != nil {
		return fmt.Sprintf("错误：查询天气时遇到网络问题 - %v", err)
	}

	if !resp.IsSuccess() {
		return fmt.Sprintf("错误：天气API返回状态码 %d", resp.StatusCode())
	}

	if len(weatherData.CurrentCondition) == 0 || len(weatherData.CurrentCondition[0].WeatherDesc) == 0 {
		return "错误：天气数据格式不正确"
	}

	condition := weatherData.CurrentCondition[0]
	weatherDesc := condition.WeatherDesc[0].Value
	tempC := condition.TempC

	return fmt.Sprintf("%s当前天气：%s，气温%s摄氏度", city, weatherDesc, tempC)
}

// getAttraction 根据城市和天气，使用Tavily Search API搜索并返回优化后的景点推荐。
func getAttraction(ctx context.Context, city, weather string) string {
	tavilyAPIKey := os.Getenv("TAVILY_API_KEY")
	if tavilyAPIKey == "" {
		fmt.Println("错误：未配置TAVILY_API_KEY。为了演示，返回模拟数据。")
		return fmt.Sprintf("根据%s的%s天气，推荐以下景点：\n- 故宫博物院：中国古代皇家宫殿\n- 长城：世界文化遗产", city, weather)
	}

	client := tavily.New(tavilyAPIKey, nil)

	query := fmt.Sprintf("'%s' 在'%s'天气下最值得去的旅游景点推荐及理由", city, weather)

	resp, err := client.Search(ctx, query, &tavily.SearchOptions{
		SearchDepth:   "basic",
		IncludeAnswer: true,
	})
	if err != nil {
		return fmt.Sprintf("错误：执行Tavily搜索时出现问题: %v", err)
	}

	if resp.Answer != "" {
		return resp.Answer
	}

	results := []string{}
	for _, result := range resp.Results {
		results = append(results, fmt.Sprintf("- %s: %s", result.Title, result.Content))
	}

	if len(results) == 0 {
		return "抱歉，没有找到相关的旅游景点推荐。"
	}

	return fmt.Sprintf("根据搜索，为您找到以下信息：\n%s", strings.Join(results, "\n"))
}

// parseAction extracts function name and arguments from action string
func parseAction(actionStr string) (functionName string, args map[string]string, err error) {
	// Match pattern: function_name(arg1="value1", arg2="value2")
	funcNameRegex := regexp.MustCompile(`(\w+)\(`)
	argsRegex := regexp.MustCompile(`(\w+)="([^"]*)"`)

	funcMatch := funcNameRegex.FindStringSubmatch(actionStr)
	if len(funcMatch) < 2 {
		return "", nil, fmt.Errorf("无法解析函数名")
	}

	functionName = funcMatch[1]
	args = make(map[string]string)

	argsMatches := argsRegex.FindAllStringSubmatch(actionStr, -1)
	for _, match := range argsMatches {
		if len(match) >= 3 {
			args[match[1]] = match[2]
		}
	}

	return functionName, args, nil
}

// executeTool executes the specified tool with given arguments
func executeTool(ctx context.Context, toolName string, args map[string]string) string {
	switch toolName {
	case "get_weather":
		if city, ok := args["city"]; ok {
			return getWeather(ctx, city)
		}
		return "错误：get_weather 需要 city 参数"

	case "get_attraction":
		city, hasCity := args["city"]
		weather, hasWeather := args["weather"]
		if hasCity && hasWeather {
			return getAttraction(ctx, city, weather)
		}
		return "错误：get_attraction 需要 city 和 weather 参数"

	case "finish":
		if answer, ok := args["answer"]; ok {
			return answer
		}
		return "错误：finish 需要 answer 参数"

	default:
		return fmt.Sprintf("错误：未定义的工具 '%s'", toolName)
	}
}

func main() {
	// Configure LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	modelID := os.Getenv("LLM_MODEL_ID")

	if apiKey == "" {
		log.Fatal("错误：请设置 OPENAI_API_KEY 环境变量")
	}
	if modelID == "" {
		modelID = "gpt-4o-mini"
	}

	llm := NewOpenAICompatibleClient(modelID, apiKey, baseURL)

	// Initialize
	userPrompt := "你好，请帮我查询一下今天北京的天气，然后根据天气推荐一个合适的旅游景点。"
	promptHistory := []string{userPrompt}

	fmt.Printf("用户输入: %s\n%s\n", userPrompt, strings.Repeat("=", 40))

	ctx := context.Background()
	maxIterations := 5

	// Run main loop
	for i := 0; i < maxIterations; i++ {
		fmt.Printf("\n--- 循环 %d ---\n\n", i+1)

		// Build full prompt
		fullPrompt := strings.Join(promptHistory, "\n")

		// Call LLM for thinking
		llmOutput, err := llm.Generate(ctx, fullPrompt, agentSystemPrompt)
		if err != nil {
			log.Printf("错误：%v\n", err)
			break
		}

		fmt.Printf("模型输出:\n%s\n\n", llmOutput)
		promptHistory = append(promptHistory, llmOutput)

		// Parse and execute action
		actionRegex := regexp.MustCompile(`(?s)Action:\s*(.*)`)
		actionMatch := actionRegex.FindStringSubmatch(llmOutput)

		if len(actionMatch) < 2 {
			fmt.Println("解析错误：模型输出中未找到 Action。")
			break
		}

		actionStr := strings.TrimSpace(actionMatch[1])

		// Check if task is finished
		if strings.HasPrefix(actionStr, "finish") {
			finishRegex := regexp.MustCompile(`finish\(answer="([^"]*)"\)`)
			finishMatch := finishRegex.FindStringSubmatch(actionStr)
			if len(finishMatch) >= 2 {
				finalAnswer := finishMatch[1]
				fmt.Printf("任务完成，最终答案: %s\n", finalAnswer)
				break
			}
		}

		// Parse tool name and arguments
		toolName, args, err := parseAction(actionStr)
		if err != nil {
			fmt.Printf("解析错误：%v\n", err)
			break
		}

		// Execute tool
		observation := executeTool(ctx, toolName, args)

		// Record observation
		observationStr := fmt.Sprintf("Observation: %s", observation)
		fmt.Printf("%s\n%s\n", observationStr, strings.Repeat("=", 40))
		promptHistory = append(promptHistory, observationStr)
	}

	fmt.Println("\n程序结束")
}
