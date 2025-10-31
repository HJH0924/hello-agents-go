package chapter04

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReactAgent_Run(t *testing.T) {
	toolExecutor := NewToolExecutor()
	toolExecutor.RegisterTool("GoogleSearch", "一个网页搜索引擎。当你需要回答关于时事、事实以及在你的知识库中找不到的信息时，应使用此工具。", GoogleSearch)

	llm := NewHelloAgentsLLM("gpt-4o-mini", os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENAI_BASE_URL"), 60)

	agent := NewReactAgent(llm, toolExecutor, 5)
	question := "华为最新的手机是哪一款？它的主要卖点是什么？"
	resp, err := agent.Run(context.Background(), question)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	fmt.Println(resp)
}
