package models

type Params struct {
	Path        string   `json:"path"`
	Overwrite   bool     `json:"overwrite"`
	Fields      []string `json:"fields"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
	Sort        string   `json:"sort"`
	PreviewCrop bool     `json:"preview_crop"`
	PreviewSize string   `json:"preview_size"`
	Permanently bool     `json:"permanently"`
}
