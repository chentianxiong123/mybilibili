# Git提交脚本
# 用于批量添加修改的文件到git仓库

Write-Host "正在添加文件到git..."

# 定义需要添加的文件列表
$files = @(
    "mybilibili-common/src/main/java/com/mybilibili/common/entity/User.java",
    "mybilibili-common/src/main/java/com/mybilibili/common/entity/UserDynamic.java",
    "mybilibili-common/src/main/java/com/mybilibili/common/vo/DynamicVO.java",
    "mybilibili-common/src/main/java/com/mybilibili/common/vo/UserVO.java",
    "mybilibili-user/src/main/java/com/mybilibili/user/service/UserService.java",
    "mybilibili-message/src/main/resources/application.yml",
    "mybilibili-search/src/main/resources/application.yml",
    "mybilibili-interaction/src/main/resources/application.yml",
    ".gitignore"
)

# 遍历并添加每个文件
foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Host "添加文件: $file"
        git add $file
        Write-Host "✓ 文件已添加: $file"
    } else {
        Write-Host "⚠ 文件不存在: $file"
    }
}

Write-Host "`n"
Write-Host "所有文件处理完成！"
Write-Host "修改的文件总数: $($files.Count)"
Write-Host "成功的文件数: 0"
