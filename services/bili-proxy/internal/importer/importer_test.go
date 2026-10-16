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
		{"GPT 大模型入门", 231},
		{"ChatGPT 实战", 231},
		{"芯片拆解: 半导体工艺", 95},
		{"微积分详解", 201},
		{"新概念英语第三册", 208},
		{"篮球场训练技巧", 164},
		{"如何缓解焦虑与抑郁", 124},
		{"Python 前端开发入门", 231},
		{"显卡评测", 95},
		{"量子力学基础", 201},
		{"发动机工作原理: 机械工程", 232},
		{"手机 5G 数码评测", 95},
		{"国际政治与选举", 203},
		{"古代王朝历史人文", 228},
		{"股票投资: 财经房价", 207},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.want, mapCategory("", tt.title), "title=%s", tt.title)
		})
	}
}

func TestMapCategory_ByPartitionName(t *testing.T) {
	for _, tname := range []string{"动画", "鬼畜", "影视", "音乐", "舞蹈", "生活", "美食", "搞笑", "娱乐"} {
		want := tidByName[tname]
		assert.Equal(t, want, mapCategory(tname, ""), "tname=%s", tname)
	}
}

func TestMapCategory_DefaultFallback(t *testing.T) {
	assert.Equal(t, 188, mapCategory("", "随便聊聊"))
	assert.Equal(t, 188, mapCategory("未知分区", "今天天气不错"))
}

func TestMapCategory_TitlePriorityOverPartition(t *testing.T) {
	// 分区名精确匹配优先于标题关键词：tname=游戏 → 游戏分区(4)
	assert.Equal(t, 4, mapCategory("游戏", "chatgpt 新模型"))
	// 无有效分区名时，标题关键词生效
	assert.Equal(t, 231, mapCategory("", "chatgpt 新模型"))
}

func TestMapCategory_CaseInsensitive(t *testing.T) {
	assert.Equal(t, 208, mapCategory("", "ENGLISH Grammar"))
	assert.Equal(t, 231, mapCategory("", "LINUX 系统编程"))
}
