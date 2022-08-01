package response

import "github.com/togls/gowarden/config"

type Domains struct {
	EquivalentDomains       [][]string
	GlobalEquivalentDomains config.GlobalDomains
	Object                  string `json:"object"`
}
