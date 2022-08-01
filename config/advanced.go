package config

type Advanced struct {
	IPHeader                  string `json:"ip_header"`
	IPHeaderEnabled           bool
	IconService               string
	IconRedirectCode          int `json:"icon_redirect_code"`
	IconCacheTTL              int `json:"icon_cache_ttl"`
	IconCacheNegttl           int `json:"icon_cache_negttl"`
	IconDownloadTimeout       int `json:"icon_download_timeout"`
	IconBlacklistRegex        string
	IconBlacklistNonGlobalIPs bool `json:"icon_blacklist_non_global_ips"`

	Disable2faRemember            bool `json:"disable_2fa_remember"`
	AuthenticatorDisableTimeDrift bool `json:"authenticator_disable_time_drift"`
	RequireDeviceEmail            bool `json:"require_device_email"`
	ReloadTemplates               bool `json:"reload_templates"`

	ExtendedLogging    bool
	LogTimestampFormat string `json:"log_timestamp_format"`
	UseSyslog          bool
	LogFile            string
	Log_level          string

	EnableDBWal         bool
	DBConnectionRetries int
	DatabaseTimeout     int
	DatabaseMaxConns    int

	DisableAdminToken      bool `json:"disable_admin_token"`
	AllowedIframeAncestors string

	LoginRatelimitSeconds  int
	LoginRatelimitMaxBurst int
	AdminRatelimitSeconds  int
	AdminRatelimitMaxBurst int
}

type IframeAncestorsGetter interface {
	IframeAncestors() string
}

func (a Advanced) IframeAncestors() string {
	return a.AllowedIframeAncestors
}
