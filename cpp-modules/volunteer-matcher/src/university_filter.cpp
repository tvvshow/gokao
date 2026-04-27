#include "university_filter.h"

#include <algorithm>
#include <cmath>
#include <memory>
#include <unordered_set>

#include "volunteer_matcher.h"

namespace volunteer_matcher {

class UniversityFilter::Impl {
public:
    std::vector<University> universities;
    std::vector<Major> majors;
};

UniversityFilter::UniversityFilter() : pimpl_(std::make_unique<Impl>()) {}

UniversityFilter::~UniversityFilter() = default;

bool UniversityFilter::Initialize(const std::string& /*config_path*/) {
    return true;
}

int UniversityFilter::SetUniversities(const std::vector<University>& universities) {
    pimpl_->universities = universities;
    return static_cast<int>(pimpl_->universities.size());
}

int UniversityFilter::SetMajors(const std::vector<Major>& majors) {
    pimpl_->majors = majors;
    return static_cast<int>(pimpl_->majors.size());
}

FilterResult UniversityFilter::IntelligentFilter(const Student& student, int max_candidates) const {
    struct Candidate {
        std::string id;
        double score;
    };

    std::vector<Candidate> candidates;
    candidates.reserve(pimpl_->universities.size());

    for (const auto& university : pimpl_->universities) {
        int reference_score = university.historical_scores.empty() ? 550 : university.historical_scores.back();
        double score_gap = std::abs(static_cast<double>(student.total_score - reference_score));
        double ranking_bonus = (1000.0 - static_cast<double>(std::min(std::max(university.ranking, 1), 1000))) / 1000.0;
        double match_score = std::max(0.0, 100.0 - score_gap) + ranking_bonus * 10.0;

        candidates.push_back({university.university_id, match_score});
    }

    std::sort(candidates.begin(), candidates.end(), [](const Candidate& a, const Candidate& b) {
        return a.score > b.score;
    });

    if (max_candidates > 0 && static_cast<int>(candidates.size()) > max_candidates) {
        candidates.resize(max_candidates);
    }

    FilterResult result;
    result.total_candidates = static_cast<int>(pimpl_->universities.size());
    result.filtered_count = static_cast<int>(candidates.size());
    result.filter_ratio = result.total_candidates == 0
        ? 0.0
        : static_cast<double>(result.filtered_count) / static_cast<double>(result.total_candidates);
    result.filter_time = std::chrono::system_clock::now();

    for (const auto& candidate : candidates) {
        result.university_ids.push_back(candidate.id);
        result.match_scores[candidate.id] = candidate.score;
        result.filter_reasons[candidate.id] = {"score_match", "ranking_balance"};
    }

    return result;
}

}  // namespace volunteer_matcher
