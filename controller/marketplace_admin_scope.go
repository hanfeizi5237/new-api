package controller

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

func requireMarketplaceRiskAction(c *gin.Context, reason string) error {
	if !hasMarketplaceRiskScope(c) {
		return errors.New("insufficient privilege for risk action")
	}
	if !hasFreshSecureVerification(c) {
		return errors.New("需要安全验证")
	}
	return validateMarketplaceRiskReason(reason)
}

func hasMarketplaceRiskScope(c *gin.Context) bool {
	role := c.GetInt("role")
	if role >= common.RoleRootUser {
		return true
	}
	if role < common.RoleAdminUser {
		return false
	}
	return groupHasMarketplaceRiskScope(c.GetString("group"))
}

func groupHasMarketplaceRiskScope(group string) bool {
	normalized := strings.ToLower(strings.TrimSpace(group))
	if normalized == "" {
		return false
	}
	tokens := strings.FieldsFunc(normalized, func(r rune) bool {
		switch r {
		case ',', ';', '|', '/', ' ':
			return true
		default:
			return false
		}
	})
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		switch trimmed {
		case "market_risk", "risk_admin", "risk":
			return true
		}
		if strings.Contains(trimmed, "risk") {
			return true
		}
	}
	return false
}
