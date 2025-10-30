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
}

// 定义规则库：模式(正则表达式) -> 响应模板列表
var rules = []Rule{
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

// respond 根据规则库生成响应
func respond(userInput string) string {
	for _, rule := range rules {
		match := rule.pattern.FindStringSubmatch(userInput)
		if match != nil {
			// 捕获匹配到的部分
			capturedGroup := ""
			if len(match) > 1 {
				capturedGroup = match[1]
			}

			// 进行代词转换
			swappedGroup := swapPronouns(capturedGroup)

			// 从模板中随机选择一个并格式化
			response := rule.responses[rand.Intn(len(rule.responses))]
			response = strings.ReplaceAll(response, "{0}", swappedGroup)

			return response
		}
	}

	// 如果没有匹配任何特定规则，使用最后的通配符规则
	lastRule := rules[len(rules)-1]
	return lastRule.responses[rand.Intn(len(lastRule.responses))]
}

func main() {
	fmt.Println("Therapist: Hello! How can I help you today?")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		userInputLower := strings.ToLower(strings.TrimSpace(userInput))

		if userInputLower == "quit" || userInputLower == "exit" || userInputLower == "bye" {
			fmt.Println("Therapist: Goodbye. It was nice talking to you.")
			break
		}

		response := respond(userInput)
		fmt.Printf("Therapist: %s\n", response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
