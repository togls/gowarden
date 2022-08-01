package model

type Collection struct {
	Uuid    string
	OrgUuid string
	Name    string

	UserUuid      string
	ReadOnly      bool
	HidePasswords bool
}

type UserCollection struct {
	CollectionUuid string
	UserUuid       string
	ReadOnly       bool
	HidePasswords  bool
}

type UCFilter struct {
	UserUuid       *string
	CollectionUuid *string
	OrgUuid        *string
}

// UCList is a list of UserCollection
type UCList []*UserCollection

type CipherCollection struct {
	CollectionUuid string
	CipherUuid     string
}

type CollectionFilter struct {
	UserUuid *string
	OrgUuid  *string
}

type CollectionList []*Collection

func (cl CollectionList) Len() int {
	return len(cl)
}

func (cl CollectionList) IDs() []string {
	ids := make([]string, len(cl))
	for i := range cl {
		ids[i] = cl[i].Uuid
	}
	return ids
}
