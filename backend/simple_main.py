"""
简化的FastAPI应用 - 用于测试基础功能
Simplified FastAPI Application - For Testing Basic Functionality
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import time
import logging

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 创建FastAPI应用实例
app = FastAPI(
    title="高考志愿填报系统 API (简化版)",
    description="为中国高中学生提供AI驱动的大学申请和职业规划服务",
    version="1.0.0-simple"
)

# 添加CORS中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.middleware("http")
async def add_process_time_header(request, call_next):
    """添加请求处理时间头"""
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response

# 健康检查端点
@app.get("/health")
async def health_check():
    """健康检查"""
    return {
        "status": "healthy",
        "service": "college-entrance-exam-api-simple",
        "version": "1.0.0-simple",
        "timestamp": time.time()
    }

@app.get("/")
async def root():
    """根路径"""
    return {
        "message": "欢迎使用高考志愿填报系统 API (简化版)",
        "docs": "/docs",
        "health": "/health",
        "features": [
            "✅ 基础API框架",
            "✅ 健康检查",
            "✅ CORS支持",
            "✅ 请求时间统计",
            "🚧 用户认证 (开发中)",
            "🚧 数据库集成 (开发中)",
            "🚧 AI推荐引擎 (开发中)"
        ]
    }

# API v1 路由
@app.get("/api/v1/ping")
async def ping():
    """API健康检查"""
    return {
        "message": "pong",
        "status": "healthy",
        "timestamp": time.time()
    }

@app.get("/api/v1/test")
async def test_endpoint():
    """测试端点"""
    return {
        "message": "高考志愿填报系统 API v1 (简化版)",
        "features": [
            "用户认证与授权",
            "高校专业查询",
            "AI智能推荐",
            "志愿模拟填报",
            "就业前景分析"
        ],
        "status": "开发中",
        "database_status": "未连接 (简化版)",
        "redis_status": "未连接 (简化版)"
    }

# 模拟数据端点
@app.get("/api/v1/colleges")
async def get_colleges():
    """获取高校列表 (模拟数据)"""
    return {
        "data": [
            {
                "id": 1,
                "name": "清华大学",
                "province": "北京",
                "level": "985",
                "type": "综合"
            },
            {
                "id": 2,
                "name": "北京大学",
                "province": "北京",
                "level": "985",
                "type": "综合"
            },
            {
                "id": 3,
                "name": "浙江大学",
                "province": "浙江",
                "level": "985",
                "type": "综合"
            }
        ],
        "total": 3,
        "note": "这是模拟数据，实际数据需要连接数据库"
    }

@app.get("/api/v1/majors")
async def get_majors():
    """获取专业列表 (模拟数据)"""
    return {
        "data": [
            {
                "id": 1,
                "name": "计算机科学与技术",
                "category": "工学",
                "code": "080901"
            },
            {
                "id": 2,
                "name": "软件工程",
                "category": "工学",
                "code": "080902"
            },
            {
                "id": 3,
                "name": "人工智能",
                "category": "工学",
                "code": "080717T"
            }
        ],
        "total": 3,
        "note": "这是模拟数据，实际数据需要连接数据库"
    }

if __name__ == "__main__":
    import uvicorn
    logger.info("🚀 启动高考志愿填报系统 API (简化版)...")
    uvicorn.run(
        "simple_main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info"
    )