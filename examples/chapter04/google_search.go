package chapter04

import (
	"fmt"
	"os"
	"strings"

	serpApi "github.com/serpapi/google-search-results-golang"
)

// GoogleSearch 一个基于SerpApi的实战网页搜索引擎工具。
// 它会智能地解析搜索结果，优先返回直接答案或知识图谱信息。
func GoogleSearch(q string) (string, error) {
	fmt.Printf("🔍 正在执行 [SerpApi] 网页搜索: %s\n", q)

	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		// 返回一个模拟数据
		fmt.Println("🔍 未配置 SERPAPI_API_KEY，返回模拟数据")
		return "华为最新手机是华为Mate 60 Pro，主要卖点是麒麟9000s芯片、5G网络、120Hz刷新率、12GB内存、512GB存储空间。", nil
	}

	params := map[string]string{
		"q":  q,
		"gl": "cn",    // 国家代码
		"hl": "zh-cn", // 语言代码
	}

	query := serpApi.NewGoogleSearch(params, apiKey)
	resp, err := query.GetJSON()
	if err != nil {
		return "", err
	}
	if len(resp) == 0 {
		return fmt.Sprintf("对不起，没有找到关于 '%s' 的信息。", q), nil
	}

	// 智能解析：优先寻找最直接的答案
	// 1. 检查 "answer_box_list"
	if answerBoxList, ok := resp["answer_box_list"].([]interface{}); ok && len(answerBoxList) > 0 {
		var results []string
		for _, item := range answerBoxList {
			if strItem, ok := item.(string); ok {
				results = append(results, strItem)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, "\n"), nil
		}
	}
	// 2. 检查 "answer_box"
	if answerBox, ok := resp["answer_box"].(map[string]interface{}); ok {
		if answer, ok := answerBox["answer"].(string); ok {
			return answer, nil
		}
	}
	// 3. 检查 "knowledge_graph"
	if knowledgeGraph, ok := resp["knowledge_graph"].(map[string]interface{}); ok {
		if description, ok := knowledgeGraph["description"].(string); ok {
			return description, nil
		}
	}
	// 4. 如果没有直接答案，则返回前三个有机结果的摘要
	if organicResults, ok := resp["organic_results"].([]interface{}); ok && len(organicResults) > 0 {
		var snippets []string
		// 限制最多只取前3个结果
		numResults := 3
		if len(organicResults) < numResults {
			numResults = len(organicResults)
		}

		for i := 0; i < numResults; i++ {
			if result, ok := organicResults[i].(map[string]interface{}); ok {
				title, _ := result["title"].(string)
				snippet, _ := result["snippet"].(string)
				snippets = append(snippets, fmt.Sprintf("[%d] %s\n%s", i+1, title, snippet))
			}
		}
		if len(snippets) > 0 {
			return strings.Join(snippets, "\n\n"), nil
		}
	}

	// 5. 如果没有任何有效信息，返回最终的提示
	return fmt.Sprintf("对不起，没有找到关于 '%s' 的信息。", q), nil
}

// 注册谷歌搜索工具
func RegisterGoogleSearchTool(executor *ToolExecutor) {
	executor.RegisterTool(
		"GoogleSearch",
		"一个网页搜索引擎。当你需要回答关于时事、事实以及在你的知识库中找不到的信息时，应使用此工具。",
		GoogleSearch,
	)
}
