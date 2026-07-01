package brief

// Port abstracts Meeting Brief Template storage (ADR-0014).
type Port interface {
	ListTemplates() ([]TemplateIndex, error)
	ReadTemplate(id string) (TemplateDetail, error)
	WriteTemplate(id string, content []byte) error
	CreateTemplate(content []byte) (string, error)
	CloneFromMeetingDoc(meetingDoc string) (LaunchDraft, error)
}
