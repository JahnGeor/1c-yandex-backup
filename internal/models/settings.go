package models

type Settings struct {
	Token  string `json:"token,omitempty"`
	DBPath string `json:"db_path,omitempty"`
}
