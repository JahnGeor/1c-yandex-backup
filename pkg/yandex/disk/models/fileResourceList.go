package models

type FileResourceList struct {
	Items  []Resource `json:"items"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}
