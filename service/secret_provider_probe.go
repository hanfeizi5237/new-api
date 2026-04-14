package service

import "github.com/QuantumNous/new-api/model"

func probeSellerSecretProviderLive(secret *model.SellerSecret, runtimeKey string) error {
	return nil
}

func SetSellerSecretLiveProbeFunc(fn func(secret *model.SellerSecret, runtimeKey string) error) {
	if fn == nil {
		sellerSecretLiveProbeFunc = probeSellerSecretProviderLive
		return
	}
	sellerSecretLiveProbeFunc = fn
}
