package models

// UserInfo represents the response from /oauth2/userinfo endpoint
type UserInfo struct {
	Sub                string      `json:"sub"`
	Name               string      `json:"name"`
	CPF                string      `json:"cpf"`
	Email              string      `json:"email,omitempty"`
	EmailVerified      bool        `json:"email_verified,omitempty"`
	PhoneNumber        string      `json:"phone_number,omitempty"`
	PhoneNumberVerified bool       `json:"phone_number_verified,omitempty"`
	Address            interface{} `json:"address,omitempty"`
	UnionUnit          interface{} `json:"union_unit,omitempty"`
	MembershipStatus   string      `json:"membership_status,omitempty"`
	EmploymentStatus   string      `json:"employment_status,omitempty"`
	MembershipType     string      `json:"membership_type,omitempty"`
	Permissions        []string    `json:"permissions,omitempty"`
}
