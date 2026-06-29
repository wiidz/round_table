package web

// Reply is an outbound browser chat message with display role.
type Reply struct {
	Content     string
	AsModerator bool // false → system bubble (platform/help/limitations)
}

func systemReply(content string) Reply {
	return Reply{Content: content, AsModerator: false}
}

func moderatorReply(content string) Reply {
	return Reply{Content: content, AsModerator: true}
}
