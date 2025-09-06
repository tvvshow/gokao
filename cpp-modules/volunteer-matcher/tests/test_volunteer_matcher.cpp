/**
 * @file test_volunteer_matcher.cpp
 * @brief 高考志愿填报系统 - 志愿匹配算法单元测试
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include <gtest/gtest.h>
#include <gmock/gmock.h>
#include <chrono>
#include <thread>

#include "volunteer_matcher.h"

using namespace volunteer_matcher;
using ::testing::_;
using ::testing::Return;
using ::testing::AtLeast;

class VolunteerMatcherTest : public ::testing::Test {
protected:
    void SetUp() override {
        matcher_ = std::make_unique<VolunteerMatcher>();
        
        // 创建测试学生数据
        test_student_.student_id = "test_001";
        test_student_.name = "张三";
        test_student_.total_score = 650;
        test_student_.ranking = 1000;
        test_student_.province = "北京";
        test_student_.subject_combination = "物理+化学+生物";
        
        test_student_.chinese_score = 130;
        test_student_.math_score = 145;
        test_student_.english_score = 140;
        test_student_.physics_score = 95;
        test_student_.chemistry_score = 92;
        test_student_.biology_score = 88;
        
        test_student_.preferred_cities = {"北京", "上海", "深圳"};
        test_student_.preferred_majors = {"计算机科学与技术", "软件工程", "人工智能"};
        test_student_.city_weight = 0.3;
        test_student_.major_weight = 0.4;
        test_student_.school_ranking_weight = 0.3;
        
        // 创建测试大学数据
        test_university_.university_id = "uni_001";
        test_university_.name = "清华大学";
        test_university_.province = "北京";
        test_university_.city = "北京";
        test_university_.level = "985";
        test_university_.ranking = 1;
        test_university_.historical_scores = {645, 648, 650, 652, 655};
        test_university_.total_enrollment = 3000;
        test_university_.employment_rate = 0.98;
        test_university_.graduate_salary = 25000.0;
        
        // 创建测试专业数据
        test_major_.major_id = "major_001";
        test_major_.name = "计算机科学与技术";
        test_major_.category = "工学";
        test_major_.subject_requirements = "物理+化学";
        test_major_.employment_rate = 0.96;
        test_major_.salary_level = 30000.0;
        test_major_.difficulty_level = 0.8;
    }
    
    void TearDown() override {
        matcher_.reset();
    }
    
    std::unique_ptr<VolunteerMatcher> matcher_;
    Student test_student_;
    University test_university_;
    Major test_major_;
};

// 基础功能测试

TEST_F(VolunteerMatcherTest, Construction) {
    EXPECT_NE(matcher_, nullptr);
}

TEST_F(VolunteerMatcherTest, InitializationWithValidConfig) {
    // 创建临时配置文件
    std::string config_content = R"({
        "algorithm": {
            "max_volunteers": 96,
            "risk_threshold": 0.7,
            "match_threshold": 0.6
        },
        "weights": {
            "score_weight": 0.4,
            "location_weight": 0.2,
            "major_weight": 0.3,
            "ranking_weight": 0.1
        }
    })";
    
    std::string config_path = "test_config.json";
    std::ofstream config_file(config_path);
    config_file << config_content;
    config_file.close();
    
    bool result = matcher_->Initialize(config_path);
    EXPECT_TRUE(result);
    
    // 清理
    std::remove(config_path.c_str());
}

TEST_F(VolunteerMatcherTest, InitializationWithInvalidConfig) {
    bool result = matcher_->Initialize("non_existent_config.json");
    EXPECT_FALSE(result);
}

TEST_F(VolunteerMatcherTest, LoadUniversitiesFromValidFile) {
    // 创建临时大学数据文件
    std::string universities_content = 
        "university_id,name,province,city,level,ranking,historical_scores\n"
        "uni_001,清华大学,北京,北京,985,1,\"645,648,650,652,655\"\n"
        "uni_002,北京大学,北京,北京,985,2,\"644,647,649,651,654\"\n";
    
    std::string file_path = "test_universities.csv";
    std::ofstream file(file_path);
    file << universities_content;
    file.close();
    
    int count = matcher_->LoadUniversities(file_path);
    EXPECT_GT(count, 0);
    
    // 清理
    std::remove(file_path.c_str());
}

TEST_F(VolunteerMatcherTest, LoadMajorsFromValidFile) {
    // 创建临时专业数据文件
    std::string majors_content = 
        "major_id,name,category,subject_requirements,employment_rate,salary_level\n"
        "major_001,计算机科学与技术,工学,物理+化学,0.96,30000\n"
        "major_002,软件工程,工学,物理+化学,0.95,28000\n";
    
    std::string file_path = "test_majors.csv";
    std::ofstream file(file_path);
    file << majors_content;
    file.close();
    
    int count = matcher_->LoadMajors(file_path);
    EXPECT_GT(count, 0);
    
    // 清理
    std::remove(file_path.c_str());
}

// 核心算法测试

TEST_F(VolunteerMatcherTest, GenerateVolunteerPlanBasic) {
    // 设置必要的初始化
    matcher_->Initialize("test_config.json");
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(test_student_, 10);
    
    EXPECT_EQ(plan.student_id, test_student_.student_id);
    EXPECT_LE(plan.total_volunteers, 10);
    EXPECT_GE(plan.total_volunteers, 0);
}

TEST_F(VolunteerMatcherTest, GenerateVolunteerPlanWithLargeVolunteerCount) {
    matcher_->Initialize("test_config.json");
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(test_student_, 96);
    
    EXPECT_EQ(plan.student_id, test_student_.student_id);
    EXPECT_LE(plan.total_volunteers, 96);
}

TEST_F(VolunteerMatcherTest, BatchGenerateVolunteerPlans) {
    matcher_->Initialize("test_config.json");
    
    std::vector<Student> students;
    for (int i = 0; i < 5; ++i) {
        Student student = test_student_;
        student.student_id = "test_" + std::to_string(i);
        student.total_score += i * 10;
        students.push_back(student);
    }
    
    auto plans = matcher_->BatchGenerateVolunteerPlans(students, 20);
    
    EXPECT_EQ(plans.size(), students.size());
    for (size_t i = 0; i < plans.size(); ++i) {
        EXPECT_EQ(plans[i].student_id, students[i].student_id);
    }
}

// 性能测试

TEST_F(VolunteerMatcherTest, PerformanceStatsTracking) {
    matcher_->Initialize("test_config.json");
    
    // 重置统计
    matcher_->ResetPerformanceStats();
    
    // 执行一些操作
    matcher_->GenerateVolunteerPlan(test_student_, 10);
    
    auto stats = matcher_->GetPerformanceStats();
    EXPECT_GT(stats.total_requests.load(), 0);
    EXPECT_GT(stats.successful_requests.load(), 0);
    EXPECT_GE(stats.avg_response_time.load(), 0.0);
}

TEST_F(VolunteerMatcherTest, ResponseTimeReasonable) {
    matcher_->Initialize("test_config.json");
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(test_student_, 50);
    
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    
    // 响应时间应小于50ms (性能要求)
    EXPECT_LT(duration.count(), 50);
}

// 并发测试

TEST_F(VolunteerMatcherTest, ConcurrentAccess) {
    matcher_->Initialize("test_config.json");
    
    const int num_threads = 10;
    std::vector<std::thread> threads;
    std::vector<bool> results(num_threads, false);
    
    for (int i = 0; i < num_threads; ++i) {
        threads.emplace_back([this, &results, i]() {
            Student student = test_student_;
            student.student_id = "concurrent_test_" + std::to_string(i);
            
            VolunteerPlan plan = matcher_->GenerateVolunteerPlan(student, 20);
            results[i] = !plan.student_id.empty();
        });
    }
    
    for (auto& thread : threads) {
        thread.join();
    }
    
    // 所有线程都应该成功
    for (bool result : results) {
        EXPECT_TRUE(result);
    }
}

// 边界条件测试

TEST_F(VolunteerMatcherTest, HandleEmptyStudent) {
    matcher_->Initialize("test_config.json");
    
    Student empty_student;
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(empty_student, 10);
    
    // 应该返回空方案或处理错误
    EXPECT_TRUE(plan.student_id.empty() || plan.total_volunteers == 0);
}

TEST_F(VolunteerMatcherTest, HandleZeroVolunteers) {
    matcher_->Initialize("test_config.json");
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(test_student_, 0);
    
    EXPECT_EQ(plan.total_volunteers, 0);
}

TEST_F(VolunteerMatcherTest, HandleNegativeVolunteers) {
    matcher_->Initialize("test_config.json");
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(test_student_, -5);
    
    // 应该处理负数情况，可能返回0或默认值
    EXPECT_GE(plan.total_volunteers, 0);
}

// 数据热更新测试

TEST_F(VolunteerMatcherTest, HotUpdateUniversityData) {
    matcher_->Initialize("test_config.json");
    
    // 创建新的大学数据文件
    std::string new_universities_content = 
        "university_id,name,province,city,level,ranking,historical_scores\n"
        "uni_003,中国科学技术大学,安徽,合肥,985,3,\"640,643,645,647,650\"\n";
    
    std::string file_path = "new_universities.csv";
    std::ofstream file(file_path);
    file << new_universities_content;
    file.close();
    
    bool result = matcher_->HotUpdateData("universities", file_path);
    EXPECT_TRUE(result);
    
    // 清理
    std::remove(file_path.c_str());
}

// 算法正确性测试

TEST_F(VolunteerMatcherTest, MatchScoreCalculation) {
    double score = CalculateMatchScore(test_student_, test_university_, test_major_);
    
    EXPECT_GE(score, 0.0);
    EXPECT_LE(score, 100.0);
}

TEST_F(VolunteerMatcherTest, SubjectCombinationParsing) {
    std::string combination = "物理+化学+生物";
    auto subjects = ParseSubjectCombination(combination);
    
    EXPECT_EQ(subjects.size(), 3);
    EXPECT_EQ(subjects[0], "物理");
    EXPECT_EQ(subjects[1], "化学");
    EXPECT_EQ(subjects[2], "生物");
}

TEST_F(VolunteerMatcherTest, SubjectRequirementsValidation) {
    std::vector<std::string> student_subjects = {"物理", "化学", "生物"};
    std::string requirements = "物理+化学";
    
    bool valid = ValidateSubjectRequirements(student_subjects, requirements);
    EXPECT_TRUE(valid);
    
    // 测试不满足要求的情况
    requirements = "物理+历史";
    valid = ValidateSubjectRequirements(student_subjects, requirements);
    EXPECT_FALSE(valid);
}

TEST_F(VolunteerMatcherTest, TrendCoefficientCalculation) {
    std::vector<int> increasing_trend = {600, 610, 620, 630, 640};
    double coefficient = CalculateTrendCoefficient(increasing_trend);
    EXPECT_GT(coefficient, 1.0); // 上升趋势
    
    std::vector<int> decreasing_trend = {640, 630, 620, 610, 600};
    coefficient = CalculateTrendCoefficient(decreasing_trend);
    EXPECT_LT(coefficient, 1.0); // 下降趋势
}

// 内存管理测试

TEST_F(VolunteerMatcherTest, MemoryLeakTest) {
    // 创建和销毁多个匹配器实例
    for (int i = 0; i < 100; ++i) {
        auto temp_matcher = std::make_unique<VolunteerMatcher>();
        temp_matcher->Initialize("test_config.json");
        temp_matcher->GenerateVolunteerPlan(test_student_, 10);
    }
    
    // 如果有内存泄漏，此测试在valgrind下会报告
    SUCCEED();
}

// 错误处理测试

TEST_F(VolunteerMatcherTest, ErrorHandlingInvalidData) {
    matcher_->Initialize("test_config.json");
    
    // 测试无效学生数据
    Student invalid_student;
    invalid_student.total_score = -100; // 无效分数
    
    VolunteerPlan plan = matcher_->GenerateVolunteerPlan(invalid_student, 10);
    
    // 应该优雅地处理错误
    EXPECT_TRUE(plan.student_id.empty() || plan.total_volunteers == 0);
}

// 引擎状态测试

TEST_F(VolunteerMatcherTest, EngineStatusReporting) {
    matcher_->Initialize("test_config.json");
    
    std::string status = matcher_->GetEngineStatus();
    
    EXPECT_FALSE(status.empty());
    EXPECT_NE(status.find("initialized"), std::string::npos);
}

// 日志级别测试

TEST_F(VolunteerMatcherTest, LogLevelSetting) {
    matcher_->Initialize("test_config.json");
    
    // 测试各种日志级别
    matcher_->SetLogLevel("DEBUG");
    matcher_->SetLogLevel("INFO");
    matcher_->SetLogLevel("WARNING");
    matcher_->SetLogLevel("ERROR");
    
    // 如果没有异常抛出，则认为成功
    SUCCEED();
}

} // 测试命名空间

// 主函数
int main(int argc, char **argv) {
    ::testing::InitGoogleTest(&argc, argv);
    
    // 设置测试环境
    std::cout << "Running Volunteer Matcher Unit Tests..." << std::endl;
    
    int result = RUN_ALL_TESTS();
    
    std::cout << "Unit tests completed." << std::endl;
    
    return result;
}