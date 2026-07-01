package brief

import "time"

const (
	SourceBuiltin = "builtin"
	SourceCustom  = "custom"
	FileBrief     = "BRIEF.yaml"
)

// Meta holds human-facing template metadata.
type Meta struct {
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Owner       string `yaml:"owner,omitempty" json:"owner,omitempty"`
}

// BriefBody is the Principal-authored meeting task book.
type BriefBody struct {
	Goal         string   `yaml:"goal,omitempty" json:"goal,omitempty"`
	Agenda       []string `yaml:"agenda,omitempty" json:"agenda,omitempty"`
	InScope      string   `yaml:"in_scope,omitempty" json:"in_scope,omitempty"`
	OutOfScope   string   `yaml:"out_of_scope,omitempty" json:"out_of_scope,omitempty"`
	DoneCriteria string   `yaml:"done_criteria,omitempty" json:"done_criteria,omitempty"`
}

// MeetingDefaults are optional launch defaults for a template.
type MeetingDefaults struct {
	Mode                       string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	MaxRounds                  int      `yaml:"max_rounds,omitempty" json:"max_rounds,omitempty"`
	MinRoundsBeforeSynthesis   int      `yaml:"min_rounds_before_synthesis,omitempty" json:"min_rounds_before_synthesis,omitempty"`
	ConfirmationMode           string   `yaml:"confirmation_mode,omitempty" json:"confirmation_mode,omitempty"`
	FreeDialogueMaxQuestions   int      `yaml:"free_dialogue_max_questions,omitempty" json:"free_dialogue_max_questions,omitempty"`
	ParticipantIDs             []string `yaml:"participant_ids,omitempty" json:"participant_ids,omitempty"`
}

// Document is the parsed BRIEF.yaml schema (ADR-0014).
type Document struct {
	Meta    Meta            `yaml:"meta" json:"meta"`
	Topic   string          `yaml:"topic,omitempty" json:"topic,omitempty"`
	Brief   BriefBody       `yaml:"brief" json:"brief"`
	Meeting MeetingDefaults `yaml:"meeting,omitempty" json:"meeting,omitempty"`
}

// LaunchDraft is the resolved launch configuration for Transport / Web.
type LaunchDraft struct {
	Topic   string          `json:"topic"`
	Brief   BriefBody       `json:"brief"`
	Meeting MeetingDefaults `json:"meeting"`
}

// TemplateIndex summarizes one template directory.
type TemplateIndex struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Source      string    `json:"source"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TemplateDetail includes raw YAML and parsed document.
type TemplateDetail struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Source      string      `json:"source"`
	Content     string      `json:"content"`
	Document    Document    `json:"document"`
	Launch      LaunchDraft `json:"launch"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
