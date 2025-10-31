# 习题4：Plan-and-Solve 分析

## 问题

`Plan-and-Solve` 范式将任务分解为"规划"和"执行"两个阶段。请深入分析：

1. 在4.3节的实现中，规划阶段生成的计划是"静态"的（一次性生成，不可修改）。如果在执行过程中发现某个步骤无法完成或结果不符合预期，应该如何设计一个"动态重规划"机制？
2. 对比 `Plan-and-Solve` 与 `ReAct`：在处理"预订一次从北京到上海的商务旅行（包括机票、酒店、租车）"这样的任务时，哪种范式更合适？为什么？
3. 尝试设计一个"分层规划"系统：先生成高层次的抽象计划，然后针对每个高层步骤再生成详细的子计划。这种设计有什么优势？

---

## 回答

### 1. 动态重规划机制设计

#### 核心思路

在Plan-and-Solve的执行阶段加入"检查点"机制，当发现问题时触发重规划。

#### 架构设计

```
执行流程：
Plan → Execute Step 1 → Check → Pass → Execute Step 2 → Check → Fail → Replan → Continue
```

#### Go 实现框架

```go
type DynamicPlanAndSolveAgent struct {
    planner  *Planner
    executor *Executor
    checker  *ExecutionChecker  // 新增：执行检查器
}

// ExecutionChecker 检查执行结果是否符合预期
type ExecutionChecker struct {
    llmClient *HelloAgentsLLM
}

func (c *ExecutionChecker) Check(ctx context.Context, step string, result string, originalQuestion string) (bool, string, error) {
    prompt := fmt.Sprintf(`
你是一个执行检查专家。请判断以下步骤的执行结果是否符合预期：

原始问题：%s
当前步骤：%s
执行结果：%s

请回答：
1. 结果是否符合预期？(是/否)
2. 如果不符合，问题是什么？
3. 是否需要重新规划？

格式：
{
  "is_valid": true/false,
  "issue": "问题描述",
  "need_replan": true/false
}
`, originalQuestion, step, result)

    // 调用LLM检查
    response, err := c.llmClient.Think(ctx, []openai.ChatCompletionMessage{
        {Role: openai.ChatMessageRoleUser, Content: prompt},
    }, 0.3)

    // 解析响应
    // ...
}

// Run 带动态重规划的执行
func (a *DynamicPlanAndSolveAgent) Run(ctx context.Context, question string) (string, error) {
    // 1. 初始规划
    plan, err := a.planner.Plan(ctx, question)
    if err != nil {
        return "", err
    }

    // 2. 执行 + 检查 + 重规划循环
    completedSteps := []string{}
    currentPlan := plan

    for i := 0; i < len(currentPlan); i++ {
        step := currentPlan[i]
        result, err := a.executor.ExecuteStep(ctx, question, step, completedSteps)
        
        // 检查执行结果
        isValid, issue, needReplan, err := a.checker.Check(ctx, step, result, question)
        
        if !isValid && needReplan {
            fmt.Printf("⚠️ 步骤执行异常：%s\n", issue)
            fmt.Println("🔄 触发重新规划...")
            
            // 重新规划剩余步骤
            remainingTask := fmt.Sprintf("继续完成任务。已完成：%v。遇到问题：%s", completedSteps, issue)
            newPlan, err := a.planner.Plan(ctx, remainingTask)
            if err != nil {
                return "", err
            }
            
            currentPlan = append(completedSteps, newPlan...)
            i = len(completedSteps) - 1  // 从已完成的下一步继续
            continue
        }
        
        completedSteps = append(completedSteps, result)
    }

    return completedSteps[len(completedSteps)-1], nil
}
```

#### 触发重规划的条件

1. **执行失败**：工具调用返回错误
2. **结果不符合预期**：LLM判断结果质量低
3. **依赖条件变化**：前提假设不再成立
4. **超时或资源不足**

#### 优势

- **容错性强**：能够应对执行中的意外情况
- **自适应**：根据实际情况调整计划
- **避免全盘重来**：只重规划剩余部分

---

### 2. Plan-and-Solve vs ReAct：商务旅行预订场景

#### 任务分析

**任务**："预订一次从北京到上海的商务旅行（包括机票、酒店、租车）"

**任务特点**：
- 多个子任务（机票、酒店、租车）
- 明确的逻辑顺序
- 子任务之间有依赖（如：到达时间影响租车时间）
- 需要考虑时间、预算等约束

#### Plan-and-Solve 方案

**规划阶段**：
```
Plan = [
    "步骤1：查询北京到上海的航班，选择合适时间和价格的航班",
    "步骤2：根据航班到达时间，预订上海的酒店（靠近会议地点）",
    "步骤3：预订租车服务，取车时间为航班到达后1小时",
    "步骤4：确认所有预订并生成行程单"
]
```

**执行阶段**：严格按照计划执行

**优点**：
- ✅ 结构清晰：所有步骤一目了然
- ✅ 避免遗漏：提前规划好所有环节
- ✅ 时间估算准确：可以预估总耗时
- ✅ 资源规划：可以优化预算分配

**缺点**：
- ❌ 不够灵活：如果首选航班没票，计划受影响
- ❌ 静态约束：无法动态调整（如发现更便宜的组合）

#### ReAct 方案

**执行过程**：
```
Thought: 需要预订航班，先搜索可用航班
Action: SearchFlight[北京-上海, 明天]
Observation: 找到3个航班，MU5178最便宜

Thought: MU5178性价比高，预订它
Action: BookFlight[MU5178]
Observation: 预订成功，到达时间14:00

Thought: 根据到达时间14:00，预订附近酒店
Action: SearchHotel[上海, 14:00入住]
...
```

**优点**：
- ✅ 高度灵活：根据实时反馈调整
- ✅ 动态优化：发现更好选择可以调整
- ✅ 处理异常：航班没票立即切换备选

**缺点**：
- ❌ 可能遗漏：没有全局视野，可能忘记租车
- ❌ 缺乏优化：步进决策可能错失整体最优方案
- ❌ 成本高：需要多次LLM调用

#### 推荐方案

**结论**：推荐使用 **Plan-and-Solve**

**理由**：

1. **任务结构明确**
   - 商务旅行预订是典型的结构化任务
   - 步骤清晰：机票 → 酒店 → 租车
   - Plan-and-Solve的优势得以充分发挥

2. **全局优化需求**
   - 需要考虑整体预算
   - 需要协调各环节时间
   - 规划阶段可以进行整体优化

3. **减少决策成本**
   - 一次性规划好，执行效率高
   - 减少LLM调用次数
   - 用户体验更好（不用等待多次决策）

4. **可预测性**
   - 用户在执行前就能看到完整计划
   - 更符合商务场景的需求

**改进建议**：
- 结合动态重规划机制，处理预订失败等意外
- 在规划阶段考虑备选方案

---

### 3. 分层规划系统设计

#### 核心思想

```
高层抽象计划（What）
    ↓
详细执行计划（How）
```

#### 架构设计

```go
type HierarchicalPlanner struct {
    highLevelPlanner *Planner  // 高层规划器
    lowLevelPlanner  *Planner  // 低层规划器
}

type HierarchicalPlan struct {
    HighLevel []HighLevelStep
}

type HighLevelStep struct {
    Goal      string        // 高层目标
    SubPlan   []string      // 详细子步骤
}

func (p *HierarchicalPlanner) Plan(ctx context.Context, question string) (*HierarchicalPlan, error) {
    // 1. 生成高层计划
    highLevelSteps, err := p.highLevelPlanner.Plan(ctx, question)
    if err != nil {
        return nil, err
    }

    // 2. 为每个高层步骤生成详细计划
    var plan HierarchicalPlan
    for _, highStep := range highLevelSteps {
        subPlan, err := p.lowLevelPlanner.Plan(ctx, highStep)
        if err != nil {
            return nil, err
        }

        plan.HighLevel = append(plan.HighLevel, HighLevelStep{
            Goal:    highStep,
            SubPlan: subPlan,
        })
    }

    return &plan, nil
}
```

#### 示例：电商系统开发

**高层计划**：
```
[
    "需求分析和架构设计",
    "数据库设计和实现",
    "后端API开发",
    "前端界面开发",
    "测试和部署"
]
```

**详细子计划（以"后端API开发"为例）**：
```
[
    "设计RESTful API接口规范",
    "实现用户认证和授权模块",
    "实现商品管理API",
    "实现订单处理API",
    "实现支付集成",
    "编写API文档"
]
```

#### 优势

1. **更好的任务分解**
   - 高层：关注目标和里程碑
   - 低层：关注具体实现步骤
   - 避免细节淹没全局

2. **灵活性**
   - 高层计划相对稳定
   - 低层计划可以灵活调整
   - 某个模块失败不影响整体框架

3. **并行执行可能**
   - 独立的高层目标可以并行
   - 例如：前后端可以并行开发

4. **进度跟踪**
   - 清晰的里程碑
   - 易于评估完成度
   - 便于向用户报告进度

5. **认知负担小**
   - 每次只需关注一个层次
   - LLM生成的计划更准确
   - 避免一次规划过于庞大

#### 实际应用场景

- ✅ 软件开发项目
- ✅ 科研项目规划
- ✅ 商业计划制定
- ✅ 教育课程设计
- ❌ 简单任务（过度设计）

---

## 总结

1. **动态重规划**：在执行阶段加入检查点，发现问题时触发重规划
2. **场景选择**：商务旅行预订适合Plan-and-Solve，因为结构清晰、需要全局优化
3. **分层规划**：适合复杂项目，提供更好的任务分解和进度跟踪能力

**关键洞察**：
- Plan-and-Solve的静态性可以通过动态重规划弥补
- 选择范式要根据任务的结构性和可预测性
- 复杂任务需要分层思维，避免认知过载
