package model

type RevokeLinkResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		InviteLink string `json:"invite_link"`
		Creator    struct {
			Id        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"creator"`
		MemberLimit        int  `json:"member_limit"`
		CreatesJoinRequest bool `json:"creates_join_request"`
		IsPrimary          bool `json:"is_primary"`
		IsRevoked          bool `json:"is_revoked"`
	} `json:"result"`
}
