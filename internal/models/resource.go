package models

import "time"

type Resource struct {
	PublicKey string `json:"public_key"`
	Embedded  struct {
	} `json:"_embedded"`
	Name             string    `json:"name"`
	Created          time.Time `json:"created"`
	CustomProperties struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	} `json:"custom_properties"`
	PublicUrl  string    `json:"public_url"`
	OriginPath string    `json:"origin_path"`
	Modified   time.Time `json:"modified"`
	Path       string    `json:"path"`
	Md5        string    `json:"md5"`
	Type       string    `json:"type"`
	MimeType   string    `json:"mime_type"`
	Size       int       `json:"size"`
}
