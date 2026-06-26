package discord

import "strings"

var meetStartTriggers = map[string]struct{}{
	"开始会议": {},
	"新会议":  {},
	"会议开始": {},
}

var meetCancelTriggers = map[string]struct{}{
	"取消会议": {},
}

func isMeetStartTrigger(content string) bool {
	_, ok := meetStartTriggers[strings.TrimSpace(content)]
	return ok
}

func isMeetCancelTrigger(content string) bool {
	_, ok := meetCancelTriggers[strings.TrimSpace(content)]
	return ok
}
