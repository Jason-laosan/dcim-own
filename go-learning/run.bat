@echo off
echo === TaskHub - Go 学习项目 ===
echo 正在启动...
echo.

cd /d %~dp0

:: 检查 Go 是否安装
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go。请先安装 Go 1.21 或更高版本。
    pause
    exit /b 1
)

:: 运行项目
go run cmd/server/main.go

pause
