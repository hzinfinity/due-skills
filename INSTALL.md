# due-skills 安装指南

due-skills 是一个专为 due 游戏服务器框架 (v2.5.2) 设计的 Claude Code 技能包。

## 前置要求

- Node.js >= 18.0.0
- npm >= 9.0.0
- Claude Code (通过 `npm install -g @anthropic/claude-code` 安装)

## 安装方法

### 方法 1: 从本地目录安装（开发/测试）

1. 克隆或下载此仓库：

```bash
git clone https://github.com/hzInfinity/skills.git
cd skills/due-skills
```

2. 使用 @anthropic/skills 安装：

```bash
npx @anthropic/skills install .
```

### 方法 2: 从 npm 安装（推荐 - 发布后）

一旦包发布到 npm：

```bash
npx @anthropic/skills install @hzinfinity/due-skills
```

### 方法 3: 手动安装

#### 项目级别（仅当前项目可用）

```bash
# 在项目根目录执行
git clone https://github.com/hzInfinity/skills.git .claude/skills/due-skills
```

#### 全局级别（所有项目可用）

```bash
# 用户主目录
git clone https://github.com/hzInfinity/skills.git ~/.claude/skills/due-skills

# Windows PowerShell
git clone https://github.com/hzInfinity/skills.git "$HOME\.claude\skills\due-skills"
```

### 方法 4: 从 tarball 安装

1. 下载或创建 `.tgz` 包：

```bash
cd due-skills
npm pack
# 生成 hzinfinity-due-skills-1.0.0.tgz
```

2. 安装包：

```bash
npx @anthropic/skills install ./hzinfinity-due-skills-1.0.0.tgz
```

## 验证安装

安装完成后，在 Claude Code 中：

1. 打开一个包含 due 框架代码的项目
2. 尝试询问 due 相关问题，例如：
   ```
   如何用 due 创建一个 WebSocket 网关？
   ```
3. 或者手动调用：
   ```
   /due-skills
   ```

## 卸载

```bash
# 删除技能目录
rm -rf ~/.claude/skills/due-skills
# 或者项目级别
rm -rf .claude/skills/due-skills
```

## 更新

```bash
# 拉取最新代码
cd due-skills
git pull origin main

# 重新安装
npx @anthropic/skills install .
```

## 发布到 npm（维护者）

```bash
# 1. 更新版本号（在 package.json 中）
npm version patch  # 或 minor/major

# 2. 发布（需要登录）
npm login
npm publish --access public
```

## 故障排除

### 技能未自动加载

确保：
1. 技能目录位于正确的路径（`~/.claude/skills/` 或 `.claude/skills/`）
2. SKILL.md 文件存在且格式正确
3. 重启 Claude Code

### npx skills 命令不存在

安装 @anthropic/skills：

```bash
npm install -g @anthropic/skills
```

### 权限问题（Linux/macOS）

```bash
# 可能需要修复权限
sudo chown -R $USER:$USER ~/.claude/skills
```

## 支持

如有问题，请提交 issue 到：https://github.com/hzInfinity/skills/issues
