/**
 * @file volunteer_matcher_example.cpp
 * @brief 高考志愿填报系统 - 志愿匹配引擎示例程序
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "volunteer_matcher.h"
#include <iostream>
#include <iomanip>
#include <chrono>

using namespace volunteer_matcher;

/**
 * @brief 创建示例学生数据
 */
Student CreateSampleStudent() {
    Student student;
    student.student_id = "20240001";
    student.name = "张同学";
    student.total_score = 620;
    student.ranking = 5000;
    student.province = "北京";
    student.subject_combination = "物理+化学+生物";
    
    // 单科成绩
    student.chinese_score = 125;
    student.math_score = 140;
    student.english_score = 130;
    student.physics_score = 88;
    student.chemistry_score = 85;
    student.biology_score = 82;
    student.politics_score = 0;
    student.history_score = 0;
    student.geography_score = 0;
    
    // 偏好设置
    student.preferred_cities = {"北京", "上海", "深圳", "广州"};
    student.preferred_majors = {"计算机科学与技术", "软件工程", "人工智能", "数据科学"};
    student.avoided_majors = {"哲学", "历史学"};
    student.city_weight = 0.3;
    student.major_weight = 0.5;
    student.school_ranking_weight = 0.2;
    
    // 特殊情况
    student.is_minority = false;
    student.has_sports_specialty = false;
    student.has_art_specialty = false;
    
    return student;
}

/**
 * @brief 创建示例大学数据
 */
std::vector<University> CreateSampleUniversities() {
    std::vector<University> universities;
    
    // 清华大学
    University tsinghua;
    tsinghua.university_id = "10003";
    tsinghua.name = "清华大学";
    tsinghua.province = "北京";
    tsinghua.city = "北京";
    tsinghua.level = "985";
    tsinghua.ranking = 1;
    tsinghua.historical_scores = {680, 685, 682, 690, 688};
    tsinghua.historical_rankings = {300, 250, 280, 200, 220};
    tsinghua.total_enrollment = 3000;
    tsinghua.strong_majors = {"计算机科学与技术", "电子信息工程", "自动化", "机械工程"};
    tsinghua.employment_rate = 0.98;
    tsinghua.graduate_salary = 15000;
    universities.push_back(tsinghua);
    
    // 北京大学
    University pku;
    pku.university_id = "10001";
    pku.name = "北京大学";
    pku.province = "北京";
    pku.city = "北京";
    pku.level = "985";
    pku.ranking = 2;
    pku.historical_scores = {675, 680, 678, 685, 683};
    pku.historical_rankings = {350, 300, 320, 250, 270};
    pku.total_enrollment = 2800;
    pku.strong_majors = {"数学与应用数学", "物理学", "化学", "生物科学"};
    pku.employment_rate = 0.97;
    pku.graduate_salary = 14500;
    universities.push_back(pku);
    
    // 上海交通大学
    University sjtu;
    sjtu.university_id = "10248";
    sjtu.name = "上海交通大学";
    sjtu.province = "上海";
    sjtu.city = "上海";
    sjtu.level = "985";
    sjtu.ranking = 3;
    sjtu.historical_scores = {665, 670, 668, 675, 672};
    sjtu.historical_rankings = {500, 450, 480, 380, 400};
    sjtu.total_enrollment = 3200;
    sjtu.strong_majors = {"机械工程", "电子信息工程", "船舶与海洋工程", "材料科学与工程"};
    sjtu.employment_rate = 0.96;
    sjtu.graduate_salary = 13800;
    universities.push_back(sjtu);
    
    // 华中科技大学
    University hust;
    hust.university_id = "10487";
    hust.name = "华中科技大学";
    hust.province = "湖北";
    hust.city = "武汉";
    hust.level = "985";
    hust.ranking = 8;
    hust.historical_scores = {615, 620, 618, 625, 622};
    hust.historical_rankings = {5500, 5000, 5200, 4500, 4800};
    hust.total_enrollment = 4000;
    hust.strong_majors = {"机械工程", "电气工程", "光电信息科学与工程", "计算机科学与技术"};
    hust.employment_rate = 0.94;
    hust.graduate_salary = 12000;
    universities.push_back(hust);
    
    // 北京理工大学
    University bit;
    bit.university_id = "10007";
    bit.name = "北京理工大学";
    bit.province = "北京";
    bit.city = "北京";
    bit.level = "985";
    bit.ranking = 15;
    bit.historical_scores = {600, 605, 602, 610, 607};
    bit.historical_rankings = {8000, 7500, 7800, 7000, 7200};
    bit.total_enrollment = 3500;
    bit.strong_majors = {"车辆工程", "兵器科学与技术", "光电信息科学与工程", "计算机科学与技术"};
    bit.employment_rate = 0.93;
    bit.graduate_salary = 11500;
    universities.push_back(bit);
    
    return universities;
}

/**
 * @brief 创建示例专业数据
 */
std::vector<Major> CreateSampleMajors() {
    std::vector<Major> majors;
    
    // 计算机科学与技术
    Major cs;
    cs.major_id = "080901";
    cs.name = "计算机科学与技术";
    cs.category = "工学";
    cs.subject_requirements = "物理";
    cs.employment_rate = 0.95;
    cs.salary_level = 12000;
    cs.career_directions = {"软件工程师", "算法工程师", "系统架构师", "产品经理"};
    cs.difficulty_level = 0.7;
    cs.requires_postgraduate = false;
    majors.push_back(cs);
    
    // 软件工程
    Major se;
    se.major_id = "080902";
    se.name = "软件工程";
    se.category = "工学";
    se.subject_requirements = "物理";
    se.employment_rate = 0.96;
    se.salary_level = 11800;
    se.career_directions = {"软件开发工程师", "测试工程师", "运维工程师", "项目经理"};
    se.difficulty_level = 0.6;
    se.requires_postgraduate = false;
    majors.push_back(se);
    
    // 电子信息工程
    Major ee;
    ee.major_id = "080701";
    ee.name = "电子信息工程";
    ee.category = "工学";
    ee.subject_requirements = "物理";
    ee.employment_rate = 0.92;
    ee.salary_level = 10500;
    ee.career_directions = {"硬件工程师", "嵌入式开发", "通信工程师", "集成电路设计"};
    ee.difficulty_level = 0.8;
    ee.requires_postgraduate = true;
    majors.push_back(ee);
    
    return majors;
}

/**
 * @brief 打印志愿方案
 */
void PrintVolunteerPlan(const VolunteerPlan& plan) {
    std::cout << "\n==================== 志愿填报方案 ====================" << std::endl;
    std::cout << "学生ID: " << plan.student_id << std::endl;
    std::cout << "方案质量: " << plan.plan_quality << std::endl;
    std::cout << "总体风险评分: " << std::fixed << std::setprecision(2) << plan.overall_risk_score << std::endl;
    std::cout << "志愿统计: 总计 " << plan.total_volunteers 
              << " (冲 " << plan.rush_count 
              << ", 稳 " << plan.stable_count 
              << ", 保 " << plan.safe_count << ")" << std::endl;
    
    std::cout << "\n志愿推荐列表:" << std::endl;
    std::cout << std::setw(4) << "序号" 
              << std::setw(20) << "大学名称"
              << std::setw(20) << "专业名称"
              << std::setw(8) << "录取概率"
              << std::setw(6) << "风险"
              << std::setw(8) << "匹配度"
              << std::setw(8) << "分差"
              << std::endl;
    std::cout << std::string(80, '-') << std::endl;
    
    for (size_t i = 0; i < plan.recommendations.size(); ++i) {
        const auto& rec = plan.recommendations[i];
        std::cout << std::setw(4) << (i + 1)
                  << std::setw(20) << rec.university_name.substr(0, 18)
                  << std::setw(20) << rec.major_name.substr(0, 18)
                  << std::setw(7) << std::fixed << std::setprecision(1) << (rec.admission_probability * 100) << "%"
                  << std::setw(6) << rec.risk_level
                  << std::setw(7) << std::fixed << std::setprecision(1) << rec.match_score
                  << std::setw(8) << rec.score_gap
                  << std::endl;
    }
    
    if (!plan.optimization_suggestions.empty()) {
        std::cout << "\n优化建议:" << std::endl;
        for (size_t i = 0; i < plan.optimization_suggestions.size(); ++i) {
            std::cout << "  " << (i + 1) << ". " << plan.optimization_suggestions[i] << std::endl;
        }
    }
    
    std::cout << "======================================================" << std::endl;
}

/**
 * @brief 打印性能统计
 */
void PrintPerformanceStats(const PerformanceStats& stats) {
    std::cout << "\n==================== 性能统计 ====================" << std::endl;
    std::cout << "总请求数: " << stats.total_requests.load() << std::endl;
    std::cout << "成功请求数: " << stats.successful_requests.load() << std::endl;
    std::cout << "平均响应时间: " << std::fixed << std::setprecision(2) 
              << stats.avg_response_time.load() << " ms" << std::endl;
    std::cout << "最大响应时间: " << std::fixed << std::setprecision(2) 
              << stats.max_response_time.load() << " ms" << std::endl;
    std::cout << "内存使用: " << stats.memory_usage.load() << " bytes" << std::endl;
    std::cout << "===================================================" << std::endl;
}

/**
 * @brief 主函数
 */
int main() {
    std::cout << "高考志愿填报系统 - 志愿匹配引擎示例程序" << std::endl;
    std::cout << "版本: 1.0.0" << std::endl;
    std::cout << "开发团队: 高考志愿填报系统开发团队" << std::endl;
    std::cout << "======================================================" << std::endl;
    
    try {
        // 创建志愿匹配器
        std::cout << "\n1. 创建志愿匹配器..." << std::endl;
        VolunteerMatcher matcher;
        
        // 初始化（使用默认配置）
        std::cout << "2. 初始化志愿匹配器..." << std::endl;
        bool init_success = matcher.Initialize("config.json");  // 这里假设有配置文件
        if (!init_success) {
            std::cout << "   警告: 无法加载配置文件，使用默认配置" << std::endl;
        }
        
        // 创建示例数据
        std::cout << "3. 创建示例数据..." << std::endl;
        auto universities = CreateSampleUniversities();
        auto majors = CreateSampleMajors();
        auto student = CreateSampleStudent();
        
        std::cout << "   创建了 " << universities.size() << " 所大学" << std::endl;
        std::cout << "   创建了 " << majors.size() << " 个专业" << std::endl;
        std::cout << "   学生信息: " << student.name << " (总分: " << student.total_score << ")" << std::endl;
        
        // 加载数据（在实际使用中，这些数据应该从文件加载）
        std::cout << "4. 加载数据到匹配器..." << std::endl;
        // 注意：这里是示例，实际的LoadUniversities和LoadMajors方法需要文件路径
        // int uni_count = matcher.LoadUniversities("universities.csv");
        // int major_count = matcher.LoadMajors("majors.csv");
        std::cout << "   数据加载完成" << std::endl;
        
        // 生成志愿方案
        std::cout << "5. 生成志愿填报方案..." << std::endl;
        auto start_time = std::chrono::high_resolution_clock::now();
        
        VolunteerPlan plan = matcher.GenerateVolunteerPlan(student, 20);  // 生成20个志愿
        
        auto end_time = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
        
        std::cout << "   方案生成完成，用时: " << duration.count() << " ms" << std::endl;
        
        // 打印结果
        PrintVolunteerPlan(plan);
        
        // 测试方案优化
        std::cout << "\n6. 测试方案优化..." << std::endl;
        auto optimized_plan = matcher.OptimizeVolunteerPlan(plan, "safety");
        std::cout << "   安全性优化完成" << std::endl;
        std::cout << "   优化前风险评分: " << std::fixed << std::setprecision(2) << plan.overall_risk_score << std::endl;
        std::cout << "   优化后风险评分: " << std::fixed << std::setprecision(2) << optimized_plan.overall_risk_score << std::endl;
        
        // 批量测试
        std::cout << "\n7. 批量处理测试..." << std::endl;
        std::vector<Student> students;
        for (int i = 0; i < 5; ++i) {
            Student s = student;
            s.student_id = "2024000" + std::to_string(i + 1);
            s.total_score = 600 + i * 10;  // 分数递增
            s.ranking = 8000 - i * 1000;   // 排名递减
            students.push_back(s);
        }
        
        start_time = std::chrono::high_resolution_clock::now();
        auto batch_plans = matcher.BatchGenerateVolunteerPlans(students, 10);
        end_time = std::chrono::high_resolution_clock::now();
        duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
        
        std::cout << "   批量处理 " << students.size() << " 个学生，用时: " << duration.count() << " ms" << std::endl;
        std::cout << "   平均每个学生: " << std::fixed << std::setprecision(1) 
                  << (double)duration.count() / students.size() << " ms" << std::endl;
        
        // 获取性能统计
        std::cout << "\n8. 性能统计..." << std::endl;
        auto stats = matcher.GetPerformanceStats();
        PrintPerformanceStats(stats);
        
        // 获取引擎状态
        std::cout << "\n9. 引擎状态..." << std::endl;
        std::string engine_status = matcher.GetEngineStatus();
        std::cout << "   引擎状态: " << engine_status << std::endl;
        
        std::cout << "\n======================================================" << std::endl;
        std::cout << "示例程序执行完成！" << std::endl;
        
    } catch (const std::exception& e) {
        std::cerr << "错误: " << e.what() << std::endl;
        return 1;
    } catch (...) {
        std::cerr << "未知错误" << std::endl;
        return 1;
    }
    
    return 0;
}