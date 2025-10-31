package chapter04

import (
	"fmt"

	"github.com/Knetic/govaluate"
)

// CalculatorTool 计算器工具，用于执行数学表达式计算
func CalculatorTool(expression string) (string, error) {
	// 使用 govaluate 库解析和计算数学表达式
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return "", fmt.Errorf("表达式解析失败: %w", err)
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return "", fmt.Errorf("计算失败: %w", err)
	}

	return fmt.Sprintf("%v", result), nil
}

// 注册计算器工具
func RegisterCalculatorTool(executor *ToolExecutor) {
	executor.RegisterTool(
		"Calculator",
		"一个精确的数学计算器。输入数学表达式，返回计算结果。支持加减乘除、括号、幂运算等。示例：(123 + 456) * 789 / 12",
		CalculatorTool,
	)
}
