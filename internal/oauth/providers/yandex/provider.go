package yandex

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/oauth"
)

var ErrInvalidAuthCode = fmt.Errorf("invalid auth code")

// more at https://yandex.ru/dev/id/doc/ru/codes/screen-code-oauth#token-request
type successTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Expires      uint64 `json:"expires_in"`
}

type errorTokenResponse struct {
	Error string `json:"error"`
}

type userData struct {
	ID    string `json:"id"`
	Email string `json:"default_email"`
}

type YandexOAuthProvider struct {
	cfg    *configs.YandexProviderConfig
	client *http.Client
}

func NewYandexOAuthProvider(cfg *configs.YandexProviderConfig, client *http.Client) *YandexOAuthProvider {
	return &YandexOAuthProvider{
		cfg:    cfg,
		client: client,
	}
}

func (p *YandexOAuthProvider) GetData(authorizationCode string) (oauth.OAuthData, error) {
	var data oauth.OAuthData
	token, err := p.accessToken(authorizationCode)
	if err != nil {
		return oauth.OAuthData{}, err
	}
	udata, err := p.getUserData(token)
	if err != nil {
		return data, err
	}
	return oauth.OAuthData{
		UserID:   udata.ID,
		Email:    udata.Email,
		Provider: oauth.YandexProvider,
	}, nil
}

func (p *YandexOAuthProvider) getUserData(accessToken string) (*userData, error) {
	req, _ := http.NewRequest(http.MethodGet, p.cfg.UserInfoURL, nil)

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	var data userData

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (p *YandexOAuthProvider) accessToken(authorizationCode string) (string, error) {
	r, err := p.client.Post(
		p.cfg.GetTokenURL,
		"application/x-www-form-urlencoded",
		p.cfg.GetTokenRequestPayload(authorizationCode),
	)
	if err != nil {
		return "", err
	}
	if r.StatusCode != 200 {
		var response errorTokenResponse
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			return "", err
		}
		switch response.Error {
		case "invalid_grant":
		case "bad_verification_code":
			return "", ErrInvalidAuthCode
		}
		return "", fmt.Errorf("failed to retrieve the token: %s", response.Error)
	}

	var response successTokenResponse

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return "", err
	}
	return response.AccessToken, nil
}
