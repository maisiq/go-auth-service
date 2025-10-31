// Config for Yandex OAuth provider

package configs

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

type YandexProviderConfig struct {
	GetTokenURL  string `mapstructure:"token_url"`
	UserInfoURL  string `mapstructure:"user_url"`
	AuthorizeURL string `mapstructure:"auth_url"`
	RedirectURL  string `mapstructure:"redirect_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Scope        string `mapstructure:"scope"`
}

func (c *YandexProviderConfig) GetAuthorizeURL(state string) string {
	query := url.Values{}
	query.Set("client_id", c.ClientID)
	query.Set("redirect_uri", c.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", c.Scope)
	query.Set("state", state)
	return fmt.Sprintf(c.AuthorizeURL+"?%s", query.Encode())
}

func (c *YandexProviderConfig) GetTokenRequestPayload(authorizationCode string) io.Reader {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authorizationCode)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	return strings.NewReader(data.Encode())
}
