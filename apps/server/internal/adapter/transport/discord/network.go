package discord

import (
	"time"
)

const maxSendRetries = 3

var sendRetryDelays = []time.Duration{
	200 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
}

func retrySend(op func() error) error {
	var last error
	for attempt := 0; attempt < maxSendRetries; attempt++ {
		if err := op(); err == nil {
			return nil
		} else {
			last = err
			if attempt < maxSendRetries-1 {
				time.Sleep(sendRetryDelays[attempt])
			}
		}
	}
	return last
}

func networkSendFailedText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ **网络异常** — 消息发送失败（已重试 3 次）。请稍后再试；若会议停在确认关，恢复连接后请重新发送指令。"
	}
	return "⚠️ **Network error** — message send failed after 3 retries. If the meeting is waiting for confirmation, resend your reply after reconnect."
}

func GatewayResumedText(loc Locale) string {
	return gatewayResumedText(loc)
}

func gatewayResumedText(loc Locale) string {
	if loc == LocaleZH {
		return "✅ **连接已恢复** — Discord 网关重连成功。若会议仍在进行或停在确认关，请重新发送上一条指令（如 **批准** / **驳回**）。"
	}
	return "✅ **Reconnected** — Discord gateway is back. If a meeting is still running or waiting for confirmation, resend your last command."
}
