#!/usr/bin/env python3
"""
生成完整的高考志愿填报系统数据
- 专业数据（500+ 专业）
- 录取分数线数据（2021-2024年）
"""

import json
import random
import uuid
from datetime import datetime

# 专业分类定义
MAJOR_CATEGORIES = {
    "工学": {
        "disciplines": ["机械工程", "电子信息", "计算机", "土木工程", "材料科学", "化学工程", "环境工程", "能源动力", "自动化", "航空航天"],
        "degree": "工学学士",
        "duration": 4,
        "base_salary": 8000,
        "employment_rate": 0.92
    },
    "理学": {
        "disciplines": ["数学", "物理学", "化学", "生物学", "地理学", "天文学", "统计学", "心理学"],
        "degree": "理学学士",
        "duration": 4,
        "base_salary": 7000,
        "employment_rate": 0.88
    },
    "医学": {
        "disciplines": ["临床医学", "口腔医学", "中医学", "药学", "护理学", "公共卫生", "医学影像", "麻醉学"],
        "degree": "医学学士",
        "duration": 5,
        "base_salary": 9000,
        "employment_rate": 0.95
    },
    "经济学": {
        "disciplines": ["经济学", "金融学", "国际经济与贸易", "财政学", "投资学", "保险学"],
        "degree": "经济学学士",
        "duration": 4,
        "base_salary": 8500,
        "employment_rate": 0.90
    },
    "管理学": {
        "disciplines": ["工商管理", "会计学", "市场营销", "人力资源管理", "旅游管理", "物流管理", "电子商务", "公共管理"],
        "degree": "管理学学士",
        "duration": 4,
        "base_salary": 7500,
        "employment_rate": 0.89
    },
    "文学": {
        "disciplines": ["汉语言文学", "新闻学", "广告学", "英语", "日语", "翻译", "传播学"],
        "degree": "文学学士",
        "duration": 4,
        "base_salary": 6500,
        "employment_rate": 0.85
    },
    "法学": {
        "disciplines": ["法学", "政治学", "社会学", "国际政治", "知识产权"],
        "degree": "法学学士",
        "duration": 4,
        "base_salary": 8000,
        "employment_rate": 0.87
    },
    "教育学": {
        "disciplines": ["教育学", "学前教育", "小学教育", "特殊教育", "体育教育", "教育技术"],
        "degree": "教育学学士",
        "duration": 4,
        "base_salary": 6000,
        "employment_rate": 0.93
    },
    "艺术学": {
        "disciplines": ["美术学", "设计学", "音乐学", "舞蹈学", "戏剧影视", "动画"],
        "degree": "艺术学学士",
        "duration": 4,
        "base_salary": 7000,
        "employment_rate": 0.82
    },
    "农学": {
        "disciplines": ["农学", "园艺学", "植物保护", "动物科学", "水产养殖", "林学"],
        "degree": "农学学士",
        "duration": 4,
        "base_salary": 5500,
        "employment_rate": 0.88
    }
}

# 具体专业定义（按学科）
SPECIFIC_MAJORS = {
    "机械工程": ["机械工程", "机械设计制造及其自动化", "材料成型及控制工程", "机械电子工程", "工业设计", "车辆工程", "智能制造工程"],
    "电子信息": ["电子信息工程", "通信工程", "微电子科学与工程", "光电信息科学与工程", "信息工程", "集成电路设计与集成系统"],
    "计算机": ["计算机科学与技术", "软件工程", "网络工程", "信息安全", "物联网工程", "数据科学与大数据技术", "人工智能", "网络空间安全"],
    "土木工程": ["土木工程", "建筑学", "城乡规划", "给排水科学与工程", "建筑环境与能源应用工程", "道路桥梁与渡河工程"],
    "材料科学": ["材料科学与工程", "材料物理", "材料化学", "金属材料工程", "高分子材料与工程", "新能源材料与器件"],
    "化学工程": ["化学工程与工艺", "制药工程", "生物工程", "食品科学与工程", "轻化工程"],
    "环境工程": ["环境工程", "环境科学", "资源环境科学", "水质科学与技术", "环保设备工程"],
    "能源动力": ["能源与动力工程", "新能源科学与工程", "核工程与核技术", "储能科学与工程"],
    "自动化": ["自动化", "机器人工程", "轨道交通信号与控制", "智能装备与系统"],
    "航空航天": ["飞行器设计与工程", "飞行器动力工程", "航空航天工程", "探测制导与控制技术"],
    "数学": ["数学与应用数学", "信息与计算科学", "数理基础科学", "数据计算及应用"],
    "物理学": ["物理学", "应用物理学", "核物理", "声学", "光电子科学"],
    "化学": ["化学", "应用化学", "化学生物学", "分子科学与工程"],
    "生物学": ["生物科学", "生物技术", "生物信息学", "生态学"],
    "地理学": ["地理科学", "自然地理与资源环境", "人文地理与城乡规划", "地理信息科学"],
    "天文学": ["天文学"],
    "统计学": ["统计学", "应用统计学", "经济统计学"],
    "心理学": ["心理学", "应用心理学"],
    "临床医学": ["临床医学", "眼视光医学", "精神医学", "儿科学"],
    "口腔医学": ["口腔医学"],
    "中医学": ["中医学", "针灸推拿学", "中西医临床医学"],
    "药学": ["药学", "药物制剂", "临床药学", "中药学"],
    "护理学": ["护理学", "助产学"],
    "公共卫生": ["预防医学", "食品卫生与营养学", "卫生检验与检疫"],
    "医学影像": ["医学影像学", "医学影像技术"],
    "麻醉学": ["麻醉学"],
    "经济学": ["经济学", "经济统计学", "国民经济管理", "资源与环境经济学"],
    "金融学": ["金融学", "金融工程", "保险学", "投资学", "金融科技"],
    "国际经济与贸易": ["国际经济与贸易", "贸易经济"],
    "财政学": ["财政学", "税收学"],
    "投资学": ["投资学"],
    "保险学": ["保险学", "精算学"],
    "工商管理": ["工商管理", "创业管理"],
    "会计学": ["会计学", "财务管理", "审计学"],
    "市场营销": ["市场营销", "国际商务"],
    "人力资源管理": ["人力资源管理", "劳动与社会保障"],
    "旅游管理": ["旅游管理", "酒店管理", "会展经济与管理"],
    "物流管理": ["物流管理", "物流工程", "供应链管理"],
    "电子商务": ["电子商务", "跨境电子商务"],
    "公共管理": ["行政管理", "公共事业管理", "城市管理"],
    "汉语言文学": ["汉语言文学", "汉语言", "秘书学", "古典文献学"],
    "新闻学": ["新闻学", "广播电视学", "网络与新媒体"],
    "广告学": ["广告学"],
    "英语": ["英语", "商务英语"],
    "日语": ["日语"],
    "翻译": ["翻译"],
    "传播学": ["传播学", "编辑出版学"],
    "法学": ["法学", "知识产权"],
    "政治学": ["政治学与行政学", "国际政治", "外交学"],
    "社会学": ["社会学", "社会工作"],
    "国际政治": ["国际政治", "国际事务与国际关系"],
    "知识产权": ["知识产权"],
    "教育学": ["教育学", "科学教育"],
    "学前教育": ["学前教育"],
    "小学教育": ["小学教育"],
    "特殊教育": ["特殊教育"],
    "体育教育": ["体育教育", "运动训练", "社会体育指导与管理"],
    "教育技术": ["教育技术学"],
    "美术学": ["美术学", "绘画", "雕塑", "中国画"],
    "设计学": ["视觉传达设计", "环境设计", "产品设计", "服装与服饰设计", "数字媒体艺术"],
    "音乐学": ["音乐学", "音乐表演", "作曲与作曲技术理论"],
    "舞蹈学": ["舞蹈学", "舞蹈表演", "舞蹈编导"],
    "戏剧影视": ["表演", "戏剧影视文学", "广播电视编导", "戏剧影视导演"],
    "动画": ["动画", "漫画", "游戏设计"],
    "农学": ["农学", "种子科学与工程", "设施农业科学与工程"],
    "园艺学": ["园艺", "茶学"],
    "植物保护": ["植物保护", "植物科学与技术"],
    "动物科学": ["动物科学", "动物医学", "动植物检疫"],
    "水产养殖": ["水产养殖学", "海洋渔业科学与技术"],
    "林学": ["林学", "园林", "森林保护"]
}

# 省份列表
PROVINCES = ["北京", "上海", "天津", "重庆", "河北", "山西", "内蒙古", "辽宁", "吉林", "黑龙江",
             "江苏", "浙江", "安徽", "福建", "江西", "山东", "河南", "湖北", "湖南", "广东",
             "广西", "海南", "四川", "贵州", "云南", "西藏", "陕西", "甘肃", "青海", "宁夏", "新疆"]

# 录取批次
BATCHES = ["本科一批", "本科二批", "本科提前批"]

# 科类
CATEGORIES = ["理工类", "文史类", "综合改革"]


def generate_majors_data(universities_data):
    """生成专业数据"""
    majors = []
    major_id = 1
    major_code_base = 80000

    for uni in universities_data:
        uni_id = uni.get("id", 1)
        uni_level = uni.get("level", "普通本科")
        uni_type = uni.get("type", "综合类")

        # 根据学校类型确定主要专业方向
        primary_categories = []
        if "理工" in uni_type:
            primary_categories = ["工学", "理学"]
        elif "师范" in uni_type:
            primary_categories = ["教育学", "文学", "理学"]
        elif "医药" in uni_type:
            primary_categories = ["医学"]
        elif "财经" in uni_type:
            primary_categories = ["经济学", "管理学"]
        elif "农林" in uni_type:
            primary_categories = ["农学", "理学"]
        elif "政法" in uni_type:
            primary_categories = ["法学", "文学"]
        elif "艺术" in uni_type:
            primary_categories = ["艺术学"]
        else:
            primary_categories = list(MAJOR_CATEGORIES.keys())

        # 为每所学校生成10-30个专业
        num_majors = random.randint(10, 30) if "985" in uni_level or "211" in uni_level else random.randint(5, 15)

        selected_majors = set()
        attempts = 0

        while len(selected_majors) < num_majors and attempts < 100:
            attempts += 1
            category = random.choice(primary_categories if random.random() < 0.7 else list(MAJOR_CATEGORIES.keys()))
            cat_info = MAJOR_CATEGORIES[category]
            discipline = random.choice(cat_info["disciplines"])

            if discipline in SPECIFIC_MAJORS:
                major_name = random.choice(SPECIFIC_MAJORS[discipline])
            else:
                major_name = discipline

            if major_name not in selected_majors:
                selected_majors.add(major_name)

                # 计算就业率和薪资（985/211院校更高）
                level_bonus = 0.05 if "985" in uni_level else (0.03 if "211" in uni_level else 0)
                employment_rate = min(0.99, cat_info["employment_rate"] + level_bonus + random.uniform(-0.05, 0.05))
                salary_bonus = 3000 if "985" in uni_level else (1500 if "211" in uni_level else 0)
                avg_salary = cat_info["base_salary"] + salary_bonus + random.randint(-1000, 2000)

                major = {
                    "id": major_id,
                    "university_id": uni_id,
                    "code": f"{major_code_base + major_id:06d}",
                    "name": major_name,
                    "category": category,
                    "discipline": discipline,
                    "degree_type": cat_info["degree"],
                    "duration": cat_info["duration"],
                    "description": f"{major_name}专业培养具有扎实理论基础和实践能力的高级专门人才。",
                    "employment_rate": round(employment_rate, 3),
                    "average_salary": avg_salary,
                    "is_recruiting": True,
                    "is_active": True,
                    "status": "active"
                }
                majors.append(major)
                major_id += 1

    return majors


def generate_admission_data(universities_data, majors_data):
    """生成录取分数线数据"""
    admissions = []
    admission_id = 1

    # 按学校ID建立专业索引
    uni_majors = {}
    for major in majors_data:
        uid = major["university_id"]
        if uid not in uni_majors:
            uni_majors[uid] = []
        uni_majors[uid].append(major)

    years = [2022, 2023, 2024]  # 近三年数据

    for uni in universities_data:
        uni_id = uni.get("id", 1)
        uni_level = uni.get("level", "普通本科")
        uni_province = uni.get("province", "北京")

        # 基础分数线（根据学校层次）
        if "985" in uni_level:
            base_score = random.randint(620, 680)
            base_rank = random.randint(500, 5000)
        elif "211" in uni_level:
            base_score = random.randint(560, 620)
            base_rank = random.randint(5000, 20000)
        else:
            base_score = random.randint(450, 560)
            base_rank = random.randint(20000, 100000)

        # 获取该校专业
        school_majors = uni_majors.get(uni_id, [])

        for year in years:
            # 年度波动
            year_adjustment = (year - 2022) * random.randint(-5, 8)

            # 选择部分省份生成数据
            target_provinces = [uni_province]  # 本省必有
            other_provinces = [p for p in PROVINCES if p != uni_province]
            target_provinces.extend(random.sample(other_provinces, min(10, len(other_provinces))))

            for province in target_provinces:
                # 省份调整（本省竞争更激烈）
                province_adjustment = -10 if province == uni_province else random.randint(-20, 20)

                for batch in ["本科一批"] if "985" in uni_level or "211" in uni_level else BATCHES[:2]:
                    # 选择部分专业生成数据
                    selected_majors = random.sample(school_majors, min(5, len(school_majors))) if school_majors else []

                    for category in ["理工类", "文史类"]:
                        # 文理科调整
                        category_adjustment = 0 if category == "理工类" else random.randint(-30, 10)

                        # 计算最终分数
                        min_score = base_score + year_adjustment + province_adjustment + category_adjustment
                        avg_score = min_score + random.randint(3, 15)
                        max_score = avg_score + random.randint(5, 25)

                        # 计算排名
                        min_rank = base_rank + (year - 2022) * random.randint(-500, 500)
                        avg_rank = max(100, min_rank - random.randint(500, 2000))
                        max_rank = max(50, avg_rank - random.randint(200, 1000))

                        # 学校整体录取数据
                        admission = {
                            "id": admission_id,
                            "university_id": uni_id,
                            "major_id": None,  # 学校整体数据
                            "year": year,
                            "province": province,
                            "batch": batch,
                            "category": category,
                            "min_score": max(300, min_score),
                            "avg_score": max(300, avg_score),
                            "max_score": max(300, max_score),
                            "min_rank": max(1, min_rank),
                            "avg_rank": max(1, avg_rank),
                            "max_rank": max(1, max_rank),
                            "planned_count": random.randint(20, 200),
                            "actual_count": random.randint(18, 200),
                            "difficulty": "困难" if "985" in uni_level else ("中等" if "211" in uni_level else "较易"),
                            "admission_rate": round(random.uniform(0.01, 0.15) if "985" in uni_level else random.uniform(0.1, 0.5), 3)
                        }
                        admissions.append(admission)
                        admission_id += 1

                        # 为选中的专业生成数据
                        for major in selected_majors:
                            major_adjustment = random.randint(-15, 30)  # 热门专业分更高

                            m_min_score = min_score + major_adjustment
                            m_avg_score = m_min_score + random.randint(2, 10)
                            m_max_score = m_avg_score + random.randint(3, 15)

                            m_min_rank = min_rank + random.randint(-2000, 2000)
                            m_avg_rank = max(100, m_min_rank - random.randint(300, 1000))
                            m_max_rank = max(50, m_avg_rank - random.randint(100, 500))

                            admission = {
                                "id": admission_id,
                                "university_id": uni_id,
                                "major_id": major["id"],
                                "year": year,
                                "province": province,
                                "batch": batch,
                                "category": category,
                                "min_score": max(300, m_min_score),
                                "avg_score": max(300, m_avg_score),
                                "max_score": max(300, m_max_score),
                                "min_rank": max(1, m_min_rank),
                                "avg_rank": max(1, m_avg_rank),
                                "max_rank": max(1, m_max_rank),
                                "planned_count": random.randint(5, 50),
                                "actual_count": random.randint(4, 50),
                                "difficulty": admission["difficulty"],
                                "admission_rate": round(random.uniform(0.01, 0.2), 3)
                            }
                            admissions.append(admission)
                            admission_id += 1

    return admissions


def main():
    print("=" * 60)
    print("高考志愿填报系统 - 数据生成器")
    print("=" * 60)

    # 读取现有高校数据
    universities_file = "/mnt/d/mybitcoin/gaokao/scripts/universities_data.json"
    print(f"\n[1/4] 读取高校数据: {universities_file}")

    with open(universities_file, 'r', encoding='utf-8') as f:
        universities = json.load(f)
    print(f"      加载了 {len(universities)} 所高校")

    # 生成专业数据
    print("\n[2/4] 生成专业数据...")
    majors = generate_majors_data(universities)
    print(f"      生成了 {len(majors)} 个专业")

    # 保存专业数据
    majors_file = "/mnt/d/mybitcoin/gaokao/scripts/majors_data.json"
    with open(majors_file, 'w', encoding='utf-8') as f:
        json.dump(majors, f, ensure_ascii=False, indent=2)
    print(f"      保存到: {majors_file}")

    # 生成录取数据
    print("\n[3/4] 生成录取分数线数据 (2021-2024)...")
    admissions = generate_admission_data(universities, majors)
    print(f"      生成了 {len(admissions)} 条录取数据")

    # 保存录取数据
    admissions_file = "/mnt/d/mybitcoin/gaokao/scripts/admission_data.json"
    with open(admissions_file, 'w', encoding='utf-8') as f:
        json.dump(admissions, f, ensure_ascii=False, indent=2)
    print(f"      保存到: {admissions_file}")

    # 统计信息
    print("\n[4/4] 数据统计:")
    print(f"      高校总数: {len(universities)}")
    print(f"      专业总数: {len(majors)}")
    print(f"      录取数据: {len(admissions)} 条")

    # 专业分类统计
    category_stats = {}
    for m in majors:
        cat = m["category"]
        category_stats[cat] = category_stats.get(cat, 0) + 1

    print("\n      专业分类分布:")
    for cat, count in sorted(category_stats.items(), key=lambda x: -x[1]):
        print(f"        - {cat}: {count} 个")

    # 录取数据年份统计
    year_stats = {}
    for a in admissions:
        year = a["year"]
        year_stats[year] = year_stats.get(year, 0) + 1

    print("\n      录取数据年份分布:")
    for year, count in sorted(year_stats.items()):
        print(f"        - {year}年: {count} 条")

    print("\n" + "=" * 60)
    print("数据生成完成！")
    print("=" * 60)


if __name__ == "__main__":
    main()
