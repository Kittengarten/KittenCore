param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "clean", "help", "all")]
    [string]$Task = "all"
)

function Build {
    Write-Host "复制 main.go 到内嵌资源……"
    $destDir = "internal/config/prio/"
    if (-not (Test-Path $destDir)) {
        New-Item -ItemType Directory -Force -Path $destDir | Out-Null
    }
    Copy-Item "main.go" -Destination "$destDir" -Force

    Write-Host "编译 KittenCore……"
    go build -ldflags="-s -w" -v -o "KittenCore.exe" -trimpath
}

function Clean {
    Write-Host "清理构建产物……"
    if (Test-Path "KittenCore.exe") {
        Remove-Item "KittenCore.exe" -Force
    }

    # 清理 Go 缓存
    go clean

    # 删除数据文件（如果存在）
    $paths = @(
        "data/ai/user.yaml",
        "data/zbp/banwords.yaml",
        "data/Stack2/tips.yaml"
    )
    foreach ($path in $paths) {
        if (Test-Path $path) {
            Remove-Item $path -Force
        }
    }
}

function Help {
    Write-Host ""
    Write-Host "使用方法："
    Write-Host "  ./build.ps1 build    编译 KittenCore"
    Write-Host "  ./build.ps1 clean    清理构建及缓存文件"
    Write-Host "  ./build.ps1 all      清理后重新编译（默认）"
    Write-Host "  ./build.ps1 help     显示帮助信息"
    Write-Host ""
}

# 主逻辑分发
switch ($Task) {
    "build" { Build }
    "clean" { Clean }
    "help" { Help }
    "all" {
        Clean
        Build
    }
}
