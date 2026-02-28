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
			name: "green state with 50% cache",
			info: &ModelContextInfo{
				TotalTokens:     44000,  // 22% of 200K
				EffectiveTokens: 22000,
				CacheReadTokens: 22000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            0.24,
			},
			wantContains:     []string{StoplightGreen, "22%", "44K/22K", "$0.24", "50% cached"},
			dontWantContains: []string{"Near limit", StoplightYellow, StoplightRed},
		},
		{
			name: "yellow state (75%) with cache",
			info: &ModelContextInfo{
				TotalTokens:     150000, // 75% of 200K
				EffectiveTokens: 45000,
				CacheReadTokens: 105000,
				CachePercent:    70.0,
				ModelContext:    200000,
				Cost:            0.80,
			},
			wantContains:     []string{StoplightYellow, "75%", "150K/45K", "$0.80", "70% cached"},
			dontWantContains: []string{StoplightGreen, StoplightRed, "Near limit"},
		},
		{
			name: "red state (88%) critical with warning",
			info: &ModelContextInfo{
				TotalTokens:     176000, // 88% of 200K
				EffectiveTokens: 25000,
				CacheReadTokens: 151000,
				CachePercent:    86.0,
				ModelContext:    200000,
				Cost:            1.20,
			},
			wantContains:     []string{StoplightRed, "88%", "176K/25K", "$1.20", "86% cached", "Near limit", "Ctrl+N"},
			dontWantContains: []string{StoplightGreen, StoplightYellow},
		},
		{
			name: "green state with no cache - no cache line",
			info: &ModelContextInfo{
				TotalTokens:     50000, // 25% of 200K
				EffectiveTokens: 50000,
				CacheReadTokens: 0,
				CachePercent:    0,
				ModelContext:    200000,
				Cost:            0.50,
			},
			wantContains:     []string{StoplightGreen, "25%", "50K/50K", "$0.50"},
			dontWantContains: []string{"cached", StoplightYellow, StoplightRed, "Near limit"},
		},
		{
			name: "exact boundary - green to yellow at 70%",
			info: &ModelContextInfo{
				TotalTokens:     140000, // exactly 70% of 200K
				EffectiveTokens: 70000,
				CacheReadTokens: 70000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            0.70,
			},
			wantContains:     []string{StoplightGreen, "70%"}, // 70% is still green
			dontWantContains: []string{StoplightYellow, StoplightRed},
		},
		{
			name: "just into yellow at 71%",
			info: &ModelContextInfo{
				TotalTokens:     142000, // 71% of 200K
				EffectiveTokens: 71000,
				CacheReadTokens: 71000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            0.71,
			},
			wantContains:     []string{StoplightYellow, "71%"},
			dontWantContains: []string{StoplightGreen, StoplightRed, "Near limit"},
		},
		{
			name: "exact red threshold at 85%",
			info: &ModelContextInfo{
				TotalTokens:     170000, // exactly 85% of 200K
				EffectiveTokens: 85000,
				CacheReadTokens: 85000,
				CachePercent:    50.0,
				ModelContext:    200000,
				Cost:            0.85,
			},
			wantContains:     []string{StoplightRed, "85%", "Near limit", "Ctrl+N"},
			dontWantContains: []string{StoplightGreen, StoplightYellow},
		},
		{
			name: "large token values with millions",
			info: &ModelContextInfo{
				TotalTokens:     1500000, // 75% of 2M
				EffectiveTokens: 1000000,
				CacheReadTokens: 500000,
				CachePercent:    33.3,
				ModelContext:    2000000,
				Cost:            15.00,
			},
			wantContains:     []string{StoplightYellow, "1.5M/1M", "$15.00", "33% cached"},
			dontWantContains: []string{StoplightGreen, StoplightRed, "Near limit"},
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
			wantContains:     []string{StoplightGreen, "500/400", "$0.01", "20% cached"},
			dontWantContains: []string{StoplightYellow, StoplightRed, "Near limit"},
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

func TestGetStoplightIndicator(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		want       string
	}{
		{"0% - green", 0.0, StoplightGreen},
		{"50% - green", 50.0, StoplightGreen},
		{"70% - green (boundary)", 70.0, StoplightGreen},
		{"71% - yellow", 71.0, StoplightYellow},
		{"80% - yellow", 80.0, StoplightYellow},
		{"84% - yellow (boundary)", 84.0, StoplightYellow},
		{"85% - red", 85.0, StoplightRed},
		{"90% - red", 90.0, StoplightRed},
		{"100% - red", 100.0, StoplightRed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getStoplightIndicator(tc.percentage)
			if result != tc.want {
				t.Errorf("getStoplightIndicator(%.1f) = %s, want %s", tc.percentage, result, tc.want)
			}
		})
	}
}

func TestIsCriticalState(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		want       bool
	}{
		{"50% - not critical", 50.0, false},
		{"70% - not critical", 70.0, false},
		{"84% - not critical", 84.0, false},
		{"85% - critical", 85.0, true},
		{"90% - critical", 90.0, true},
		{"100% - critical", 100.0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isCriticalState(tc.percentage)
			if result != tc.want {
				t.Errorf("isCriticalState(%.1f) = %v, want %v", tc.percentage, result, tc.want)
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

func TestStoplightThresholds(t *testing.T) {
	// Verify the threshold constants are what we expect
	if StoplightGreenThreshold != 70.0 {
		t.Errorf("StoplightGreenThreshold = %.1f, want 70.0", StoplightGreenThreshold)
	}
	if StoplightYellowThreshold != 84.0 {
		t.Errorf("StoplightYellowThreshold = %.1f, want 84.0", StoplightYellowThreshold)
	}
	if StoplightRedThreshold != 85.0 {
		t.Errorf("StoplightRedThreshold = %.1f, want 85.0", StoplightRedThreshold)
	}
}

func TestBulletSeparator(t *testing.T) {
	// Verify that the bullet separator is used in the output
	tt := styles.DefaultStyles()
	info := &ModelContextInfo{
		TotalTokens:     50000,
		EffectiveTokens: 25000,
		CacheReadTokens: 25000,
		CachePercent:    50.0,
		ModelContext:    200000,
		Cost:            0.50,
	}

	result := formatTokensAndCost(&tt, info)

	// Should contain bullet separator (•)
	if !contains(result, "•") {
		t.Errorf("Expected bullet separator (•) in output, got: %s", result)
	}

	// Should NOT contain old-style "total / effective" format
	if contains(result, "total /") {
		t.Errorf("Expected NOT to contain old 'total /' format, got: %s", result)
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
