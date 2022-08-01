package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type GlobalDomain struct {
	Type     int      `json:"Type"`
	Domains  []string `json:"Domains"`
	Excluded bool     `json:"Excluded"`
}

type GlobalDomains []GlobalDomain

func LoadGlobalDomain() (GlobalDomains, error) {
	fs, err := os.Open("./static/global_domains.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open global_domains.json: %w", err)
	}
	defer fs.Close()

	dec := json.NewDecoder(fs)

	var domains GlobalDomains
	if err := dec.Decode(&domains); err != nil {
		return nil, fmt.Errorf("failed to decode global domains: %w", err)
	}

	return domains, nil
}
