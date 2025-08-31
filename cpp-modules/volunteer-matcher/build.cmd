@echo off
REM 高考志愿填报系统 - C++志愿匹配引擎构建脚本 (Windows)
REM 作者: 高考志愿填报系统开发团队
REM 版本: 1.0.0
REM 日期: 2025-01-18

setlocal enabledelayedexpansion

REM 颜色代码 (在Windows 10+中支持)
set "RED=[31m"
set "GREEN=[32m"
set "YELLOW=[33m"
set "BLUE=[34m"
set "NC=[0m"

REM 日志函数
:log_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:log_warn
echo %YELLOW%[WARN]%NC% %~1
goto :eof

:log_error
echo %RED%[ERROR]%NC% %~1
goto :eof

:log_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

REM 检查命令是否存在
:check_command
where %~1 >nul 2>&1
if errorlevel 1 (
    call :log_error "Command '%~1' not found. Please install %~1 first."
    exit /b 1
)
goto :eof

REM 检查依赖
:check_dependencies
call :log_info "Checking dependencies..."

call :check_command cmake
if errorlevel 1 exit /b 1

call :check_command cl
if errorlevel 1 (
    call :check_command g++
    if errorlevel 1 (
        call :log_error "No C++ compiler found. Please install Visual Studio or MinGW."
        exit /b 1
    ) else (
        call :log_info "Found g++ compiler (MinGW)"
        set "COMPILER=MinGW"
    )
) else (
    call :log_info "Found MSVC compiler"
    set "COMPILER=MSVC"
)

REM 检查vcpkg (可选)
where vcpkg >nul 2>&1
if not errorlevel 1 (
    call :log_info "Found vcpkg package manager"
    set "USE_VCPKG=1"
) else (
    call :log_warn "vcpkg not found. Using system libraries."
    set "USE_VCPKG=0"
)

call :log_success "Dependencies check completed"
goto :eof

REM 清理构建目录
:clean_build
call :log_info "Cleaning build directory..."
if exist build (
    rmdir /s /q build
)
mkdir build
call :log_success "Build directory cleaned"
goto :eof

REM CMake配置
:configure_cmake
set BUILD_TYPE=%~1
if "%BUILD_TYPE%"=="" set BUILD_TYPE=Release

call :log_info "Configuring CMake..."
call :log_info "Build type: %BUILD_TYPE%"

cd build

REM CMake选项
set CMAKE_OPTIONS=-DCMAKE_BUILD_TYPE=%BUILD_TYPE% -DBUILD_TESTS=ON -DBUILD_EXAMPLES=ON -DBUILD_GO_BINDINGS=ON

REM 根据编译器选择生成器
if "%COMPILER%"=="MSVC" (
    set CMAKE_OPTIONS=%CMAKE_OPTIONS% -G "Visual Studio 17 2022" -A x64
) else (
    set CMAKE_OPTIONS=%CMAKE_OPTIONS% -G "MinGW Makefiles"
)

REM 如果使用vcpkg
if "%USE_VCPKG%"=="1" (
    if defined VCPKG_ROOT (
        set CMAKE_OPTIONS=%CMAKE_OPTIONS% -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake
    )
)

REM 运行CMake
cmake %CMAKE_OPTIONS% ..
if errorlevel 1 (
    call :log_error "CMake configuration failed"
    cd ..
    exit /b 1
)

call :log_success "CMake configuration successful"
cd ..
goto :eof

REM 编译项目
:build_project
call :log_info "Building project..."
cd build

REM 检测CPU核心数
for /f "tokens=2 delims==" %%i in ('wmic cpu get NumberOfLogicalProcessors /value ^| find "="') do set CORES=%%i
if "%CORES%"=="" set CORES=4

call :log_info "Using %CORES% parallel jobs"

if "%COMPILER%"=="MSVC" (
    cmake --build . --config %BUILD_TYPE% --parallel %CORES%
) else (
    mingw32-make -j%CORES%
)

if errorlevel 1 (
    call :log_error "Build failed"
    cd ..
    exit /b 1
)

call :log_success "Build completed successfully"
cd ..
goto :eof

REM 运行测试
:run_tests
call :log_info "Running tests..."
cd build

if "%BUILD_TYPE%"=="Debug" (
    set TEST_DIR=Debug
) else (
    set TEST_DIR=Release
)

if exist "%TEST_DIR%\volunteer_matcher_tests.exe" (
    "%TEST_DIR%\volunteer_matcher_tests.exe"
    if errorlevel 1 (
        call :log_error "Some tests failed"
        cd ..
        exit /b 1
    )
    call :log_success "All tests passed"
) else if exist "volunteer_matcher_tests.exe" (
    volunteer_matcher_tests.exe
    if errorlevel 1 (
        call :log_error "Some tests failed"
        cd ..
        exit /b 1
    )
    call :log_success "All tests passed"
) else (
    call :log_warn "Test executable not found"
)

cd ..
goto :eof

REM 安装库文件
:install_library
call :log_info "Installing library..."
cd build

cmake --install . --config %BUILD_TYPE%
if errorlevel 1 (
    call :log_warn "Install failed, trying local install"
    cmake --install . --config %BUILD_TYPE% --prefix ./install_local
)

call :log_success "Library installed"
cd ..
goto :eof

REM 生成Go绑定
:generate_go_bindings
call :log_info "Generating Go bindings..."
cd build

if exist "bindings\go" (
    call :log_info "Go bindings generated in build\bindings\go\"
    dir bindings\go
) else (
    call :log_warn "Go bindings not found"
)

cd ..
goto :eof

REM 显示构建信息
:show_build_info
call :log_info "Build Summary:"
echo ======================================

if exist build (
    cd build
    
    if "%BUILD_TYPE%"=="Debug" (
        set LIB_DIR=Debug
    ) else (
        set LIB_DIR=Release
    )
    
    if exist "%LIB_DIR%\volunteer_matcher.lib" (
        for %%i in ("%LIB_DIR%\volunteer_matcher.lib") do set SIZE=%%~zi
        call :log_info "Static library: volunteer_matcher.lib (!SIZE! bytes)"
    ) else if exist "libvolunteer_matcher.a" (
        for %%i in ("libvolunteer_matcher.a") do set SIZE=%%~zi
        call :log_info "Static library: libvolunteer_matcher.a (!SIZE! bytes)"
    )
    
    if exist "%LIB_DIR%\volunteer_matcher.dll" (
        for %%i in ("%LIB_DIR%\volunteer_matcher.dll") do set SIZE=%%~zi
        call :log_info "Shared library: volunteer_matcher.dll (!SIZE! bytes)"
    ) else if exist "libvolunteer_matcher.dll" (
        for %%i in ("libvolunteer_matcher.dll") do set SIZE=%%~zi
        call :log_info "Shared library: libvolunteer_matcher.dll (!SIZE! bytes)"
    )
    
    if exist "%LIB_DIR%\volunteer_matcher_example.exe" (
        call :log_info "Example executable: volunteer_matcher_example.exe"
    ) else if exist "volunteer_matcher_example.exe" (
        call :log_info "Example executable: volunteer_matcher_example.exe"
    )
    
    if exist "%LIB_DIR%\volunteer_matcher_tests.exe" (
        call :log_info "Test executable: volunteer_matcher_tests.exe"
    ) else if exist "volunteer_matcher_tests.exe" (
        call :log_info "Test executable: volunteer_matcher_tests.exe"
    )
    
    cd ..
)

echo ======================================
call :log_success "Build completed successfully!"
goto :eof

REM 显示帮助信息
:show_help
echo Usage: %~nx0 [OPTIONS]
echo Options:
echo   --debug    Build in debug mode
echo   --test     Run tests after build
echo   --install  Install library after build
echo   --clean    Clean build directory before build
echo   --help     Show this help message
goto :eof

REM 主函数
:main
call :log_info "Starting C++ Volunteer Matcher build process..."

REM 解析命令行参数
set BUILD_TYPE=Release
set RUN_TESTS=0
set INSTALL=0
set CLEAN=0

:parse_args
if "%~1"=="" goto end_parse
if "%~1"=="--debug" (
    set BUILD_TYPE=Debug
    shift
    goto parse_args
)
if "%~1"=="--test" (
    set RUN_TESTS=1
    shift
    goto parse_args
)
if "%~1"=="--install" (
    set INSTALL=1
    shift
    goto parse_args
)
if "%~1"=="--clean" (
    set CLEAN=1
    shift
    goto parse_args
)
if "%~1"=="--help" (
    call :show_help
    exit /b 0
)
call :log_error "Unknown option: %~1"
exit /b 1

:end_parse

REM 执行构建步骤
call :check_dependencies
if errorlevel 1 exit /b 1

if "%CLEAN%"=="1" (
    call :clean_build
    if errorlevel 1 exit /b 1
)

if not exist build mkdir build

call :configure_cmake "%BUILD_TYPE%"
if errorlevel 1 exit /b 1

call :build_project
if errorlevel 1 exit /b 1

call :generate_go_bindings
if errorlevel 1 exit /b 1

if "%RUN_TESTS%"=="1" (
    call :run_tests
    if errorlevel 1 exit /b 1
)

if "%INSTALL%"=="1" (
    call :install_library
    if errorlevel 1 exit /b 1
)

call :show_build_info

call :log_success "Build process completed!"
goto :eof

REM 错误处理
if not "%~1"=="" goto main
call :main %*