package webhook

import "testing"

func TestNotionPageURL(t *testing.T) {
	cases := []struct {
		name   string
		pageID string
		want   string
	}{
		{
			name:   "dashed uuid from the notion api",
			pageID: "2ce64b29-b84c-8164-b9f5-ce5a4797ada5",
			want:   "https://www.notion.so/2ce64b29b84c8164b9f5ce5a4797ada5",
		},
		{
			name:   "already undashed",
			pageID: "2ce64b29b84c8164b9f5ce5a4797ada5",
			want:   "https://www.notion.so/2ce64b29b84c8164b9f5ce5a4797ada5",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := notionPageURL(tc.pageID); got != tc.want {
				t.Errorf("notionPageURL(%q) = %q, want %q", tc.pageID, got, tc.want)
			}
		})
	}
}
