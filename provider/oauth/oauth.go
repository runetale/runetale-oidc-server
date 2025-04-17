package oauth

type UserInfo struct {
	Name          string `json:"name,omitempty"`
	User          string `json:"user"`
	Picture       string `json:"picture,omitempty"`
	Subject       string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
