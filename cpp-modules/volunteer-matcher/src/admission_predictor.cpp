#include "admission_predictor.h"

#include <algorithm>
#include <memory>
#include <sstream>

#include "volunteer_matcher.h"

namespace volunteer_matcher {

class AdmissionPredictor::Impl {
public:
    bool initialized = false;
};

AdmissionPredictor::AdmissionPredictor() : pimpl_(std::make_unique<Impl>()) {}

AdmissionPredictor::~AdmissionPredictor() = default;

bool AdmissionPredictor::Initialize(const std::string& /*config_path*/) {
    pimpl_->initialized = true;
    return true;
}

PredictionResult AdmissionPredictor::PredictAdmissionProbability(
    const Student& student,
    const University& university,
    const Major& major) const {

    PredictionResult result{};

    double baseline_score = static_cast<double>(student.total_score);
    if (!university.historical_scores.empty()) {
        baseline_score -= static_cast<double>(university.historical_scores.back());
    } else {
        baseline_score -= 550.0;
    }

    double probability = 0.5 + baseline_score / 120.0;
    probability += (major.employment_rate - 0.75) * 0.1;
    probability += (100.0 - static_cast<double>(std::max(university.ranking, 1))) / 1000.0;
    probability = std::clamp(probability, 0.02, 0.98);

    result.probability = probability;
    result.confidence = 0.82;
    result.feature_importance = {
        {"score_gap", 0.55},
        {"university_ranking", 0.25},
        {"major_employment", 0.20},
    };

    std::ostringstream explanation;
    explanation << "Predicted by lightweight fallback model with probability "
                << probability;
    result.explanation = explanation.str();

    return result;
}

}  // namespace volunteer_matcher
