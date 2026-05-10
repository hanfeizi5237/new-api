package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

var sellerSecretLiveProbeHTTPClient = &http.Client{Timeout: 5 * time.Second}

func probeSellerSecretProviderLive(secret *model.SellerSecret, runtimeKey string) error {
	channel, err := resolveSellerSecretProbeChannel(secret)
	if err != nil {
		return err
	}
	if !supportsSellerSecretLiveProbe(secret, channel) {
		return nil
	}

	endpoint, err := buildSellerSecretProbeModelsURL(channel)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(runtimeKey))
	req.Header.Set("Accept", "application/json")

	client := sellerSecretLiveProbeHTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("seller secret live probe failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = resp.Status
	}
	return fmt.Errorf("seller secret live probe failed: %s", message)
}

func resolveSellerSecretProbeChannel(secret *model.SellerSecret) (*model.Channel, error) {
	if secret == nil {
		return nil, fmt.Errorf("seller secret is required")
	}
	bindings, err := getActiveSupplyBindingsTx(model.DB, secret.SupplyAccountId)
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return nil, fmt.Errorf("no active supply channel bindings available for live probe")
	}
	channel, err := model.GetChannelById(bindings[0].ChannelId, true)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func supportsSellerSecretLiveProbe(secret *model.SellerSecret, channel *model.Channel) bool {
	if secret == nil || channel == nil {
		return false
	}
	if strings.TrimSpace(channel.GetBaseURL()) == "" {
		return false
	}
	switch strings.TrimSpace(secret.ProviderCode) {
	case "", "openai", "openai_compatible":
		return true
	}
	switch channel.Type {
	case constant.ChannelTypeOpenAI, constant.ChannelTypeOpenAIMax, constant.ChannelTypeOpenRouter, constant.ChannelTypeCustom, constant.ChannelTypeAIProxy, constant.ChannelTypeAIProxyLibrary, constant.ChannelTypeOhMyGPT:
		return true
	default:
		return false
	}
}

func buildSellerSecretProbeModelsURL(channel *model.Channel) (string, error) {
	if channel == nil {
		return "", fmt.Errorf("probe channel is required")
	}
	baseURL := strings.TrimSpace(channel.GetBaseURL())
	if baseURL == "" {
		return "", fmt.Errorf("probe channel %d has no base URL", channel.Id)
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/models", nil
	}
	return baseURL + "/v1/models", nil
}

func SetSellerSecretLiveProbeHTTPClient(client *http.Client) {
	if client == nil {
		sellerSecretLiveProbeHTTPClient = &http.Client{Timeout: 5 * time.Second}
		return
	}
	sellerSecretLiveProbeHTTPClient = client
}

func SetSellerSecretLiveProbeFunc(fn func(secret *model.SellerSecret, runtimeKey string) error) {
	if fn == nil {
		sellerSecretLiveProbeFunc = probeSellerSecretProviderLive
		return
	}
	sellerSecretLiveProbeFunc = fn
}
