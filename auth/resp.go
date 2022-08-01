package auth

type RespRefreshToken struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	RefreshToken string  `json:"refresh_token"`
	Key          *string `json:"Key"`
	PrivateKey   *string `json:"PrivateKey"`

	Kdf              int    `json:"Kdf"`
	KdfIterations    int    `json:"KdfIterations"`
	Scope            string `json:"scope"`
	UnofficialServer bool   `json:"unofficialServer"`

	ResetMasterPassword bool `json:"ResetMasterPassword"`
}
