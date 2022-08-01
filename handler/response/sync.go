package response

type Sync struct {
	Ciphers          []*Cipher            `json:"Ciphers"`
	Collections      []*CollectionDetails `json:"Collections"`
	Domains          *Domains             `json:"Domains"`
	Folders          []*Folder            `json:"Folders"`
	Object           string               `json:"Object"`
	Policies         []*Policy            `json:"Policies"`
	Profile          *Profile             `json:"Profile"`
	Sends            []*Send              `json:"Sends"`
	UnofficialServer bool                 `json:"unofficialServer"`
}

func NewSync(
	ciphers []*Cipher,
	collections []*CollectionDetails,
	domains *Domains,
	folders []*Folder,
	policies []*Policy,
	profile *Profile,
	sends []*Send,
) *Sync {
	return &Sync{
		Ciphers:     ciphers,
		Collections: collections,
		Domains:     domains,
		Folders:     folders,
		Policies:    policies,
		Profile:     profile,
		Sends:       sends,

		Object:           "sync",
		UnofficialServer: true,
	}
}
