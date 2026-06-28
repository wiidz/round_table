package discord

import "errors"

var (
	errReceptionNeedDisplayName    = errors.New("请说明新专家的显示名，例如：新增专家「玩家代表小美」")
	errReceptionNeedParticipantRef = errors.New("请指定要操作的专家代号或名称")
	errReceptionNeedTopic          = errors.New("请说明会议主题，例如：开个会聊聊骑士职业设计")
	errReceptionNeedProfileFile    = errors.New("请指定档案文件：SOUL、AGENTS 或 TOOLS")
	errReceptionNeedProfileContent = errors.New("need profile content") // internal; user sees ask prompt
)
