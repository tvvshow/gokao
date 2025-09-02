#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Firecrawl MCP University Data Crawler
使用Firecrawl MCP工具爬取全国高校数据的Python脚本

Author: AI Assistant
Date: 2025-01-22
"""

import json
import os
import sys
import time
import logging
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict
from pathlib import Path

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('firecrawl_crawler.log', encoding='utf-8'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class UniversityInfo:
    """高校信息数据结构"""
    name: str
    province: str
    city: str
    university_type: str  # 本科/专科
    is_985: bool
    is_211: bool
    website: str = ""
    description: str = ""
    established_year: int = 0
    
@dataclass
class ProvinceConfig:
    """省份配置信息"""
    name: str
    code: str
    base_url: str
    search_keywords: List[str]
    
@dataclass
class CrawlResult:
    """爬取结果"""
    province: str
    success: bool
    universities: List[UniversityInfo]
    error_message: str = ""
    urls_discovered: int = 0
    processing_time: float = 0.0

class FirecrawlUniversityCrawler:
    """Firecrawl MCP高校数据爬取器"""
    
    def __init__(self, config_file: str = "province_config.json"):
        self.config_file = config_file
        self.provinces = self._load_province_config()
        self.university_schema = self._create_university_schema()
        
        # 检查Firecrawl API密钥
        if not os.getenv('FIRECRAWL_API_KEY'):
            logger.warning("FIRECRAWL_API_KEY环境变量未设置，可能影响爬取功能")
    
    def _load_province_config(self) -> List[ProvinceConfig]:
        """加载省份配置"""
        try:
            with open(self.config_file, 'r', encoding='utf-8') as f:
                data = json.load(f)
                return [ProvinceConfig(**item) for item in data['provinces']]
        except FileNotFoundError:
            logger.error(f"配置文件 {self.config_file} 不存在")
            return self._get_default_provinces()
        except Exception as e:
            logger.error(f"加载配置文件失败: {e}")
            return self._get_default_provinces()
    
    def _get_default_provinces(self) -> List[ProvinceConfig]:
        """获取默认省份配置"""
        return [
            ProvinceConfig("北京", "BJ", "https://www.bjeea.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("上海", "SH", "https://www.shmeea.edu.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("天津", "TJ", "https://www.zhaokao.net", ["高校", "大学", "学院"]),
            ProvinceConfig("重庆", "CQ", "https://www.cqksy.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("河北", "HE", "https://www.hebeea.edu.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("山西", "SX", "https://www.sxkszx.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("内蒙古", "NM", "https://www.nm.zsks.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("辽宁", "LN", "https://www.lnzsks.com", ["高校", "大学", "学院"]),
            ProvinceConfig("吉林", "JL", "https://www.jleea.edu.cn", ["高校", "大学", "学院"]),
            ProvinceConfig("黑龙江", "HL", "https://www.lzk.hl.cn", ["高校", "大学", "学院"]),
            # 添加更多省份...
        ]
    
    def _create_university_schema(self) -> Dict[str, Any]:
        """创建高校数据提取的JSON Schema"""
        return {
            "type": "object",
            "properties": {
                "universities": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {
                                "type": "string",
                                "description": "高校全称",
                                "minLength": 2
                            },
                            "province": {
                                "type": "string",
                                "description": "所在省份",
                                "enum": ["北京", "上海", "天津", "重庆", "河北", "山西", "内蒙古", 
                                        "辽宁", "吉林", "黑龙江", "江苏", "浙江", "安徽", "福建",
                                        "江西", "山东", "河南", "湖北", "湖南", "广东", "广西",
                                        "海南", "四川", "贵州", "云南", "西藏", "陕西", "甘肃",
                                        "青海", "宁夏", "新疆"]
                            },
                            "city": {
                                "type": "string",
                                "description": "所在城市"
                            },
                            "university_type": {
                                "type": "string",
                                "description": "高校类型",
                                "enum": ["本科", "专科", "独立学院", "民办"]
                            },
                            "is_985": {
                                "type": "boolean",
                                "description": "是否为985高校"
                            },
                            "is_211": {
                                "type": "boolean",
                                "description": "是否为211高校"
                            },
                            "website": {
                                "type": "string",
                                "description": "官方网站",
                                "format": "uri"
                            },
                            "description": {
                                "type": "string",
                                "description": "高校简介"
                            },
                            "established_year": {
                                "type": "integer",
                                "description": "建校年份",
                                "minimum": 1900,
                                "maximum": 2025
                            }
                        },
                        "required": ["name", "province", "city", "university_type", "is_985", "is_211"]
                    }
                }
            },
            "required": ["universities"]
        }
    
    def discover_urls(self, province: ProvinceConfig) -> List[str]:
        """使用firecrawl_map发现相关URL"""
        logger.info(f"开始发现 {province.name} 的相关URL")
        
        # 这里需要调用Firecrawl MCP的firecrawl_map工具
        # 由于这是Python脚本，无法直接调用MCP工具
        # 需要通过其他方式（如HTTP API或命令行）调用
        
        # 模拟URL发现结果
        discovered_urls = [
            f"{province.base_url}/gaoxiao/list.html",
            f"{province.base_url}/university/index.html",
            f"{province.base_url}/colleges/directory.html"
        ]
        
        logger.info(f"发现 {len(discovered_urls)} 个相关URL")
        return discovered_urls
    
    def extract_university_data(self, urls: List[str], province: ProvinceConfig) -> List[UniversityInfo]:
        """使用firecrawl_extract提取高校数据"""
        logger.info(f"开始从 {len(urls)} 个URL提取 {province.name} 的高校数据")
        
        universities = []
        
        # 创建提取提示词
        extraction_prompt = f"""
        请从页面中提取{province.name}省的高校信息，包括：
        1. 高校全称（必须完整准确）
        2. 所在城市
        3. 高校类型（本科/专科/独立学院/民办）
        4. 是否为985工程高校
        5. 是否为211工程高校
        6. 官方网站（如果有）
        7. 高校简介（如果有）
        8. 建校年份（如果有）
        
        请确保提取的数据准确完整，特别注意985/211状态的判断。
        """
        
        # 这里需要调用Firecrawl MCP的firecrawl_extract工具
        # 模拟数据提取结果
        mock_data = {
            "universities": [
                {
                    "name": f"{province.name}大学",
                    "province": province.name,
                    "city": f"{province.name}市",
                    "university_type": "本科",
                    "is_985": province.name in ["北京", "上海"],
                    "is_211": province.name in ["北京", "上海", "天津", "重庆"],
                    "website": f"https://www.{province.code.lower()}u.edu.cn",
                    "description": f"{province.name}省重点综合性大学",
                    "established_year": 1950
                }
            ]
        }
        
        for uni_data in mock_data.get("universities", []):
            universities.append(UniversityInfo(**uni_data))
        
        logger.info(f"成功提取 {len(universities)} 所高校数据")
        return universities
    
    def validate_university_data(self, universities: List[UniversityInfo]) -> List[UniversityInfo]:
        """验证和清理高校数据"""
        logger.info(f"开始验证 {len(universities)} 所高校数据")
        
        valid_universities = []
        seen_names = set()
        
        for uni in universities:
            # 检查必填字段
            if not uni.name or not uni.province or not uni.city:
                logger.warning(f"跳过不完整的高校数据: {uni.name}")
                continue
            
            # 去重
            if uni.name in seen_names:
                logger.warning(f"跳过重复的高校: {uni.name}")
                continue
            
            # 数据清理
            uni.name = uni.name.strip()
            uni.province = uni.province.strip()
            uni.city = uni.city.strip()
            
            valid_universities.append(uni)
            seen_names.add(uni.name)
        
        logger.info(f"验证完成，有效高校数据: {len(valid_universities)} 所")
        return valid_universities
    
    def crawl_province(self, province: ProvinceConfig) -> CrawlResult:
        """爬取单个省份的高校数据"""
        start_time = time.time()
        logger.info(f"开始爬取 {province.name} 的高校数据")
        
        try:
            # 第一阶段：URL发现
            urls = self.discover_urls(province)
            
            # 第二阶段：数据提取
            universities = self.extract_university_data(urls, province)
            
            # 第三阶段：数据验证
            valid_universities = self.validate_university_data(universities)
            
            processing_time = time.time() - start_time
            
            result = CrawlResult(
                province=province.name,
                success=True,
                universities=valid_universities,
                urls_discovered=len(urls),
                processing_time=processing_time
            )
            
            logger.info(f"{province.name} 爬取完成，获得 {len(valid_universities)} 所高校数据，耗时 {processing_time:.2f}秒")
            return result
            
        except Exception as e:
            processing_time = time.time() - start_time
            logger.error(f"{province.name} 爬取失败: {str(e)}")
            
            return CrawlResult(
                province=province.name,
                success=False,
                universities=[],
                error_message=str(e),
                processing_time=processing_time
            )
    
    def save_results(self, results: List[CrawlResult], output_file: str = "crawl_results.json"):
        """保存爬取结果"""
        logger.info(f"保存爬取结果到 {output_file}")
        
        # 转换为可序列化的格式
        serializable_results = []
        for result in results:
            result_dict = asdict(result)
            result_dict['universities'] = [asdict(uni) for uni in result.universities]
            serializable_results.append(result_dict)
        
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump({
                'timestamp': time.strftime('%Y-%m-%d %H:%M:%S'),
                'total_provinces': len(results),
                'successful_provinces': sum(1 for r in results if r.success),
                'total_universities': sum(len(r.universities) for r in results),
                'results': serializable_results
            }, f, ensure_ascii=False, indent=2)
        
        logger.info("结果保存完成")
    
    def generate_report(self, results: List[CrawlResult]) -> Dict[str, Any]:
        """生成爬取报告"""
        total_provinces = len(results)
        successful_provinces = sum(1 for r in results if r.success)
        total_universities = sum(len(r.universities) for r in results)
        total_time = sum(r.processing_time for r in results)
        
        report = {
            'summary': {
                'total_provinces': total_provinces,
                'successful_provinces': successful_provinces,
                'success_rate': f"{successful_provinces/total_provinces*100:.1f}%" if total_provinces > 0 else "0%",
                'total_universities': total_universities,
                'total_processing_time': f"{total_time:.2f}秒",
                'average_time_per_province': f"{total_time/total_provinces:.2f}秒" if total_provinces > 0 else "0秒"
            },
            'province_details': []
        }
        
        for result in results:
            detail = {
                'province': result.province,
                'success': result.success,
                'universities_count': len(result.universities),
                'urls_discovered': result.urls_discovered,
                'processing_time': f"{result.processing_time:.2f}秒",
                'error_message': result.error_message if not result.success else None
            }
            report['province_details'].append(detail)
        
        return report

def main():
    """主函数"""
    logger.info("启动Firecrawl MCP高校数据爬取器")
    
    # 创建爬取器实例
    crawler = FirecrawlUniversityCrawler()
    
    # 检查命令行参数
    if len(sys.argv) > 1:
        # 单省份爬取模式
        province_name = sys.argv[1]
        target_provinces = [p for p in crawler.provinces if p.name == province_name]
        if not target_provinces:
            logger.error(f"未找到省份: {province_name}")
            return
    else:
        # 全部省份爬取模式
        target_provinces = crawler.provinces
    
    logger.info(f"准备爬取 {len(target_provinces)} 个省份的高校数据")
    
    # 执行爬取
    results = []
    for province in target_provinces:
        result = crawler.crawl_province(province)
        results.append(result)
        
        # 添加延迟，避免请求过于频繁
        time.sleep(2)
    
    # 保存结果
    crawler.save_results(results)
    
    # 生成报告
    report = crawler.generate_report(results)
    
    # 输出报告
    logger.info("\n=== 爬取报告 ===")
    logger.info(f"总省份数: {report['summary']['total_provinces']}")
    logger.info(f"成功省份数: {report['summary']['successful_provinces']}")
    logger.info(f"成功率: {report['summary']['success_rate']}")
    logger.info(f"总高校数: {report['summary']['total_universities']}")
    logger.info(f"总耗时: {report['summary']['total_processing_time']}")
    
    # 保存报告
    with open('crawl_report.json', 'w', encoding='utf-8') as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
    
    logger.info("爬取任务完成")

if __name__ == "__main__":
    main()