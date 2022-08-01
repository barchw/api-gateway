package validation

import (
	"net/url"
	"regexp"
	"strings"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func hasDuplicates(rules []gatewayv1beta1.Rule) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range rules {
		encountered[rules[v].Path] = true
	}
	return len(encountered) != len(rules)
}

func isInvalidURL(toTest string) bool {
	if len(toTest) == 0 {
		return true
	}
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

func isUnsecuredURL(toTest string) bool {
	if len(toTest) == 0 {
		return false
	}
	return strings.HasPrefix(toTest, "http://")
}

//ValidateDomainName ?
func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$`)
	return RegExp.MatchString(domain)
}

//ValidateSubdomainName ?
func ValidateSubdomainName(subdomain string) bool {
	RegExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return RegExp.MatchString(subdomain)
}

//ValidateServiceName ?
func ValidateServiceName(service string) bool {
	regExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?\.[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return regExp.MatchString(service)
}

func validateGatewayName(gateway string) bool {
	regExp := regexp.MustCompile(`^[0-9a-z-_]+(\/[0-9a-z-_]+|(\.[0-9a-z-_]+)*)$`)
	return regExp.MatchString(gateway)
}
