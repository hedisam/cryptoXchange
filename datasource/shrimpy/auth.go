package shrimpy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func getToken(cfg *shrimpyConfig) (string, error) {
	resp, err := Get(cfg, tokenPath)
	if err != nil {
		return "", fmt.Errorf("[getToken] request failed: %w", err)
	}

	var token wsToken
	err = json.Unmarshal(resp, &token)
	if err != nil {
		return "", fmt.Errorf("[getToken] failed to unmarshal the wsToken: %w", err)
	}

	return token.Token, err
}

func createSignature(secretKey, requestPath string, method string, body []byte) (string, int64, error) {
	nonce := time.Now().Unix()

	var bodyStr string
	if body != nil {
		bodyStr = string(body)
	}

	preHash := fmt.Sprint(requestPath, method, nonce, bodyStr)
	secretKeyDecoded, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("[createSignature] decoding secret key failed: %w", err)
	}

	h := hmac.New(sha256.New, secretKeyDecoded)
	_, err = h.Write([]byte(preHash))
	if err != nil {
		return "", 0, fmt.Errorf("[createSignature] failed to generate the hmac sha256 signature: %w", err)
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nonce, nil
}
