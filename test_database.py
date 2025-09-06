#!/usr/bin/env python3
"""
数据库连接测试脚本
Database Connection Test Script
"""

import sys
import os

def test_postgres_connection():
    """测试PostgreSQL连接"""
    print("🔍 测试PostgreSQL连接...")
    
    try:
        import psycopg2
        
        # 连接参数
        conn_params = {
            'host': 'localhost',
            'port': 5433,
            'database': 'gaokao_db',
            'user': 'postgres',
            'password': 'password'
        }
        
        # 尝试连接
        conn = psycopg2.connect(**conn_params)
        cursor = conn.cursor()
        
        # 执行测试查询
        cursor.execute("SELECT version();")
        version = cursor.fetchone()
        
        print(f"✅ PostgreSQL连接成功")
        print(f"   版本: {version[0]}")
        
        # 测试创建表
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS test_table (
                id SERIAL PRIMARY KEY,
                name VARCHAR(100),
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
        """)
        
        # 插入测试数据
        cursor.execute("INSERT INTO test_table (name) VALUES (%s);", ("测试数据",))
        conn.commit()
        
        # 查询测试数据
        cursor.execute("SELECT * FROM test_table;")
        rows = cursor.fetchall()
        print(f"   测试表记录数: {len(rows)}")
        
        # 清理测试表
        cursor.execute("DROP TABLE test_table;")
        conn.commit()
        
        cursor.close()
        conn.close()
        
        return True
        
    except ImportError:
        print("❌ psycopg2未安装，请运行: pip install psycopg2-binary")
        return False
    except Exception as e:
        print(f"❌ PostgreSQL连接失败: {e}")
        return False

def test_redis_connection():
    """测试Redis连接"""
    print("🔍 测试Redis连接...")
    
    try:
        import redis
        
        # 连接Redis
        r = redis.Redis(host='localhost', port=6380, db=0, decode_responses=True)
        
        # 测试连接
        r.ping()
        print("✅ Redis连接成功")
        
        # 测试基本操作
        r.set('test_key', '测试值')
        value = r.get('test_key')
        print(f"   测试读写: {value}")
        
        # 清理测试数据
        r.delete('test_key')
        
        return True
        
    except ImportError:
        print("❌ redis未安装，请运行: pip install redis")
        return False
    except Exception as e:
        print(f"❌ Redis连接失败: {e}")
        return False

def main():
    """主函数"""
    print("🚀 高考志愿填报系统 - 数据库连接测试")
    print("=" * 50)
    
    tests = [
        test_postgres_connection,
        test_redis_connection
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        try:
            if test():
                passed += 1
            print()
        except Exception as e:
            print(f"❌ 测试异常: {e}")
            print()
    
    print("=" * 50)
    print(f"📊 测试结果: {passed}/{total} 通过")
    
    if passed == total:
        print("🎉 数据库连接测试全部通过！")
        print("📝 下一步:")
        print("   1. 安装Python依赖: pip install -r backend/requirements.txt")
        print("   2. 启动FastAPI服务: cd backend && python -m uvicorn app.main:app --reload")
        print("   3. 访问API文档: http://localhost:8000/docs")
        return True
    else:
        print("⚠️  部分测试失败")
        print("💡 解决方案:")
        print("   - 安装缺失的Python包")
        print("   - 检查Docker服务是否正常运行")
        print("   - 确认端口没有被其他程序占用")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)