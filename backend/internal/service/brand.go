package service

import "strings"

// DefaultSiteName is the product name shown in user-facing surfaces when an
// operator has not configured a custom site name.
const DefaultSiteName = "3API"

// DefaultSiteSubtitle is the neutral default subtitle used by public pages and
// authentication screens. Operators can still replace it in site settings.
const DefaultSiteSubtitle = "AI API gateway for unified model access"

const legacyDefaultSiteName = "Sub2API"

// NormalizeSiteName prevents the source project's historical default brand
// from leaking through public settings, emails, payment labels, or auth flows.
// Deliberately only maps the exact legacy default so a real operator-chosen
// custom name is not unexpectedly rewritten.
func NormalizeSiteName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" || strings.EqualFold(name, legacyDefaultSiteName) {
		return DefaultSiteName
	}
	return name
}

// NormalizeSiteSubtitle applies the same compatibility rule to the old
// generated subtitle while preserving operator-authored copy.
func NormalizeSiteSubtitle(raw string) string {
	subtitle := strings.TrimSpace(raw)
	if subtitle == "" || strings.EqualFold(subtitle, "Subscription to API Conversion Platform") {
		return DefaultSiteSubtitle
	}
	return subtitle
}
