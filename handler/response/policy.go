package response

import (
	"encoding/json"

	"github.com/togls/gowarden/model"
)

type Policy struct {
	Id             string          `json:"Id"`
	OrganizationId string          `json:"OrganizationId"`
	Type           int             `json:"Type"`
	Data           json.RawMessage `json:"Data"`
	Enabled        bool            `json:"Enabled"`
	Object         string          `json:"Object"`
}

func NewPolicy(policy *model.OrgPolicy) *Policy {
	return &Policy{
		Id:             policy.Uuid,
		OrganizationId: policy.OrgUuid,
		Type:           int(policy.Atype),
		Data:           policy.Data,
		Enabled:        policy.Enabled,
		Object:         "policy",
	}
}

func NewPolicies(policies []*model.OrgPolicy) []*Policy {
	var result []*Policy
	for _, policy := range policies {
		result = append(result, NewPolicy(policy))
	}
	return result
}
