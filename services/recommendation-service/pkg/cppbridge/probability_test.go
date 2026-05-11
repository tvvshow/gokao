package cppbridge

import (
	"math"
	"testing"
)

const cdfTolerance = 1e-4

func TestNormalCDF(t *testing.T) {
	cases := []struct {
		name string
		z    float64
		want float64
	}{
		{"zero", 0.0, 0.5},
		{"one_sigma_above", 1.0, 0.8413447},
		{"one_sigma_below", -1.0, 0.1586553},
		{"two_sigma_above", 2.0, 0.9772499},
		{"two_sigma_below", -2.0, 0.0227501},
		{"three_sigma_above", 3.0, 0.9986501},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalCDF(tc.z)
			if math.Abs(got-tc.want) > cdfTolerance {
				t.Fatalf("normalCDF(%v) = %v, want %v (±%v)", tc.z, got, tc.want, cdfTolerance)
			}
		})
	}
}

func TestEstimateStdDev(t *testing.T) {
	cases := []struct {
		name   string
		record AdmissionRecord
		want   float64
	}{
		{
			name:   "normal_spread",
			record: AdmissionRecord{MinScore: 540, MaxScore: 600},
			want:   15.0,
		},
		{
			name:   "narrow_spread_hits_floor",
			record: AdmissionRecord{MinScore: 579, MaxScore: 580},
			want:   sigmaFloor,
		},
		{
			name:   "equal_min_max_uses_fallback",
			record: AdmissionRecord{MinScore: 560, MaxScore: 560},
			want:   sigmaFallback,
		},
		{
			name:   "inverted_dirty_data_uses_fallback",
			record: AdmissionRecord{MinScore: 600, MaxScore: 540},
			want:   sigmaFallback,
		},
		{
			name:   "wide_spread",
			record: AdmissionRecord{MinScore: 500, MaxScore: 660},
			want:   40.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := estimateStdDev(tc.record)
			if got != tc.want {
				t.Fatalf("estimateStdDev() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAdmissionProbability(t *testing.T) {
	baseRecord := AdmissionRecord{
		MinScore:     540,
		AvgScore:     560,
		MaxScore:     600,
		StudentCount: 50,
	}

	cases := []struct {
		name    string
		score   int
		record  AdmissionRecord
		wantLow float64
		wantHi  float64
	}{
		{
			name:    "score_equals_average",
			score:   560,
			record:  baseRecord,
			wantLow: 0.49,
			wantHi:  0.51,
		},
		{
			name:    "score_one_sigma_above",
			score:   575,
			record:  baseRecord,
			wantLow: 0.83,
			wantHi:  0.85,
		},
		{
			name:    "score_one_sigma_below",
			score:   545,
			record:  baseRecord,
			wantLow: 0.15,
			wantHi:  0.17,
		},
		{
			name:    "extremely_high_score_clamps_to_0_99",
			score:   10000,
			record:  baseRecord,
			wantLow: 0.985,
			wantHi:  0.99,
		},
		{
			name:    "extremely_low_score_clamps_to_0_01",
			score:   0,
			record:  baseRecord,
			wantLow: 0.01,
			wantHi:  0.015,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := admissionProbability(tc.score, tc.record)
			if got < tc.wantLow || got > tc.wantHi {
				t.Fatalf("admissionProbability(%d, %+v) = %v, want [%v, %v]",
					tc.score, tc.record, got, tc.wantLow, tc.wantHi)
			}
		})
	}
}

func TestAdmissionProbability_StudentCountAdjustment(t *testing.T) {
	base := AdmissionRecord{MinScore: 540, AvgScore: 560, MaxScore: 600}

	mediumSample := base
	mediumSample.StudentCount = 50
	largeSample := base
	largeSample.StudentCount = 200
	smallSample := base
	smallSample.StudentCount = 10

	pMedium := admissionProbability(560, mediumSample)
	pLarge := admissionProbability(560, largeSample)
	pSmall := admissionProbability(560, smallSample)

	if math.Abs((pLarge-pMedium)-0.02) > cdfTolerance {
		t.Fatalf("large sample bonus expected +0.02, got %v", pLarge-pMedium)
	}
	if math.Abs((pMedium-pSmall)-0.02) > cdfTolerance {
		t.Fatalf("small sample penalty expected -0.02, got %v", pMedium-pSmall)
	}
}

// TestAdmissionProbability_BridgeMethodParity 双 bridge 的 method 必须返回同一份算法结果.
func TestAdmissionProbability_BridgeMethodParity(t *testing.T) {
	record := AdmissionRecord{MinScore: 540, AvgScore: 560, MaxScore: 600, StudentCount: 80}
	score := 565

	simple := &SimpleRuleRecommendationBridge{}
	enhanced := &EnhancedRuleRecommendationBridge{}

	if simple.calculateAdmissionProbability(score, record) != enhanced.calculateAdmissionProbability(score, record) {
		t.Fatal("simple and enhanced bridge must share the admission probability model")
	}
}
