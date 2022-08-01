package response

import (
	"time"

	"github.com/togls/gowarden/model"
)

type Folder struct {
	Id           string    `json:"Id"`
	RevisionDate time.Time `json:"RevisionDate"`
	Name         string    `json:"Name"`
	Object       string    `json:"Object"`
}

func NewFolder(folder *model.Folder) *Folder {
	return &Folder{
		Id:           folder.Uuid,
		RevisionDate: folder.UpdatedAt,
		Name:         folder.Name,
		Object:       "folder",
	}
}

func NewFolders(folders []*model.Folder) []*Folder {
	var result []*Folder
	for _, folder := range folders {
		result = append(result, NewFolder(folder))
	}
	return result
}
