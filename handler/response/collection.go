package response

import "github.com/togls/gowarden/model"

type Collection struct {
	ExternalId     any    `json:"ExternalId"`
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	Object         string `json:"Object"`
	OrganizationId string `json:"OrganizationId"`
}

func NewCollection(collection *model.Collection) *Collection {
	return &Collection{
		ID:             collection.Uuid,
		Name:           collection.Name,
		OrganizationId: collection.OrgUuid,
		ExternalId:     nil,
		Object:         "collection",
	}
}

func NewCollections(collections model.CollectionList) []*Collection {
	var result []*Collection
	for _, collection := range collections {
		result = append(result, NewCollection(collection))
	}
	return result
}

type CollectionDetails struct {
	*Collection
	ReadOnly      bool `json:"ReadOnly"`
	HidePasswords bool `json:"HidePasswords"`
}

func NewCollectionDetails(collection *model.Collection) *CollectionDetails {
	details := &CollectionDetails{
		Collection:    NewCollection(collection),
		ReadOnly:      collection.ReadOnly,
		HidePasswords: collection.HidePasswords,
	}

	details.Object = "collectionDetails"

	return details
}

func NewCollectionDetailsList(collections model.CollectionList) []*CollectionDetails {
	var result []*CollectionDetails
	for _, collection := range collections {
		result = append(result, NewCollectionDetails(collection))
	}
	return result
}
