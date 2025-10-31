package repository

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/logger"
)

type VaultSecretRepository struct {
	cfg    *configs.VaultConfig
	client *http.Client
}

func NewVaultSecretRepository(cfg *configs.VaultConfig, client *http.Client) *VaultSecretRepository {
	return &VaultSecretRepository{
		cfg:    cfg,
		client: client,
	}
}

func (r *VaultSecretRepository) getKeyData(ctx context.Context, keyName string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		r.cfg.BaseURL+fmt.Sprintf("/v1/transit/keys/%s", keyName),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.cfg.TOKEN)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("failed to decode response from vault: %w", err)
		}
		return nil, fmt.Errorf("bad request to vault: %+v", data)
	}
	var data map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response from vault: %w", err)
	}
	return data, nil
}

func (r *VaultSecretRepository) GetKID(ctx context.Context, keyName string) (string, error) {
	data, err := r.getKeyData(ctx, keyName)
	if err != nil {
		return "", err
	}

	if d, ok := data["data"].(map[string]interface{}); !ok {
		return "", fmt.Errorf("unpredictable response from vault")
	} else {
		ver := d["latest_version"]
		version := fmt.Sprintf("%.f", ver)
		return version, nil
	}
}

func (r *VaultSecretRepository) GetPublicKeys(ctx context.Context, keyName string) (map[string]string, error) {
	keys := make(map[string]string)

	data, err := r.getKeyData(ctx, keyName)
	if err != nil {
		return nil, err
	}
	d, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unpredictable response from vault")
	}
	keysData := d["keys"].(map[string]interface{})

	for k, v := range keysData {
		pk := v.(map[string]interface{})["public_key"]
		keys[k] = pk.(string)
	}

	if len(keys) < 1 {
		return nil, fmt.Errorf("no keys in vault")
	}
	return keys, nil
}

type ECDSASignature struct {
	R *big.Int
	S *big.Int
}

func (r *VaultSecretRepository) derToRaw(derSig []byte) ([]byte, error) {
	var sig ECDSASignature
	_, err := asn1.Unmarshal(derSig, &sig)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling ASN.1 signature: %w", err)
	}

	size := 32
	rBytes := sig.R.Bytes()
	sBytes := sig.S.Bytes()

	rPadded := make([]byte, size)
	sPadded := make([]byte, size)

	copy(rPadded[size-len(rBytes):], rBytes)
	copy(sPadded[size-len(sBytes):], sBytes)

	return append(rPadded, sPadded...), nil
}

func (r *VaultSecretRepository) SignJWT(ctx context.Context, data string, keyName string) (string, error) {
	b64data := base64.StdEncoding.EncodeToString([]byte(data))

	payload, err := json.Marshal(map[string]string{"input": b64data})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.cfg.BaseURL+fmt.Sprintf("/v1/transit/sign/%s", keyName),
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", r.cfg.TOKEN)

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		return "", fmt.Errorf("failed to sign jwt token: %s", string(data))
	}
	var respData map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if d, ok := respData["data"].(map[string]interface{}); !ok {
		return "", fmt.Errorf("unexpected response body from vault: %+v", err)
	} else {
		sig := d["signature"].(string)

		logger.GetLogger().Debugf("signature from vault: %s", sig)

		sigb64slice := strings.Split(sig, ":")
		sigb64 := sigb64slice[len(sigb64slice)-1]

		var sigDer []byte
		sigDer, err := base64.StdEncoding.DecodeString(sigb64)
		if err != nil {
			return "", fmt.Errorf("failed to decode sigDer: %w", err)
		}
		sigRaw, err := r.derToRaw(sigDer)
		if err != nil {
			return "", err
		}
		sigb64url := base64.RawURLEncoding.EncodeToString(sigRaw)
		return fmt.Sprintf("%s.%s", data, sigb64url), nil
	}
}
