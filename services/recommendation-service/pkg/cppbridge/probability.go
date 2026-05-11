package cppbridge

import "math"

// Admission probability model based on a normal-distribution CDF.
//
// Replaces the 7-step hardcoded switch (D.1 in TECHNICAL_DEBT_AUDIT.md).
// Treats the school+major admission scores as N(μ, σ²) and returns Φ((s-μ)/σ).
//
// μ comes from AdmissionRecord.AvgScore. σ has to be estimated because the
// upstream record only carries MinScore/MaxScore — we use the empirical
// 4-sigma rule (σ ≈ (max - min) / 4) with a floor of sigmaFloor so a year
// of narrow data does not collapse σ → 0 and make the CDF a step function.

const (
	// sigmaFloor 防止极小样本年份的 σ 估计塌缩为 0 导致 CDF 退化为阶跃。
	sigmaFloor = 5.0

	// sigmaFallback 当 MaxScore/MinScore 缺失或并列时的回退值，
	// 经验上多数省份重点专业一年内分数带宽约 30 分。
	sigmaFallback = 15.0
)

// normalCDF 标准正态分布累积分布函数 Φ(z) = (1 + erf(z/√2)) / 2.
func normalCDF(z float64) float64 {
	return 0.5 * (1.0 + math.Erf(z/math.Sqrt2))
}

// estimateStdDev 从录取记录估计分数 σ.
// (max - min) ≈ 4σ 经验法则；min == max 视为数据缺失走 fallback.
func estimateStdDev(record AdmissionRecord) float64 {
	if record.MaxScore <= record.MinScore {
		return sigmaFallback
	}
	sigma := float64(record.MaxScore-record.MinScore) / 4.0
	if sigma < sigmaFloor {
		return sigmaFloor
	}
	return sigma
}

// admissionProbability 返回 (studentScore, record) 在正态分布下的录取概率，
// 已 clamp 至 [0.01, 0.99] 避免 0 / 1 极端值掩盖剩余不确定性。
//
// StudentCount 视为可信度调整：>100 上调 0.02，<30 下调 0.02 — 反映
// 大样本年份的方差估计更可信、小样本年份保留更多不确定性。
func admissionProbability(studentScore int, record AdmissionRecord) float64 {
	mu := float64(record.AvgScore)
	sigma := estimateStdDev(record)
	z := (float64(studentScore) - mu) / sigma
	p := normalCDF(z)

	switch {
	case record.StudentCount > 100:
		p += 0.02
	case record.StudentCount < 30:
		p -= 0.02
	}

	return math.Max(0.01, math.Min(0.99, p))
}
