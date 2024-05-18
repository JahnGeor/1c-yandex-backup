package models

type Params struct {
	Path      string   `json:"path"`
	Overwrite bool     `json:"overwrite"`
	Fields    []string `json:"fields"`
}
