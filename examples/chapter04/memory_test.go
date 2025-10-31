package chapter04

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory_Operations(t *testing.T) {
	// 测试 Memory 模块的基本操作
	memory := NewMemory()

	// 添加执行记录
	memory.AddRecord(RecordTypeExecution, "def test(): pass")
	assert.Equal(t, 1, len(memory.records))

	// 添加反思记录
	memory.AddRecord(RecordTypeReflection, "代码需要改进")
	assert.Equal(t, 2, len(memory.records))

	// 获取最后一次执行
	lastExec := memory.GetLastExecution()
	assert.Equal(t, "def test(): pass", lastExec)

	// 添加另一次执行
	memory.AddRecord(RecordTypeExecution, "def test_v2(): pass")
	lastExec = memory.GetLastExecution()
	assert.Equal(t, "def test_v2(): pass", lastExec)

	// 获取轨迹
	trajectory := memory.GetTrajectory()
	fmt.Println(trajectory)
	assert.Contains(t, trajectory, "上一轮尝试")
	assert.Contains(t, trajectory, "评审员反馈")
}

func TestMemory_EmptyGetLastExecution(t *testing.T) {
	// 测试空记忆时获取最后执行
	memory := NewMemory()
	lastExec := memory.GetLastExecution()
	assert.Equal(t, "", lastExec)

	// 只有反思记录时
	memory.AddRecord(RecordTypeReflection, "some feedback")
	lastExec = memory.GetLastExecution()
	assert.Equal(t, "", lastExec)
}
