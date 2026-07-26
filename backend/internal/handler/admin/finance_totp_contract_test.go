package admin

import (
	"reflect"
	"strings"
	"testing"
)

func TestFinanceAdminRequestsDoNotRequireTOTP(t *testing.T) {
	requests := []any{
		tierOverrideRequest{},
		distributionReversalRequest{},
		exchangeRateRequest{},
		distributionConfigRequest{},
		financialRuntimeConfigRequest{},
		distributionPolicyRequest{},
		withdrawalTransitionRequest{},
		voucherConfigRequest{},
		voucherRiskRequest{},
	}

	for _, request := range requests {
		typeOf := reflect.TypeOf(request)
		for i := 0; i < typeOf.NumField(); i++ {
			field := typeOf.Field(i)
			jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
			if field.Name == "TOTPCode" || jsonName == "totp_code" {
				t.Fatalf("%s must not require a TOTP field", typeOf.Name())
			}
		}
	}
}
