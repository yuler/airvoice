# Homebrew Tap 设置指南

## 前置条件

1. 创建 GitHub 仓库 `yuler/homebrew-tap`
2. 在 `airvoice` 仓库添加 Secret `TAP_REPO_TOKEN`

## 步骤

### 1. 创建 homebrew-tap 仓库

```bash
# 在 GitHub 上创建新仓库
gh repo create yuler/homebrew-tap --public --description "Homebrew tap for yuler's tools"
```

### 2. 添加 TAP_REPO_TOKEN

1. 去 GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. 创建新 token，权限：
   - Repository access: Only select repositories → `homebrew-tap`
   - Permissions: Contents → Read and write
3. 在 `airvoice` 仓库 Settings → Secrets and variables → Actions 添加：
   - Name: `TAP_REPO_TOKEN`
   - Value: 刚创建的 token

### 3. 发布新版本

```bash
# 更新 version in cli/main.go
# 提交更改
git tag v0.2.0
git push origin v0.2.0
```

工作流会自动：
- 构建 macOS/Linux (ARM64/AMD64) 二进制
- 创建 GitHub Release
- 更新 homebrew-tap 仓库的 formula

### 4. 用户安装

```bash
brew tap yuler/tap
brew install airvoice
```

### 5. 更新版本

每次发布新版本，只需打 tag，工作流会自动更新 formula：
```bash
git tag v0.3.0
git push origin v0.3.0
```

用户升级：
```bash
brew upgrade airvoice
```
