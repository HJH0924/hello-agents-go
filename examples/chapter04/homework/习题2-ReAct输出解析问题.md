# 习题2：ReAct 输出解析问题

## 问题

在4.2节的 `ReAct` 实现中，我们使用了正则表达式来解析大语言模型的输出（如 `Thought` 和 `Action`）。请思考：

1. 当前的解析方法存在哪些潜在的脆弱性？在什么情况下可能会失败？
2. 除了正则表达式，还有哪些更鲁棒的输出解析方案？
3. 尝试修改本章的代码，使用一种更可靠的输出格式，并对比两种方案的优缺点

---

## 回答

### 1. 当前解析方法的脆弱性分析

#### 当前实现回顾

```go
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
```

#### 潜在的脆弱性

##### 1. 大小写敏感性问题

**问题**：正则表达式对大小写敏感

**失败情况**：
```
LLM输出：thought: 我需要搜索华为手机信息
        action: GoogleSearch[华为最新手机]

当前正则：`Thought: (.*)`  # 匹配失败，因为首字母小写
```

**影响**：无法解析出思考和行动，导致流程中断

##### 2. 多行内容截断问题

**问题**：`.` 不匹配换行符，只能捕获单行

**失败情况**：
```
LLM输出：
Thought: 用户询问华为最新手机信息。
         我需要先搜索华为官网，
         然后总结主要卖点。
Action: GoogleSearch[华为最新手机]

当前正则：只捕获第一行 "用户询问华为最新手机信息。"
```

**影响**：思考内容被截断，丢失重要信息

##### 3. 格式变体问题

**问题**：LLM可能输出格式的轻微变体

**失败情况**：
```
# 变体1：多余空格
Thought:  我需要搜索  # 冒号后有多个空格

# 变体2：中文冒号
Thought：我需要搜索  # 使用了中文冒号

# 变体3：缩进
  Thought: 我需要搜索  # 有前导空格

# 变体4：后缀内容
Thought: 我需要搜索
（这是一个重要的步骤）
Action: GoogleSearch[...]
```

**影响**：解析失败或解析不完整

##### 4. 嵌套内容干扰

**问题**：内容中包含关键字

**失败情况**：
```
LLM输出：
Thought: 用户在问"什么是Thought和Action的区别"
Action: Search[Thought和Action的区别]

当前正则：可能匹配到错误的位置
```

##### 5. 缺失字段处理

**问题**：LLM有时只输出部分字段

**失败情况**：
```
# 只输出 Action，没有 Thought
Action: GoogleSearch[华为手机]

# 只输出 Thought，没有 Action
Thought: 我已经有答案了
```

**影响**：当前代码返回空字符串，但没有明确的错误处理

##### 6. Action格式解析脆弱

**问题**：`ToolName[input]` 格式依赖严格的括号匹配

**失败情况**：
```
# 括号不匹配
Action: Search华为手机  # 缺少括号
Action: Search[华为[新款]手机]  # 嵌套括号
Action: Search华为手机]  # 左括号缺失
```

---

### 2. 更鲁棒的输出解析方案

#### 方案1：结构化输出 - JSON格式

**思路**：要求LLM输出标准JSON格式

**Prompt设计**：
```
请严格按照以下JSON格式进行回应：
{
  "thought": "你的思考过程",
  "action": {
    "tool": "工具名称",
    "input": "工具输入"
  }
}

如果已经得到最终答案：
{
  "thought": "我已经得到答案",
  "action": {
    "tool": "Finish",
    "input": "最终答案"
  }
}
```

**Go实现**：
```go
type LLMResponse struct {
    Thought string `json:"thought"`
    Action  struct {
        Tool  string `json:"tool"`
        Input string `json:"input"`
    } `json:"action"`
}

func (a *ReactAgent) parseJSONOutput(text string) (*LLMResponse, error) {
    // 提取JSON部分（处理markdown代码块）
    jsonStr := extractJSON(text)

    var response LLMResponse
    err := json.Unmarshal([]byte(jsonStr), &response)
    if err != nil {
        return nil, fmt.Errorf("解析JSON失败: %w", err)
    }

    return &response, nil
}

func extractJSON(text string) string {
    // 处理 ```json ... ``` 格式
    jsonRegex := regexp.MustCompile("```(?:json)?\\s*\\n?([\\s\\S]*?)```")
    if matches := jsonRegex.FindStringSubmatch(text); len(matches) > 1 {
        return matches[1]
    }

    // 尝试直接查找JSON对象
    start := strings.Index(text, "{")
    end := strings.LastIndex(text, "}")
    if start != -1 && end != -1 && end > start {
        return text[start : end+1]
    }

    return text
}
```

**优点**：
- 解析稳定，不受格式变化影响
- 可以包含复杂的嵌套结构
- 错误检测明确（JSON解析失败会报错）

**缺点**：
- LLM可能不严格遵循JSON格式
- JSON输出对LLM的要求更高
- 需要额外的JSON提取逻辑

#### 方案2：XML格式

**Prompt设计**：
```
请严格按照以下XML格式进行回应：
<response>
  <thought>你的思考过程</thought>
  <action>
    <tool>工具名称</tool>
    <input>工具输入</input>
  </action>
</response>
```

**Go实现**：
```go
type XMLResponse struct {
    XMLName xml.Name `xml:"response"`
    Thought string   `xml:"thought"`
    Action  struct {
        Tool  string `xml:"tool"`
        Input string `xml:"input"`
    } `xml:"action"`
}

func (a *ReactAgent) parseXMLOutput(text string) (*XMLResponse, error) {
    var response XMLResponse
    err := xml.Unmarshal([]byte(text), &response)
    if err != nil {
        return nil, fmt.Errorf("解析XML失败: %w", err)
    }
    return &response, nil
}
```

#### 方案3：改进的正则表达式（多行+容错）

**实现**：
```go
func (a *ReactAgent) parseOutputRobust(text string) (string, string) {
    // 使用多行模式和更宽松的匹配
    thoughtRegex := regexp.MustCompile(`(?i)Thought[：:]\s*(.+?)(?:\n(?:Action|$)|$)`)
    actionRegex := regexp.MustCompile(`(?i)Action[：:]\s*(.+?)(?:\n(?:Observation|Thought|$)|$)`)

    var thought, action string

    // 匹配Thought（忽略大小写，支持中英文冒号）
    if thoughtMatch := thoughtRegex.FindStringSubmatch(text); len(thoughtMatch) > 1 {
        thought = strings.TrimSpace(thoughtMatch[1])
    }

    // 匹配Action
    if actionMatch := actionRegex.FindStringSubmatch(text); len(actionMatch) > 1 {
        action = strings.TrimSpace(actionMatch[1])
    }

    return thought, action
}
```

**改进点**：
- `(?i)`：忽略大小写
- `[：:]`：同时支持中英文冒号
- `(.+?)`：非贪婪匹配，支持多行
- `(?:\n(?:Action|$)|$)`：匹配到下一个关键字或文本结尾

#### 方案4：基于分隔符的解析

**Prompt设计**：
```
请使用以下分隔符格式：
---THOUGHT---
你的思考过程
---ACTION---
ToolName[input]
---END---
```

**Go实现**：
```go
func (a *ReactAgent) parseDelimitedOutput(text string) (string, string) {
    thoughtDelim := "---THOUGHT---"
    actionDelim := "---ACTION---"
    endDelim := "---END---"

    thoughtStart := strings.Index(text, thoughtDelim)
    actionStart := strings.Index(text, actionDelim)
    endStart := strings.Index(text, endDelim)

    var thought, action string

    if thoughtStart != -1 && actionStart != -1 {
        thought = strings.TrimSpace(text[thoughtStart+len(thoughtDelim) : actionStart])
    }

    if actionStart != -1 && endStart != -1 {
        action = strings.TrimSpace(text[actionStart+len(actionDelim) : endStart])
    }

    return thought, action
}
```

#### 方案5：Function Calling（函数调用）

**思路**：使用OpenAI的Function Calling特性

**实现**：
```go
func (a *ReactAgent) Think(ctx context.Context, messages []openai.ChatCompletionMessage) error {
    // 定义可用的函数
    functions := []openai.FunctionDefinition{
        {
            Name:        "execute_action",
            Description: "执行一个具体的行动",
            Parameters: json.RawMessage(`{
                "type": "object",
                "properties": {
                    "thought": {
                        "type": "string",
                        "description": "当前的思考过程"
                    },
                    "tool": {
                        "type": "string",
                        "description": "要调用的工具名称"
                    },
                    "input": {
                        "type": "string",
                        "description": "工具的输入参数"
                    }
                },
                "required": ["thought", "tool", "input"]
            }`),
        },
    }

    response, err := a.llmClient.CreateChatCompletion(ctx,
        openai.ChatCompletionRequest{
            Model:     "gpt-4",
            Messages:  messages,
            Functions: functions,
            FunctionCall: openai.FunctionCall{
                Name: "execute_action",
            },
        },
    )

    if err != nil {
        return err
    }

    // 解析函数调用
    if response.Choices[0].Message.FunctionCall != nil {
        var params struct {
            Thought string `json:"thought"`
            Tool    string `json:"tool"`
            Input   string `json:"input"`
        }
        json.Unmarshal([]byte(response.Choices[0].Message.FunctionCall.Arguments), &params)

        // 使用解析出的参数
        fmt.Printf("Thought: %s\n", params.Thought)
        fmt.Printf("Action: %s[%s]\n", params.Tool, params.Input)
    }

    return nil
}
```

**优点**：
- 最稳定的方案，由模型原生支持
- 自动参数校验
- 不需要自定义解析逻辑

**缺点**：
- 只有部分模型支持
- 依赖特定API特性

---

### 3. 代码实现对比

#### 方案选择：JSON格式（平衡稳定性和通用性）

让我实现一个完整的对比示例：

```go
// 方案1：原始正则表达式解析
func (a *ReactAgent) parseOutputRegex(text string) (thought, action string, err error) {
    thoughtRegex := regexp.MustCompile(`Thought: (.*)`)
    actionRegex := regexp.MustCompile(`Action: (.*)`)

    if thoughtMatch := thoughtRegex.FindStringSubmatch(text); len(thoughtMatch) > 1 {
        thought = strings.TrimSpace(thoughtMatch[1])
    } else {
        return "", "", fmt.Errorf("无法解析Thought")
    }

    if actionMatch := actionRegex.FindStringSubmatch(text); len(actionMatch) > 1 {
        action = strings.TrimSpace(actionMatch[1])
    } else {
        return "", "", fmt.Errorf("无法解析Action")
    }

    return thought, action, nil
}

// 方案2：JSON格式解析
func (a *ReactAgent) parseOutputJSON(text string) (thought, action string, err error) {
    type Response struct {
        Thought string `json:"thought"`
        Action  string `json:"action"`
    }

    // 提取JSON
    jsonStr := extractJSON(text)
    if jsonStr == "" {
        return "", "", fmt.Errorf("未找到JSON内容")
    }

    var response Response
    if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
        return "", "", fmt.Errorf("JSON解析失败: %w", err)
    }

    return response.Thought, response.Action, nil
}

func extractJSON(text string) string {
    // 处理markdown代码块
    jsonRegex := regexp.MustCompile("```(?:json)?\\s*\\n?([\\s\\S]*?)```")
    if matches := jsonRegex.FindStringSubmatch(text); len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }

    // 查找JSON对象
    start := strings.Index(text, "{")
    end := strings.LastIndex(text, "}")
    if start != -1 && end != -1 && end > start {
        return text[start : end+1]
    }

    return ""
}
```

#### 对比测试

```go
func TestOutputParsing(t *testing.T) {
    testCases := []struct {
        name   string
        input  string
        wantOK bool
    }{
        {
            name: "标准格式-正则",
            input: `Thought: 我需要搜索信息
Action: GoogleSearch[华为手机]`,
            wantOK: true,
        },
        {
            name: "多行思考-正则失败",
            input: `Thought: 我需要搜索信息
这是第二行思考
Action: GoogleSearch[华为手机]`,
            wantOK: false,  // 正则只能捕获第一行
        },
        {
            name: "JSON格式",
            input: `{
  "thought": "我需要搜索信息\\n这是第二行思考",
  "action": "GoogleSearch[华为手机]"
}`,
            wantOK: true,
        },
    }

    agent := &ReactAgent{}

    for _, tc := range testCases {
        t.Run(tc.name+"-Regex", func(t *testing.T) {
            _, _, err := agent.parseOutputRegex(tc.input)
            if (err == nil) != tc.wantOK {
                t.Errorf("Regex解析: got error=%v, want success=%v", err, tc.wantOK)
            }
        })

        t.Run(tc.name+"-JSON", func(t *testing.T) {
            _, _, err := agent.parseOutputJSON(tc.input)
            // JSON格式对于测试1会失败（因为不是JSON）
            // 但对测试3成功
        })
    }
}
```

#### 方案对比总结

| 维度 | 正则表达式 | JSON格式 | Function Calling |
|------|-----------|----------|------------------|
| 稳定性 | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 通用性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| LLM要求 | 低 | 中 | 高 |
| 解析复杂度 | 低 | 中 | 低 |
| 错误提示 | 模糊 | 明确 | 明确 |
| 扩展性 | 差 | 好 | 最好 |

#### 推荐方案

1. **生产环境**：Function Calling（如果模型支持）> JSON格式 > 改进的正则表达式
2. **学习/原型**：正则表达式（简单直接）
3. **通用性优先**：JSON格式（平衡了稳定性和通用性）

---

## 总结

当前基于正则表达式的解析方法存在多个脆弱点，包括大小写敏感、多行内容截断、格式变体等问题。更鲁棒的方案包括：

1. **JSON格式**：结构化、易于解析、错误明确
2. **XML格式**：类似JSON但更冗长
3. **Function Calling**：最稳定但依赖模型支持
4. **改进的正则**：容错性更好的中间方案
5. **自定义分隔符**：简单但需要严格约定

推荐在实际应用中使用 **JSON格式** 或 **Function Calling**，它们提供了更好的稳定性和可维护性。
