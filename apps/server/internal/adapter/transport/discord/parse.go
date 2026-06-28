package discord

import (
	"strings"

	"round_table/apps/server/internal/engine"
)

func parseParticipants(raw string) ([]engine.ParticipantInput, error) {
	return parseParticipantsFiltered(raw, nil)
}

func parseParticipantsFiltered(raw string, onlyIDs []string) ([]engine.ParticipantInput, error) {
	var out []engine.ParticipantInput
	allowed := make(map[string]struct{})
	for _, id := range onlyIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}
	filter := len(allowed) > 0
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		p, err := parseParticipantItem(item)
		if err != nil {
			return nil, err
		}
		if filter {
			if _, ok := allowed[p.ID]; !ok {
				continue
			}
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, errNoParticipants
	}
	return out, nil
}

func parseParticipantItem(item string) (engine.ParticipantInput, error) {
	first := strings.Index(item, ":")
	if first <= 0 || first >= len(item)-1 {
		return engine.ParticipantInput{}, errInvalidParticipant
	}
	id := item[:first]
	rest := item[first+1:]
	last := strings.LastIndex(rest, ":")
	if last <= 0 {
		return engine.ParticipantInput{ID: id, Role: rest}, nil
	}
	return engine.ParticipantInput{
		ID:        id,
		Role:      rest[:last],
		Expertise: rest[last+1:],
	}, nil
}

type meetParseResult struct {
	Topic string
	Mode  string
}

func parseMeetArgs(args []string, defaultMode string) (meetParseResult, error) {
	mode := defaultMode
	var topicParts []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-mode" || a == "--mode" {
			if i+1 >= len(args) {
				return meetParseResult{}, errMeetModeFlag
			}
			mode = args[i+1]
			i++
			continue
		}
		topicParts = append(topicParts, a)
	}
	topic := strings.TrimSpace(strings.Join(topicParts, " "))
	if topic == "" {
		return meetParseResult{}, errMeetTopicRequired
	}
	return meetParseResult{Topic: topic, Mode: mode}, nil
}
