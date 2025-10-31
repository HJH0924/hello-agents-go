package chapter04

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type HelloAgentsLLM struct {
	model  string
	client *openai.Client
}

func NewHelloAgentsLLM(model, apiKey, baseUrl string, timeout int) *HelloAgentsLLM {
	config := openai.DefaultConfig(apiKey)
	if baseUrl != "" {
		config.BaseURL = baseUrl
	}
	if timeout > 0 {
		config.HTTPClient = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	} else {
		config.HTTPClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}
	return &HelloAgentsLLM{
		model:  model,
		client: openai.NewClientWithConfig(config),
	}
}

func (l *HelloAgentsLLM) Think(ctx context.Context, messages []openai.ChatCompletionMessage, temperature float32) (string, error) {
	fmt.Printf("🧠 正在调用 %s 模型...\n", l.model)
	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       l.model,
		Messages:    messages,
		Temperature: temperature,
		Stream:      true,
	})
	if err != nil {
		return "", fmt.Errorf("CreateChatCompletionStream: %w", err)
	}
	defer stream.Close()
	// 处理流式响应
	fmt.Println("✅ 大语言模型响应成功:")
	collectedContent := []string{}
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println()
			break
		}
		if err != nil {
			return "", fmt.Errorf("Stream: %w", err)
		}
		if len(resp.Choices) == 0 {
			continue
		}
		content := resp.Choices[0].Delta.Content
		fmt.Printf("%s", content)
		collectedContent = append(collectedContent, content)
	}

	return strings.Join(collectedContent, ""), nil
}
