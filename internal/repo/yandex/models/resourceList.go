package models

type ResourceList struct {
	Sort      string     `json:"sort"`
	PublicKey string     `json:"public_key"`
	Items     []Resource `json:"items"`
	Path      string     `json:"path"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	Total     int        `json:"total"`
}
