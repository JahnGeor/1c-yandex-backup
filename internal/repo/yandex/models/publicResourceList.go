package models

type PublicResourceList struct {
	Items  []Resource `json:"items"`
	Type   string     `json:"type"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}
