package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

// Rule 表示一个规则，包含正则表达式和响应模板列表
type Rule struct {
	pattern   *regexp.Regexp
	responses []string
	extractor func([]string) map[string]string // 用于提取关键信息
}

// Context 存储对话上下文信息
type Context struct {
	name       string
	age        string
	occupation string
	hobbies    []string
	topics     []string // 用户提到过的话题
}

// 全局上下文
var context = &Context{
	hobbies: make([]string, 0),
	topics:  make([]string, 0),
}

// 定义规则库
var rules = []Rule{
	// === 记忆提取规则（优先匹配） ===
	{
		pattern: regexp.MustCompile(`(?i)(?:my name is|call me|i'm) ([a-zA-Z]+)`),
		responses: []string{
			"Nice to meet you, {name}! How can I help you today?",
			"Hello {name}! What brings you here?",
			"It's a pleasure to meet you, {name}. What would you like to talk about?",
		},
		extractor: func(match []string) map[string]string {
			if len(match) > 1 && !isNumber(match[1]) {
				context.name = match[1]
				return map[string]string{"name": match[1]}
			}
			return nil
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)i am (\d+) years old`),
		responses: []string{
			"{age} is a great age! Tell me more about yourself.",
			"I see, you're {age}. How do you feel about that?",
			"At {age}, what are your main concerns?",
		},
		extractor: func(match []string) map[string]string {
			if len(match) > 1 {
				context.age = match[1]
				return map[string]string{"age": match[1]}
			}
			return nil
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)i (?:work as|am a|am an) (.+?)(?:\.|$)`),
		responses: []string{
			"Being a {occupation} must be interesting. Do you enjoy it?",
			"How long have you been a {occupation}?",
			"What do you like most about being a {occupation}?",
		},
		extractor: func(match []string) map[string]string {
			if len(match) > 1 {
				context.occupation = strings.TrimSpace(match[1])
				return map[string]string{"occupation": context.occupation}
			}
			return nil
		},
	},

	// === 工作相关规则 ===
	{
		pattern: regexp.MustCompile(`(?i).*\b(work|job|career|office|boss|colleague)\b.*`),
		responses: []string{
			"Work can be challenging. What aspect of your work concerns you most?",
			"Tell me more about your work situation.",
			"How does your work make you feel?",
			"What would you change about your work if you could?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)i (?:hate|dislike|don't like) (?:my )?work`),
		responses: []string{
			"What specifically bothers you about your work?",
			"Have you thought about what kind of work would make you happier?",
			"How long have you felt this way about your work?",
		},
	},

	// === 学习相关规则 ===
	{
		pattern: regexp.MustCompile(`(?i).*\b(study|studying|learn|learning|school|university|exam|course)\b.*`),
		responses: []string{
			"Learning is important. What are you studying?",
			"How are your studies going?",
			"What challenges are you facing in your studies?",
			"What motivates you to learn?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)i am studying (.*)`),
		responses: []string{
			"Why did you choose to study {0}?",
			"What do you find most interesting about {0}?",
			"How is your study of {0} going?",
		},
	},

	// === 爱好相关规则 ===
	{
		pattern: regexp.MustCompile(`(?i)i (?:like|love|enjoy) (.+?)(?:\.|$)`),
		responses: []string{
			"That's wonderful! What do you like most about {0}?",
			"How long have you been interested in {0}?",
			"Does {0} help you relax?",
		},
		extractor: func(match []string) map[string]string {
			if len(match) > 1 {
				hobby := strings.TrimSpace(match[1])
				context.hobbies = append(context.hobbies, hobby)
				return map[string]string{"hobby": hobby}
			}
			return nil
		},
	},
	{
		pattern: regexp.MustCompile(`(?i).*\b(hobby|hobbies|pastime|interest)\b.*`),
		responses: []string{
			"What do you like to do in your free time?",
			"Hobbies are important for balance. What are yours?",
			"Tell me about your interests.",
		},
	},

	// === 情感相关规则 ===
	{
		pattern: regexp.MustCompile(`(?i)i (?:feel|am feeling) (sad|depressed|anxious|worried|stressed|happy|excited)`),
		responses: []string{
			"What makes you feel {0}?",
			"How long have you been feeling {0}?",
			"When you feel {0}, what do you usually do?",
		},
	},

	// === 原有规则 ===
	{
		pattern: regexp.MustCompile(`(?i)I need (.*)`),
		responses: []string{
			"Why do you need {0}?",
			"Would it really help you to get {0}?",
			"Are you sure you need {0}?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)Why don't you (.*)(\?)?`),
		responses: []string{
			"Do you really think I don't {0}?",
			"Perhaps eventually I will {0}.",
			"Do you really want me to {0}?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)Why can't I (.*)(\?)?`),
		responses: []string{
			"Do you think you should be able to {0}?",
			"If you could {0}, what would you do?",
			"I don't know -- why can't you {0}?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)I am (.*)`),
		responses: []string{
			"Did you come to me because you are {0}?",
			"How long have you been {0}?",
			"How do you feel about being {0}?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i).* mother .*`),
		responses: []string{
			"Tell me more about your mother.",
			"What was your relationship with your mother like?",
			"How do you feel about your mother?",
		},
	},
	{
		pattern: regexp.MustCompile(`(?i).* father .*`),
		responses: []string{
			"Tell me more about your father.",
			"How did your father make you feel?",
			"What has your father taught you?",
		},
	},
	{
		pattern: regexp.MustCompile(`.*`),
		responses: []string{
			"Please tell me more.",
			"Let's change focus a bit... Tell me about your family.",
			"Can you elaborate on that?",
		},
	},
}

// pronounSwap 定义代词转换规则
var pronounSwap = map[string]string{
	"i":     "you",
	"you":   "i",
	"me":    "you",
	"my":    "your",
	"am":    "are",
	"are":   "am",
	"was":   "were",
	"i'd":   "you would",
	"i've":  "you have",
	"i'll":  "you will",
	"yours": "mine",
	"mine":  "yours",
}

// isNumber 检查字符串是否为数字
func isNumber(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

// extractAllInfo 从输入中提取所有可能的信息，返回提取到的信息数量
func extractAllInfo(userInput string) int {
	extracted := 0
	for _, rule := range rules {
		if rule.extractor != nil {
			matches := rule.pattern.FindAllStringSubmatch(userInput, -1)
			for _, match := range matches {
				if result := rule.extractor(match); result != nil {
					extracted++
				}
			}
		}
	}
	return extracted
}

// generateMultiInfoResponse 当提取到多个信息时，生成综合响应
func generateMultiInfoResponse() string {
	parts := make([]string, 0)

	if context.name != "" {
		parts = append(parts, fmt.Sprintf("Nice to meet you, %s!", context.name))
	}
	if context.age != "" {
		parts = append(parts, fmt.Sprintf("You're %s years old.", context.age))
	}
	if len(context.hobbies) > 0 {
		parts = append(parts, fmt.Sprintf("I see you enjoy %s.", context.hobbies[len(context.hobbies)-1]))
	}

	if len(parts) > 0 {
		return strings.Join(parts, " ") + " Tell me more about yourself."
	}

	return "That's interesting! Tell me more."
}

// swapPronouns 对输入短语中的代词进行第一/第二人称转换
func swapPronouns(phrase string) string {
	words := strings.Fields(strings.ToLower(phrase))
	swappedWords := make([]string, len(words))

	for i, word := range words {
		if swapped, exists := pronounSwap[word]; exists {
			swappedWords[i] = swapped
		} else {
			swappedWords[i] = word
		}
	}

	return strings.Join(swappedWords, " ")
}

// fillTemplate 使用上下文信息填充响应模板
func fillTemplate(template string, capturedGroup string) string {
	result := template

	// 替换捕获的组
	result = strings.ReplaceAll(result, "{0}", swapPronouns(capturedGroup))

	// 替换上下文信息
	if context.name != "" {
		result = strings.ReplaceAll(result, "{name}", context.name)
	}
	if context.age != "" {
		result = strings.ReplaceAll(result, "{age}", context.age)
	}
	if context.occupation != "" {
		result = strings.ReplaceAll(result, "{occupation}", context.occupation)
	}

	return result
}

// respond 根据规则库生成响应
func respond(userInput string) string {
	// 第一步：提取所有可能的信息
	extractedCount := extractAllInfo(userInput)

	// 如果提取到多个信息，使用综合响应
	if extractedCount >= 2 {
		return generateMultiInfoResponse()
	}

	// 第二步：找到最佳匹配规则并生成响应
	for _, rule := range rules {
		match := rule.pattern.FindStringSubmatch(userInput)
		if match != nil {
			// 捕获匹配到的部分
			capturedGroup := ""
			if len(match) > 1 {
				capturedGroup = match[1]
			}

			// 从模板中随机选择一个并格式化
			template := rule.responses[rand.Intn(len(rule.responses))]
			response := fillTemplate(template, capturedGroup)

			return response
		}
	}

	// 如果没有匹配任何特定规则，使用最后的通配符规则
	lastRule := rules[len(rules)-1]
	return lastRule.responses[rand.Intn(len(lastRule.responses))]
}

// showContext 显示当前记忆的上下文信息
func showContext() {
	fmt.Println("\n=== Context Memory ===")
	if context.name != "" {
		fmt.Printf("Name: %s\n", context.name)
	}
	if context.age != "" {
		fmt.Printf("Age: %s\n", context.age)
	}
	if context.occupation != "" {
		fmt.Printf("Occupation: %s\n", context.occupation)
	}
	if len(context.hobbies) > 0 {
		fmt.Printf("Hobbies: %s\n", strings.Join(context.hobbies, ", "))
	}
	fmt.Printf("======================\n\n")
}

func main() {
	fmt.Println("=== Enhanced ELIZA with Context Memory ===")
	fmt.Println("Therapist: Hello! How can I help you today?")
	fmt.Println("(Type 'context' to see what I remember, 'quit' to exit)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		userInputLower := strings.ToLower(strings.TrimSpace(userInput))

		// 特殊命令
		if userInputLower == "quit" || userInputLower == "exit" || userInputLower == "bye" {
			if context.name != "" {
				fmt.Printf("Therapist: Goodbye, %s. It was nice talking to you.\n", context.name)
			} else {
				fmt.Println("Therapist: Goodbye. It was nice talking to you.")
			}
			break
		}

		if userInputLower == "context" {
			showContext()
			continue
		}

		response := respond(userInput)
		fmt.Printf("Therapist: %s\n", response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
