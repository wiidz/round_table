package workspace

// PaginatedMeetings is a page of meeting indexes plus total count.
type PaginatedMeetings struct {
	Meetings []MeetingIndex `json:"meetings"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
