package web

// roleAssignsTurn reports whether outbound speech gets a global turn index (ADR-0013).
func roleAssignsTurn(role string) bool {
	return role == RoleModerator || role == RoleParticipant || role == RoleUser
}
