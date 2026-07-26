package service

import "testing"

func TestNormalizeSiteName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: DefaultSiteName},
		{name: "legacy default", input: "Sub2API", want: DefaultSiteName},
		{name: "legacy default case insensitive", input: " sub2api ", want: DefaultSiteName},
		{name: "custom name", input: "Acme AI", want: "Acme AI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeSiteName(tt.input); got != tt.want {
				t.Fatalf("NormalizeSiteName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSiteSubtitle(t *testing.T) {
	if got := NormalizeSiteSubtitle("Subscription to API Conversion Platform"); got != DefaultSiteSubtitle {
		t.Fatalf("legacy subtitle normalized to %q, want %q", got, DefaultSiteSubtitle)
	}
	if got := NormalizeSiteSubtitle("统一模型接入"); got != "统一模型接入" {
		t.Fatalf("custom subtitle changed to %q", got)
	}
}
