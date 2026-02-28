package common

import (
	"testing"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

func TestFormatTokensAndCost(t *testing.T) {
	// Create styles instance for testing
	tt := styles.DefaultStyles()

	tests := []struct {
		name             string
		info             *ModelContextInfo
		wantContains     []string // strings that should appear in output
		dontWantContains []string // strings that should NOT appear in output
	}{
		{
			name: "token display with 50% cache",
			info: &ModelContextInfo{
				TotalTokens:     100000,
				EffectiveTokens: 50000,
				CacheReadTokens: 50000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            1.25,
			},
			wantContains: []string{"100K total", "50K effective", "50% cached", "$1.25", "\n"}, // cache on separate line
		},
		{
			name: "large token values with millions",
			info: &ModelContextInfo{
				TotalTokens:     1500000,
				EffectiveTokens: 1000000,
				CacheReadTokens: 500000,
				CachePercent:    33.3,
				ModelContext:    2000000,
				Cost:            15.00,
			},
			wantContains: []string{"1.5M total", "1M effective", "33% cached", "\n"},
		},
		{
			name: "small token values",
			info: &ModelContextInfo{
				TotalTokens:     500,
				EffectiveTokens: 400,
				CacheReadTokens: 100,
				CachePercent:    20.0,
				ModelContext:    200000,
				Cost:            0.01,
			},
			wantContains: []string{"500 total", "400 effective", "20% cached", "\n"},
		},
		{
			name: "high context usage triggers warning with cache",
			info: &ModelContextInfo{
				TotalTokens:     170000,
				EffectiveTokens: 85000,
				CacheReadTokens: 85000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            2.00,
			},
			wantContains: []string{"W", "85%", "50% cached", "\n"}, // Warning "W", cache on separate line
		},
		{
			name: "no cache - single line",
			info: &ModelContextInfo{
				TotalTokens:     50000,
				EffectiveTokens: 50000,
				CacheReadTokens: 0,
				CachePercent:    0,
				ModelContext:    200000,
				Cost:            0.50,
			},
			wantContains:     []string{"25%", "50K total", "50K effective", "$0.50"},
			dontWantContains: []string{"cached", "\n"}, // no newline when no cache
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatTokensAndCost(&tt, tc.info)

			for _, want := range tc.wantContains {
				if !contains(result, want) {
					t.Errorf("Expected output to contain %q, got: %s", want, result)
				}
			}

			for _, dontWant := range tc.dontWantContains {
				if contains(result, dontWant) {
					t.Errorf("Expected output NOT to contain %q, got: %s", dontWant, result)
				}
			}
		})
	}
}

func TestModelContextInfoCalculations(t *testing.T) {
	tests := []struct {
		name             string
		promptTokens     int64
		completionTokens int64
		cacheReadTokens  int64
		wantTotal        int64
		wantEffective    int64
		wantCachePercent float64
	}{
		{
			name:             "no cache",
			promptTokens:     10000,
			completionTokens: 5000,
			cacheReadTokens:  0,
			wantTotal:        15000,
			wantEffective:    15000,
			wantCachePercent: 0,
		},
		{
			name:             "half cached",
			promptTokens:     10000,
			completionTokens: 5000,
			cacheReadTokens:  7500,
			wantTotal:        15000,
			wantEffective:    7500,
			wantCachePercent: 50.0,
		},
		{
			name:             "all cached",
			promptTokens:     10000,
			completionTokens: 5000,
			cacheReadTokens:  15000,
			wantTotal:        15000,
			wantEffective:    0,
			wantCachePercent: 100.0,
		},
		{
			name:             "partial cache",
			promptTokens:     85000,
			completionTokens: 15000,
			cacheReadTokens:  50000,
			wantTotal:        100000,
			wantEffective:    50000,
			wantCachePercent: 50.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total := tc.promptTokens + tc.completionTokens
			effective := total - tc.cacheReadTokens

			var cachePercent float64
			if total > 0 {
				cachePercent = (float64(tc.cacheReadTokens) / float64(total)) * 100
			}

			if total != tc.wantTotal {
				t.Errorf("Total: got %d, want %d", total, tc.wantTotal)
			}
			if effective != tc.wantEffective {
				t.Errorf("Effective: got %d, want %d", effective, tc.wantEffective)
			}
			if cachePercent != tc.wantCachePercent {
				t.Errorf("CachePercent: got %.2f, want %.2f", cachePercent, tc.wantCachePercent)
			}
		})
	}
}

func TestPercentageCalculations(t *testing.T) {
	tests := []struct {
		name          string
		totalTokens   int64
		contextWindow int64
		wantPercent   float64
		shouldWarn    bool
	}{
		{
			name:          "10% usage - no warning",
			totalTokens:   20000,
			contextWindow: 200000,
			wantPercent:   10.0,
			shouldWarn:    false,
		},
		{
			name:          "50% usage - no warning",
			totalTokens:   100000,
			contextWindow: 200000,
			wantPercent:   50.0,
			shouldWarn:    false,
		},
		{
			name:          "79% usage - no warning",
			totalTokens:   158000,
			contextWindow: 200000,
			wantPercent:   79.0,
			shouldWarn:    false,
		},
		{
			name:          "80% usage - warning threshold",
			totalTokens:   160000,
			contextWindow: 200000,
			wantPercent:   80.0,
			shouldWarn:    true,
		},
		{
			name:          "90% usage - should warn",
			totalTokens:   180000,
			contextWindow: 200000,
			wantPercent:   90.0,
			shouldWarn:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			percent := (float64(tc.totalTokens) / float64(tc.contextWindow)) * 100

			if percent != tc.wantPercent {
				t.Errorf("Percent: got %.2f, want %.2f", percent, tc.wantPercent)
			}

			// Check warning threshold at 80%
			willWarn := percent >= 80.0
			if tc.shouldWarn && !willWarn {
				t.Errorf("Expected warning at %.0f%% but shouldWarn=false", tc.wantPercent)
			}
			if !tc.shouldWarn && willWarn {
				t.Errorf("Did not expect warning at %.0f%% but shouldWarn=true", tc.wantPercent)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
