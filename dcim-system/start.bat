@echo off
chcp 65001 >nul
echo ====================================
echo DCIM系统快速启动脚本
echo ====================================
echo.

echo [1/4] 检查Docker是否运行...
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker未运行，请先启动Docker Desktop
    pause
    exit /b 1
)
echo Docker运行正常

echo.
echo [2/4] 检查docker-compose是否可用...
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: docker-compose未安装
    pause
    exit /b 1
)
echo docker-compose可用

echo.
echo [3/4] 初始化配置文件...
if not exist "collector-agent\config.yaml" (
    copy "collector-agent\config.yaml.example" "collector-agent\config.yaml" >nul
    echo 已创建collector-agent配置文件
)
if not exist "services\collector-mgmt\config.yaml" (
    copy "services\collector-mgmt\config.yaml.example" "services\collector-mgmt\config.yaml" >nul
    echo 已创建collector-mgmt配置文件
)
if not exist "services\data-processor\config.yaml" (
    copy "services\data-processor\config.yaml.example" "services\data-processor\config.yaml" >nul
    echo 已创建data-processor配置文件
)

echo.
echo [4/4] 启动所有服务...
docker-compose up -d

if %errorlevel% equ 0 (
    echo.
    echo ====================================
    echo 服务启动成功！
    echo ====================================
    echo.
    echo 访问地址：
    echo   - EMQX Dashboard: http://localhost:18083 ^(admin/public^)
    echo   - InfluxDB UI:    http://localhost:8086
    echo   - MinIO Console:  http://localhost:9001 ^(minioadmin/minioadmin^)
    echo   - 采集管理服务:    http://localhost:8080
    echo.
    echo 查看日志命令：
    echo   docker-compose logs -f
    echo.
    echo 停止服务命令：
    echo   docker-compose down
    echo.
) else (
    echo.
    echo 错误: 服务启动失败
    echo 请检查docker-compose.yml配置
)

pause
