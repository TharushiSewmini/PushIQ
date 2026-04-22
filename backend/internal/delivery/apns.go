package delivery

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pushiq/pushiq-backend/internal/config"
	"github.com/pushiq/pushiq-backend/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

type APNSProvider struct {
	keyID  string
	teamID string
	topic  string
	signer *ecdsa.PrivateKey
	logger *logrus.Logger
	client *http.Client
}

func NewAPNSProvider(cfg *config.Config, logger *logrus.Logger) (*APNSProvider, error) {
	rawKey, err := os.ReadFile(cfg.APNSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read apns key: %w", err)
	}

	signer, err := parseAPNSPrivateKey(rawKey)
	if err != nil {
		return nil, err
	}

	transport := &http2.Transport{}
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second}
	return &APNSProvider{
		keyID:  cfg.APNSKeyID,
		teamID: cfg.APNSTeamID,
		topic:  cfg.APNSTopic,
		signer: signer,
		logger: logger,
		client: client,
	}, nil
}

func parseAPNSPrivateKey(raw []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("invalid apns private key data")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse apns p8 key: %w", err)
	}
	privKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("apns key is not ECDSA private key")
	}
	return privKey, nil
}

func (p *APNSProvider) buildJWT() (string, error) {
	header := map[string]string{"alg": "ES256", "kid": p.keyID}
	claims := map[string]any{"iss": p.teamID, "iat": time.Now().Unix()}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsigned := fmt.Sprintf("%s.%s", encodedHeader, encodedClaims)

	hash := sha256.Sum256([]byte(unsigned))
	signature, err := ecdsa.SignASN1(rand.Reader, p.signer, hash[:])
	if err != nil {
		return "", err
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	return fmt.Sprintf("%s.%s", unsigned, encodedSignature), nil
}

func (p *APNSProvider) Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error) {
	payload := map[string]any{
		"aps": map[string]any{
			"alert": map[string]string{
				"title": request.Title,
				"body":  request.Body,
			},
			"sound": "default",
		},
	}
	for k, v := range request.Data {
		payload[k] = v
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	jwt, err := p.buildJWT()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.push.apple.com/3/device/%s", request.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", fmt.Sprintf("bearer %s", jwt))
	req.Header.Set("apns-topic", p.topic)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("content-type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(respBody),
		}).Warn("APNs returned non-OK response")
		return nil, fmt.Errorf("apns server returned status %d", resp.StatusCode)
	}

	rawResponse := map[string]any{"body": string(respBody)}
	return &ProviderResponse{
		Provider:          model.ProviderAPNS,
		ProviderMessageID: "",
		RawResponse:       rawResponse,
	}, nil
}
