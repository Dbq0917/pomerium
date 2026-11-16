package mcp

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/caddyserver/certmagic"
)

// ErrDomainNotAllowed is returned when a client ID URL's domain is not in the allowed list.
var ErrDomainNotAllowed = errors.New("client_id domain not allowed")

// DomainMatcher checks if domains match against a list of allowed patterns.
type DomainMatcher struct {
	allowedDomains []string
}

// NewDomainMatcher creates a new DomainMatcher with the given allowed domain patterns.
// Patterns support wildcards like "*.example.com".
func NewDomainMatcher(allowedDomains []string) *DomainMatcher {
	return &DomainMatcher{allowedDomains: allowedDomains}
}

// IsAllowed checks if the given hostname matches any of the allowed domain patterns.
// A leading wildcard (e.g., "*.github.com") matches any number of subdomain
// labels, but not the bare domain. Other patterns use certmagic.MatchWildcard.
func (m *DomainMatcher) IsAllowed(hostname string) bool {
	for _, pattern := range m.allowedDomains {
		if certmagic.MatchWildcard(hostname, pattern) {
			return true
		}
		if matchWildcardSuffix(hostname, pattern) {
			return true
		}
	}
	return false
}

// matchWildcardSuffix matches a leading-wildcard pattern against any number of
// subdomain labels. certmagic.MatchWildcard only replaces one label at a time,
// so it does not match nested subdomains such as foo.api.github.com.
func matchWildcardSuffix(hostname, pattern string) bool {
	suffix, ok := strings.CutPrefix(pattern, "*")
	if !ok || !strings.HasPrefix(suffix, ".") || len(suffix) == 1 {
		return false
	}

	hostname = strings.ToLower(hostname)
	suffix = strings.ToLower(suffix)
	if !strings.HasSuffix(hostname, suffix) {
		return false
	}

	// Require at least one non-empty label before the suffix.
	prefix := hostname[:len(hostname)-len(suffix)]
	return prefix != "" && !strings.HasSuffix(prefix, ".")
}

// ValidateURLDomain checks if the URL's hostname is in the allowed domains list.
// Returns an error if the domain is not allowed.
func (m *DomainMatcher) ValidateURLDomain(u *url.URL) error {
	hostname := u.Hostname()
	if !m.IsAllowed(hostname) {
		return fmt.Errorf("%w: %q is not in allowed domains", ErrDomainNotAllowed, hostname)
	}
	return nil
}
