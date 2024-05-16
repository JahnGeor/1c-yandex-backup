package models

type LastUploadedResourceList struct {
	Items []Resource `json:"items"`
	Limit int        `json:"limit"`
}
