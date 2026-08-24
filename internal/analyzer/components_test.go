package analyzer

import "testing"

func TestMatchesObviousPatternUsesWordBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		patternKey string
		want       bool
	}{
		{"class matches keyword exactly", "div.card", true},
		{"keyword as a word inside a compound class", "div.product-card", true},
		{"multi word keyword", "li.nav-item", true},
		{"keyword in the id", "div.wrapper#main-modal", true},
		{"underscore separated class", "div.product_card", true},
		{"multiple classes", "button.btn.btn-primary", true},
		{"tab must not match table", "table.data-table", false},
		{"tab must not match portable", "div.portable", false},
		{"tag must not match stage", "div.stage", false},
		{"no class at all", "div", false},
		{"unrelated class", "div.layout-grid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesObviousPattern(tt.patternKey, obviousPatterns)
			if got != tt.want {
				t.Errorf("matchesObviousPattern(%q) = %v, want %v", tt.patternKey, got, tt.want)
			}
		})
	}
}

func TestAnalyzeComponentsSuggestsRepeatedElements(t *testing.T) {
	input := `<html><body>
		<button class="btn-primary">One</button>
		<button class="btn-primary">Two</button>
		<button class="btn-primary">Three</button>
	</body></html>`

	suggestions, err := AnalyzeComponents(input)
	if err != nil {
		t.Fatalf("AnalyzeComponents returned error: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d: %+v", len(suggestions), suggestions)
	}
	if suggestions[0].Count != 3 {
		t.Errorf("suggestion count = %d, want 3", suggestions[0].Count)
	}
	if suggestions[0].TagName != "button" {
		t.Errorf("suggestion tag = %q, want %q", suggestions[0].TagName, "button")
	}
}

func TestAnalyzeComponentsIgnoresPatternsBelowThreshold(t *testing.T) {
	input := `<html><body>
		<button class="btn-primary">One</button>
		<button class="btn-primary">Two</button>
	</body></html>`

	suggestions, err := AnalyzeComponents(input)
	if err != nil {
		t.Fatalf("AnalyzeComponents returned error: %v", err)
	}

	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions below the 3-occurrence threshold, got %+v", suggestions)
	}
}

func TestAnalyzeComponentsSkipsStructuralTags(t *testing.T) {
	input := `<html><body>
		<div class="card">One</div>
		<div class="card">Two</div>
		<div class="card">Three</div>
	</body></html>`

	suggestions, err := AnalyzeComponents(input)
	if err != nil {
		t.Fatalf("AnalyzeComponents returned error: %v", err)
	}

	if len(suggestions) != 0 {
		t.Errorf("expected structural tags to be skipped, got %+v", suggestions)
	}
}
