package subtitle

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SRTCue struct {
	Index    int
	Start    time.Duration
	End      time.Duration
	Text     string
}

func ParseSRT(content string) ([]SRTCue, error) {
	content = strings.TrimSpace(strings.ReplaceAll(content, "\r\n", "\n"))
	blocks := strings.Split(content, "\n\n")
	var cues []SRTCue
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		lines := strings.SplitN(block, "\n", 3)
		if len(lines) < 3 {
			continue
		}
		index, err := strconv.Atoi(strings.TrimSpace(lines[0]))
		if err != nil {
			continue
		}
		start, end, err := parseTimeRange(strings.TrimSpace(lines[1]))
		if err != nil {
			continue
		}
		text := strings.TrimSpace(lines[2])
		cues = append(cues, SRTCue{Index: index, Start: start, End: end, Text: text})
	}
	return cues, nil
}

func parseTimeRange(s string) (time.Duration, time.Duration, error) {
	parts := strings.Split(s, " --> ")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time range: %s", s)
	}
	start, err := parseTime(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	end, err := parseTime(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func parseTime(s string) (time.Duration, error) {
	s = strings.ReplaceAll(s, ",", ".")
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time: %s", s)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	secParts := strings.Split(parts[2], ".")
	sec, err3 := strconv.Atoi(secParts[0])
	ms := 0
	if len(secParts) > 1 {
		ms, _ = strconv.Atoi(secParts[1])
	}
	// FINDING-011 修复：校验范围，防 time.Duration 溢出为负（如 parseTime("8000000::")）。
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("invalid time components: %s", s)
	}
	if h < 0 || h > 100 { // 长视频最多 100h，超出视为畸形
		return 0, fmt.Errorf("hours out of range: %s", s)
	}
	if m < 0 || m > 59 || sec < 0 || sec > 59 || ms < 0 || ms > 999 {
		return 0, fmt.Errorf("minutes/seconds/millis out of range: %s", s)
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second + time.Duration(ms)*time.Millisecond, nil
}

func (c SRTCue) ToCueMap() map[string]interface{} {
	return map[string]interface{}{
		"index":     c.Index,
		"startTime": c.Start.Seconds(),
		"endTime":   c.End.Seconds(),
		"text":      c.Text,
	}
}

func SRTCuesToJSON(cues []SRTCue) string {
	out := make([]map[string]interface{}, 0, len(cues))
	for _, c := range cues {
		out = append(out, c.ToCueMap())
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func (c SRTCue) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"index": c.Index,
		"start": c.Start.String(),
		"end":   c.End.String(),
		"text":  c.Text,
	}
}