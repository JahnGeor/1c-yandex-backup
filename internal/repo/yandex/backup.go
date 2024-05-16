package yandex

type BackupYandex struct {
	Token string
}

func (y *BackupYandex) GetToken() string {
	return y.Token
}

func (y *BackupYandex) SetToken(token string) {
	y.Token = token
}

func NewBackupYandex(token string) *BackupYandex {
	return &BackupYandex{
		Token: token,
	}
}

func (y *BackupYandex) createResource(resourcePath string) error {

}

func (y *BackupYandex) createLink(resourceName string) error {

}

func (y *BackupYandex) uploadFile(link string, path string) error {

}
