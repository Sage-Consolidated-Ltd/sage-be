package domain_test

import (
	"strings"
	"testing"

	"sage-backend/internal/identity/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityDomain_EmailValueObject(t *testing.T) {
	t.Run("Valid emails pass and normalize to lowercase", func(t *testing.T) {
		validCases := []struct {
			input    string
			expected string
		}{
			{"user@example.com", "user@example.com"},
			{"User.Name+filter@SAGE.io", "user.name+filter@sage.io"},
			{"support_123@sub.domain.org", "support_123@sub.domain.org"},
			{"first.last@company.co.uk", "first.last@company.co.uk"},
			{"  trimmed@example.com  ", "trimmed@example.com"},
		}

		for _, tc := range validCases {
			email, err := domain.NewEmail(tc.input)
			require.NoError(t, err, "expected valid: %s", tc.input)
			assert.Equal(t, tc.expected, email.String())
			assert.False(t, email.IsZero())
		}
	})

	t.Run("Invalid emails fail pattern and structural validation", func(t *testing.T) {
		invalidCases := []string{
			"",
			"   ",
			"plainaddress",
			"@missinglocal.com",
			"missingdomain@",
			"missingatsign.com",
			"user@localhost",            // no TLD
			"user@internal_domain",      // no dot or TLD
			"user@domain",               // no TLD
			"user@.com",                 // empty domain label
			"user@com.",                 // trailing dot in domain
			".leadingdot@example.com",   // leading dot in local part
			"trailingdot.@example.com",  // trailing dot in local part
			"double..dot@example.com",   // consecutive dots in local part
			"user@double..dot.com",      // consecutive dots in domain
			"user@domain.c",             // single char TLD
			"user@-domain.com",          // leading dash in domain
			"user@domain-.com",          // trailing dash in domain
			"spaces in@example.com",     // space in local part
			"user@exam ple.com",         // space in domain
			strings.Repeat("a", 65) + "@example.com", // local part > 64 chars
			"user@" + strings.Repeat("a", 250) + ".com", // total > 254 chars
		}

		for _, invalid := range invalidCases {
			_, err := domain.NewEmail(invalid)
			assert.Error(t, err, "expected invalid email to be rejected: %s", invalid)
		}
	})
}
