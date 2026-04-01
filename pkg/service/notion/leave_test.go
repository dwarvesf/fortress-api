package notion

import (
	"testing"

	nt "github.com/dstotijn/go-notion"
)

func TestBuildContractorEmailLookupFilter(t *testing.T) {
	t.Parallel()

	email := "chauptb@d.foundation"
	filter := buildContractorEmailLookupFilter(email)

	if filter == nil {
		t.Fatal("expected filter")
	}

	if len(filter.Or) != 2 {
		t.Fatalf("expected 2 OR clauses, got %d", len(filter.Or))
	}

	assertEmailClause := func(t *testing.T, clause nt.DatabaseQueryFilter, property string) {
		t.Helper()

		if clause.Property != property {
			t.Fatalf("expected property %q, got %q", property, clause.Property)
		}

		if clause.DatabaseQueryPropertyFilter.Email == nil {
			t.Fatalf("expected email filter for property %q", property)
		}

		if clause.DatabaseQueryPropertyFilter.Email.Equals != email {
			t.Fatalf("expected email %q for property %q, got %q", email, property, clause.DatabaseQueryPropertyFilter.Email.Equals)
		}
	}

	assertEmailClause(t, filter.Or[0], "Team Email")
	assertEmailClause(t, filter.Or[1], "Personal Email")
}

func TestContractorDetailsPreferredEmailSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		details  ContractorDetails
		expected string
	}{
		{
			name: "prefers team email",
			details: ContractorDetails{
				TeamEmail:     "team@d.foundation",
				PersonalEmail: "personal@example.com",
			},
			expected: "team@d.foundation",
		},
		{
			name: "falls back to personal email",
			details: ContractorDetails{
				PersonalEmail: "personal@example.com",
			},
			expected: "personal@example.com",
		},
		{
			name:     "returns empty when no email available",
			details:  ContractorDetails{},
			expected: "",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.details.TeamEmail
			if got == "" {
				got = tc.details.PersonalEmail
			}

			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
