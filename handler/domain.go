package handler

import (
	"encoding/json"
	"fmt"

	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler/response"
	"github.com/togls/gowarden/model"
)

func getEqDomains(globals config.GlobalDomains, user *model.User, noExcluded bool) (*response.Domains, error) {
	eq := new([][]string)

	if err := json.Unmarshal([]byte(user.EquivalentDomains), eq); err != nil {
		return nil, err
	}

	ex := new([]string)
	if err := json.Unmarshal([]byte(user.ExcludedGlobals), ex); err != nil {
		return nil, err
	}

	isContained := func(arr []string, str string) bool {
		for _, v := range arr {
			if v == str {
				return true
			}
		}
		return false
	}

	list := config.GlobalDomains{}
	for i := range globals {
		contained := isContained(*ex, fmt.Sprintf("%d", globals[i].Type))
		if noExcluded && !contained {
			continue
		}

		globals[i].Excluded = contained

		list = append(list, globals[i])
	}

	return &response.Domains{
		EquivalentDomains:       *eq,
		GlobalEquivalentDomains: list,
		Object:                  "domains",
	}, nil
}
