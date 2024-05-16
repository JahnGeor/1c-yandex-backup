package models

type Disk struct {
	TrashSize     int64 `json:"trash_size"`
	TotalSpace    int64 `json:"total_space"`
	UsedSpace     int64 `json:"used_space"`
	SystemFolders struct {
		Applications string `json:"applications"`
		Downloads    string `json:"downloads"`
	} `json:"system_folders"`
}
