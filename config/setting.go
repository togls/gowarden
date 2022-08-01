package config

import "strings"

type Settings struct {
	Addr                     string `json:"addr"`
	Domain                   string `json:"domain"`
	DomainSet                bool
	DomainOrigin             string // extract_url_origin(c.domain)
	DomainPath               string // extract_url_path(c.domain)
	WebEnabled               bool
	SendsAllowed             bool `json:"sends_allowed"`
	HIBPApiKey               string
	UserAttachmentLimit      int
	OrgAttachmentLimit       int
	TrashAutoDelete          int
	Incomplete2faTimeLimit   int  `json:"incomplete_2fa_time_limit"`
	DisableIconDownload      bool `json:"disable_icon_download"`
	SignupsAllowed           bool `json:"signups_allowed"`
	SignupsVerify            bool `json:"signups_verify"`
	SignupsVerifyResendTime  int  `json:"signups_verify_resend_time"`
	SignupsVerifyResendLimit int  `json:"signups_verify_resend_limit"`
	SignupsDomainsWhitelist  string
	OrgCreationUsers         string
	InvitationsAllowed       bool   `json:"invitations_allowed"`
	EmergencyAccessAllowed   bool   `json:"emergency_access_allowed"`
	PasswordIterations       int    `json:"password_iterations"`
	ShowPasswordHint         bool   `json:"show_password_hint"`
	AdminToken               string `json:"admin_token"`
	InvitationOrgName        string `json:"invitation_org_name"`
}

func (s Settings) IsOrgCreationAllowed(email string) bool {
	ocu := s.OrgCreationUsers

	if ocu == "" || ocu == "all" {
		return true
	}

	if ocu == "none" {
		return false
	}

	users := strings.Split(ocu, ",")

	return any(users, email)
}

func (s Settings) IsInvitationsAllowed() bool {
	return s.InvitationsAllowed
}

func (s Settings) IsEmailDomainAllowed(email string) bool {
	if s.SignupsDomainsWhitelist == "" {
		return true
	}

	domains := strings.Split(s.SignupsDomainsWhitelist, ",")

	for _, domain := range domains {
		if strings.HasSuffix(email, "@"+strings.TrimSpace(domain)) {
			return true
		}
	}

	return false
}

func any(a []string, b string) bool {
	for _, v := range a {
		if strings.TrimSpace(v) == b {
			return true
		}
	}

	return false
}
