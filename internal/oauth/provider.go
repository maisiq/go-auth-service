package oauth

type OAuthData struct {
	UserID   string
	Email    string
	Provider OAuthProviderT
}

type OAuthProviderT string

const (
	YandexProvider OAuthProviderT = "yandex"
)

type OAuthProvider interface {
	GetData(authorizationCode string) (OAuthData, error)
}
