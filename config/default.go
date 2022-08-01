package config

func defaultConfig() *Core {

	f := Folders{}

	return &Core{
		Folders: f,
		WS: WS{
			Enabled: false,
			Addres:  "0.0.0.0",
			Port:    3012,
		},
		Jobs: Jobs{
			PollInterval:                  30_000,
			SendPurge:                     "0 5 * * * *",
			TrashPurge:                    "0 5 0 * * *",
			Incomplete2fa:                 "30 * * * * *",
			EmergencyNotificationReminder: "0 5 * * * *",
			EmergencyRequestTimeout:       "0 5 * * * *",
		},
		Settings: Settings{
			Domain:    "http://localhost",
			DomainSet: false,
			// DomainOrigin:             "",
			// DomainPath:               "",
			WebEnabled:   true,
			SendsAllowed: true,
			HIBPApiKey:   "",
			// UserAttachmentLimit:      0,
			// OrgAttachmentLimit:       0,
			// TrashAutoDelete:          0,
			Incomplete2faTimeLimit:   3,
			DisableIconDownload:      false,
			SignupsAllowed:           true,
			SignupsVerify:            false,
			SignupsVerifyResendTime:  3_600,
			SignupsVerifyResendLimit: 6,
			SignupsDomainsWhitelist:  "",
			OrgCreationUsers:         "",
			InvitationsAllowed:       true,
			EmergencyAccessAllowed:   true,
			PasswordIterations:       100_000,
			ShowPasswordHint:         false,
			// AdminToken:               "",
			InvitationOrgName: "Vaultwarden",
		},
		Advanced: Advanced{
			IPHeader:         "X-Real-IP",
			IPHeaderEnabled:  false,
			IconService:      "internal",
			IconRedirectCode: 302,
			IconCacheTTL:     2_592_000,
			IconCacheNegttl:  259_200,
		},
	}
}
