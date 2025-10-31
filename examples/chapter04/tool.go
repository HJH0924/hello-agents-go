package chapter04

import "fmt"

type tool struct {
	description string
	call        func(string) (string, error)
}

// ToolExecutor 一个工具执行器，负责管理和执行工具。
type ToolExecutor struct {
	tools map[string]tool
}

func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools: make(map[string]tool),
	}
}

// RegisterTool 注册一个工具。
func (t *ToolExecutor) RegisterTool(name string, description string, call func(string) (string, error)) {
	if _, ok := t.tools[name]; ok {
		fmt.Printf("警告：工具 '%s' 已存在，将被覆盖。\n", name)
	}
	t.tools[name] = tool{
		description: description,
		call:        call,
	}
	fmt.Printf("工具 '%s' 已注册。\n", name)
}

func (t *ToolExecutor) GetToolCall(name string) (func(string) (string, error), error) {
	if _, ok := t.tools[name]; ok {
		return t.tools[name].call, nil
	}
	return nil, fmt.Errorf("工具 '%s' 不存在。", name)
}

func (t *ToolExecutor) GetAvailableTools() string {
	var tools string
	for name, tool := range t.tools {
		tools += fmt.Sprintf("- %s: %s\n", name, tool.description)
	}
	return tools
}
