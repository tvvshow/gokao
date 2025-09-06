@echo off
echo 📊 查看服务状态...
echo.

docker-compose ps

echo.
echo 💡 服务说明:
echo   postgres - PostgreSQL数据库
echo   redis    - Redis缓存
echo   backend  - FastAPI后端服务
echo   nginx    - Nginx API网关

pause