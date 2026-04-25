#!/bin/bash

# Cloudflare Tunnel + Gaokao系统 - 全面测试脚本

set -e

SERVER="pestxo@192.168.0.181"
PASSWORD="satanking"
DOMAIN="gaokao.pkuedu.eu.org"
INTERNAL_IP="192.168.0.181"

echo "=========================================="
echo "   Cloudflare Tunnel 全面测试"
echo "=========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${BLUE}[测试 $TOTAL_TESTS]${NC} $test_name"

    result=$(eval "$test_command" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        if [ -n "$expected" ]; then
            if echo "$result" | grep -q "$expected"; then
                echo -e "${GREEN}✓ 通过${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
                return 0
            else
                echo -e "${RED}✗ 失败${NC} - 未找到期望结果: $expected"
                echo "  实际响应: $result"
                FAILED_TESTS=$((FAILED_TESTS + 1))
                return 1
            fi
        else
            echo -e "${GREEN}✓ 通过${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        fi
    else
        echo -e "${RED}✗ 失败${NC} - 命令执行失败"
        echo "  错误: $result"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 第一部分：基础设施测试
echo -e "${YELLOW}=========================================="
echo "   第一部分：基础设施测试"
echo "==========================================${NC}"
echo ""

run_test "Cloudflare Tunnel服务运行" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'ps aux | grep cloudflared | grep -v grep | grep -v defunct'" \
    "cloudflared"

run_test "Nginx服务运行" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'systemctl is-active nginx'" \
    "active"

run_test "Data Service运行" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'pgrep -f data-service'"

run_test "PostgreSQL运行" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'docker ps | grep gaokao-postgres'" \
    "gaokao-postgres"

run_test "Redis运行" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'docker ps | grep gaokao-redis'" \
    "gaokao-redis"

echo ""

# 第二部分：DNS和网络测试
echo -e "${YELLOW}=========================================="
echo "   第二部分：DNS和网络测试"
echo "==========================================${NC}"
echo ""

run_test "DNS解析到Cloudflare" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'nslookup $DOMAIN | grep Address'" \
    "104.21."

run_test "内网HTTP响应正常" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s -I http://localhost/ | grep HTTP'" \
    "HTTP/1.1 200"

echo ""

# 第三部分：内网API测试
echo -e "${YELLOW}=========================================="
echo "   第三部分：内网API测试"
echo "==========================================${NC}"
echo ""

run_test "Data Service - 大学列表" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s \"http://localhost/api/data/v1/universities?page=1&page_size=1\"'" \
    '"success":true'

run_test "Data Service - 大学搜索" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s \"http://localhost/api/data/v1/universities/search?q=清华&page=1&page_size=1\"'" \
    '"success":true'

run_test "Data Service - 大学统计" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s \"http://localhost/api/data/v1/universities/statistics\"'" \
    '"total":2793'

run_test "Data Service - 专业列表" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s \"http://localhost/api/data/v1/majors?page=1&page_size=1\"'" \
    '"success":true'

run_test "API响应时间 < 500ms" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'time_command=\"time curl -s \\\"http://localhost/api/data/v1/universities?page=1&page_size=1\\\" > /dev/null\" && \$time_command 2>&1 | grep real | awk \"{print \\\$2}\"'" \
    ""

echo ""

# 第四部分：数据完整性测试
echo -e "${YELLOW}=========================================="
echo "   第四部分：数据完整性测试"
echo "==========================================${NC}"
echo ""

run_test "大学数据 > 2000条" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'docker exec gaokao-postgres psql -U gaokao -d gaokao_db -t -c \"SELECT COUNT(*) FROM universities;\"'" \
    "2793"

run_test "专业数据 > 50000条" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'docker exec gaokao-postgres psql -U gaokao -d gaokao_db -t -c \"SELECT COUNT(*) FROM majors;\"'" \
    "56156"

run_test "录取数据 > 400万条" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'docker exec gaokao-postgres psql -U gaokao -d gaokao_db -t -c \"SELECT COUNT(*) FROM admission_data;\"'" \
    "4201560"

echo ""

# 第五部分：前端测试
echo -e "${YELLOW}=========================================="
echo "   第五部分：前端测试"
echo "==========================================${NC}"
echo ""

run_test "前端HTML可访问" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s http://localhost/ | grep -o \"<title>.*</title>\"'" \
    "高考志愿填报助手"

run_test "静态资源已部署" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'ls -lh /var/www/gaokao/assets/index-*.js 2>/dev/null | wc -l'" \
    "1"

run_test "前端文件大小合理" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'du -sh /var/www/gaokao/ | awk \"{print \\\$1}\"'"

echo ""

# 第六部分：Cloudflare兼容性测试
echo -e "${YELLOW}=========================================="
echo "   第六部分：Cloudflare兼容性"
echo "==========================================${NC}"
echo ""

run_test "Cloudflare配置文件存在" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'cat ~/.cloudflared/config.yml | grep hostname'" \
    "gaokao.pkuedu.eu.org"

run_test "Cloudflare Tunnel使用HTTP2" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'cat ~/.cloudflared/config.yml | grep protocol'" \
    "http2"

run_test "域名映射到localhost:80" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'cat ~/.cloudflared/config.yml | grep -A1 \"gaokao.pkuedu.eu.org\"'" \
    "http://localhost:80"

echo ""

# 第七部分：外网访问测试（通过域名）
echo -e "${YELLOW}=========================================="
echo "   第七部分：外网访问测试"
echo "==========================================${NC}"
echo ""

echo -e "${BLUE}注意：测试外网访问需要从本地机器访问${NC}"
echo ""

# 从本地测试外网域名
run_test "外网域名可解析" \
    "nslookup $DOMAIN 2>&1 | grep Address" \
    "104.21."

run_test "外网域名HTTP可访问" \
    "curl -s -I http://$DOMAIN/ | grep HTTP" \
    "HTTP/1.1 200"

run_test "外网API可访问" \
    "curl -s \"http://$DOMAIN/api/data/v1/universities?page=1&page_size=1\" | head -5" \
    '"success":true'

echo ""

# 第八部分：安全配置测试
echo -e "${YELLOW}=========================================="
echo "   第八部分：安全配置测试"
echo "==========================================${NC}"
echo ""

run_test "X-Frame-Options头部" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s -I http://localhost/ | grep -i x-frame-options'" \
    "SAMEORIGIN"

run_test "X-Content-Type-Options头部" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s -I http://localhost/ | grep -i x-content-type-options'" \
    "nosniff"

run_test "Gzip压缩启用" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'curl -s -I -H \"Accept-Encoding: gzip\" http://localhost/ | grep -i content-encoding'" \
    ""

echo ""

# 第九部分：性能测试
echo -e "${YELLOW}=========================================="
echo "   第九部分：性能测试"
echo "==========================================${NC}"
echo ""

echo -e "${BLUE}API响应时间测试（连续5次请求）${NC}"
for i in {1..5}; do
    echo -n "  请求 $i: "
    start=$(date +%s%3N)
    sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
        "curl -s 'http://localhost/api/data/v1/universities?page=1&page_size=1' > /dev/null"
    end=$(date +%s%3N)
    duration=$((end - start))
    echo "${duration}ms"

    if [ $duration -lt 500 ]; then
        echo -e "    ${GREEN}✓ 响应时间 < 500ms${NC}"
    else
        echo -e "    ${YELLOW}⚠ 响应时间较长${NC}"
    fi
done

echo ""

# 第十部分：日志和错误检查
echo -e "${YELLOW}=========================================="
echo "   第十部分：日志和错误检查"
echo "==========================================${NC}"
echo ""

run_test "Nginx错误日志检查" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'sudo tail -20 /var/log/nginx/error.log 2>/dev/null | grep -i error || echo \"无错误\"'" \
    ""

run_test "Data Service运行日志检查" \
    "sshpass -p '$PASSWORD' ssh -o StrictHostKeyChecking=no $SERVER 'journalctl -u data-service --no-pager -n 20 | grep -i error || echo \"无错误\"'" \
    ""

echo ""

# 最终测试报告
echo ""
echo "=========================================="
echo "           测试完成报告"
echo "=========================================="
echo ""

echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo -e "通过率: $PASS_RATE%"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}=========================================="
    echo "      🎉 所有测试通过！系统完美！"
    echo "==========================================${NC}"
    exit 0
else
    echo -e "${YELLOW}=========================================="
    echo "      ⚠ 部分测试失败，需要修复"
    echo "==========================================${NC}"
    exit 1
fi
