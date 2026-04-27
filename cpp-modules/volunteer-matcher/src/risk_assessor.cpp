#include "risk_assessor.h"

#include <algorithm>
#include <fstream>
#include <memory>

#include "volunteer_matcher.h"

namespace volunteer_matcher {

class RiskAssessor::Impl {
public:
    bool initialized = false;
    bool historical_data_loaded = false;
};

RiskAssessor::RiskAssessor() : pimpl_(std::make_unique<Impl>()) {}

RiskAssessor::~RiskAssessor() = default;

bool RiskAssessor::Initialize(const std::string& /*config_path*/) {
    pimpl_->initialized = true;
    return true;
}

bool RiskAssessor::SetHistoricalData(const std::string& historical_data_path) {
    std::ifstream file(historical_data_path);
    pimpl_->historical_data_loaded = file.good();
    return pimpl_->historical_data_loaded;
}

RiskAssessment RiskAssessor::AssessVolunteerRisk(
    const Student& student,
    const University& university,
    const Major& major) const {

    RiskAssessment assessment{};
    assessment.target_id = university.university_id + ":" + major.major_id;
    assessment.assessment_time = std::chrono::system_clock::now();
    assessment.confidence = pimpl_->historical_data_loaded ? 0.85 : 0.65;

    int reference_score = university.historical_scores.empty() ? 550 : university.historical_scores.back();
    double score_gap = static_cast<double>(reference_score - student.total_score);
    double risk = 50.0 + score_gap * 1.2;
    risk -= major.employment_rate * 10.0;
    risk = std::clamp(risk, 5.0, 95.0);

    assessment.overall_risk_score = risk;
    if (risk >= 80.0) {
        assessment.overall_risk_level = RiskLevel::VERY_HIGH;
    } else if (risk >= 65.0) {
        assessment.overall_risk_level = RiskLevel::HIGH;
    } else if (risk >= 45.0) {
        assessment.overall_risk_level = RiskLevel::MEDIUM;
    } else if (risk >= 25.0) {
        assessment.overall_risk_level = RiskLevel::LOW;
    } else {
        assessment.overall_risk_level = RiskLevel::VERY_LOW;
    }

    assessment.assessment_summary = "Lightweight fallback risk assessment";
    assessment.risk_type_scores[RiskType::ADMISSION_RISK] = risk;
    assessment.risk_type_scores[RiskType::EMPLOYMENT_RISK] = std::max(0.0, 100.0 - major.employment_rate * 100.0);

    RiskFactor factor;
    factor.type = RiskType::ADMISSION_RISK;
    factor.level = assessment.overall_risk_level;
    factor.probability = std::clamp(risk / 100.0, 0.05, 0.95);
    factor.impact = 0.8;
    factor.severity = risk / 100.0;
    factor.description = "Admission competitiveness assessment";
    factor.causes = {"score_gap", "historical_cutoff"};
    factor.consequences = {"lower_admission_probability"};
    assessment.risk_factors.push_back(factor);

    return assessment;
}

}  // namespace volunteer_matcher
