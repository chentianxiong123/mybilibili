package importer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration_UnderOneHour(t *testing.T) {
	assert.Equal(t, "05:03", formatDuration(303))
	assert.Equal(t, "00:00", formatDuration(0))
	assert.Equal(t, "59:59", formatDuration(3599))
}

func TestFormatDuration_OverOneHour(t *testing.T) {
	assert.Equal(t, "1:00:00", formatDuration(3600))
	assert.Equal(t, "1:02:03", formatDuration(3723))
	assert.Equal(t, "2:00:05", formatDuration(7205))
}

func TestMapCategory_ByTitleKeyword(t *testing.T) {
	tests := []struct {
		title string
		want  int
	}{
		{"GPT 大模型入门", 1},
		{"ChatGPT 实战", 1},
		{"芯片拆解: 半导体工艺", 2},
		{"微积分详解", 3},
		{"新概念英语第三册", 4},
		{"篮球场训练技巧", 5},
		{"如何缓解焦虑与抑郁", 6},
		{"Python 前端开发入门", 7},
		{"显卡评测", 8},
		{"量子力学基础", 9},
		{"发动机工作原理: 机械工程", 10},
		{"手机 5G 数码评测", 11},
		{"国际政治与选举", 12},
		{"毛泽东历史与古代王朝", 13},
		{"股票投资: 财经房价", 14},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.want, mapCategory("", tt.title), "title=%s", tt.title)
		})
	}
}

func TestMapCategory_ByPartitionName(t *testing.T) {
	for _, tname := range []string{"动画", "鬼畜", "影视", "音乐", "舞蹈", "生活", "美食", "搞笑", "娱乐"} {
		assert.Equal(t, 11, mapCategory(tname, ""), "tname=%s", tname)
	}
}

func TestMapCategory_DefaultFallback(t *testing.T) {
	assert.Equal(t, 11, mapCategory("", "随便聊聊"))
	assert.Equal(t, 11, mapCategory("未知分区", "今天天气不错"))
}

func TestMapCategory_TitlePriorityOverPartition(t *testing.T) {
	// 标题命中 AI 关键词 → 1，即使分区名是"游戏"
	assert.Equal(t, 1, mapCategory("游戏", "chatgpt 新模型"))
}

func TestMapCategory_CaseInsensitive(t *testing.T) {
	assert.Equal(t, 4, mapCategory("", "ENGLISH Grammar"))
	assert.Equal(t, 7, mapCategory("", "LINUX 系统编程"))
}
