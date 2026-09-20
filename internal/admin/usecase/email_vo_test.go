package usecase_test

import (
	"strings"
	"testing"

	"sage-backend/internal/admin/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminDomain_EmailValueObject(t *testing.T) {
	t.Run("Valid emails pass and normalize to lowercase", func(t *testing.T) {
		validCases := []struct {
			input    string
			expected string
		}{
			{"admin@sageconsolidated.com", "admin@sageconsolidated.com"},
			{"Super.Admin+filter@SAGE.io", "super.admin+filter@sage.io"},
			{"ops_support-123@sub.domain.org", "ops_support-123@sub.domain.org"},
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
			"admin@localhost",           // no TLD
			"admin@internal_domain",     // no dot or TLD
			"admin@domain",              // no TLD
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

	t.Run("Equals method works correctly", func(t *testing.T) {
		e1, err1 := domain.NewEmail("admin@sage.com")
		require.NoError(t, err1)
		e2, err2 := domain.NewEmail("ADMIN@sage.COM")
		require.NoError(t, err2)
		e3, err3 := domain.NewEmail("other@sage.com")
		require.NoError(t, err3)

		assert.True(t, e1.Equals(e2))
		assert.False(t, e1.Equals(e3))
	})
}
