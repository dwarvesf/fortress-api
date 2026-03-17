package workupdates

import "testing"

func TestShouldIncludeContractorInMissingReport(t *testing.T) {
	tests := []struct {
		name            string
		includeComplete bool
		result          deploymentResult
		want            bool
	}{
		{
			name:            "exclude complete contractor by default",
			includeComplete: false,
			result:          deploymentResult{MissingDates: nil},
			want:            false,
		},
		{
			name:            "include complete contractor in full mode",
			includeComplete: true,
			result:          deploymentResult{MissingDates: nil},
			want:            true,
		},
		{
			name:            "include contractor with missing dates",
			includeComplete: false,
			result:          deploymentResult{MissingDates: []string{"1/2"}},
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIncludeContractorInMissingReport(tt.result, tt.includeComplete)
			if got != tt.want {
				t.Fatalf("shouldIncludeContractorInMissingReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
