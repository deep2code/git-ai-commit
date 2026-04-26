# <div align="center">🤖 Git AI Commit</div>

<div align="center">

**智能生成 Git 提交信息的 CLI 工具**，基于 AI 模型自动分析代码变更，生成符合 Conventional Commits 规范的简洁提交信息。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/deep2code/git-ai-commit/pulls)

</div>

---

## ✨ 功能特性

| 特性 | 说明 |
|------|------|
| 🌐 **多 AI 提供商支持** | 支持智谱 GLM、百度千帆、SiliconFlow、ModelScope、阿里 Qwen、DeepSeek、腾讯等多种 AI 服务 |
| ⚡ **双模式运行** | CLI 模式：直接命令行调用<br>Git Hook 模式：作为 commit-msg hook 使用 |
| 🧠 **智能变更压缩** | 自动过滤锁文件、二进制文件等无需关注的变更 |
| 📏 **极简输出** | 强制限制提交信息在 30 字符以内，符合 Conventional Commits 规范 |

---

## 🚀 快速开始

### 安装

```bash
git clone https://github.com/deep2code/git-ai-commit.git
cd git-ai-commit
./build.sh
```

### 使用方法

#### 1️⃣ 下载并编译工具

```bash
git clone https://github.com/deep2code/git-ai-commit.git
cd git-ai-commit
./build.sh
# 将编译后的二进制文件添加到 PATH
cp git-ai-commit /usr/local/bin/  # 或其他 PATH 目录
```

#### 2️⃣ 配置环境变量

设置对应 AI 提供商的 API Key：

```bash
# 示例：设置智谱 API Key
export ZHIPU_API_KEY=your-api-key-here

# 或设置其他提供商
# export BAIDU_API_KEY=your-api-key-here
# export SILICONFLOW_API_KEY=your-api-key-here
# export MODELSCOPE_API_KEY=your-api-key-here
# export QWEN_API_KEY=your-api-key-here
# export TENCENT_API_KEY=your-api-key-here
# export VOLCENGINE_API_KEY=your-api-key-here
# export DEEPSEEK_API_KEY=your-api-key-here
```

#### 3️⃣ 一键提交

```bash
git add . && git commit -m "$(git-ai-commit)" && git push
```

#### 4️⃣ Git Hook 模式

```bash
# 只需做一次：创建 hook 链接
ln -sf /path/to/git-ai-commit ~/.git-template/hooks/commit-msg
# 或直接链接到项目
ln -sf /path/to/git-ai-commit .git/hooks/commit-msg

# 使用时只需正常 git commit，工具会自动生成提交信息
git add .
git commit
```

---

## 🔧 开发调试 - CLI 模式

```bash
# 使用指定 API Key 和提供商
git-ai-commit --zhipu-api-key <your-key>
git-ai-commit --qwen-api-key <your-key>
git-ai-commit --baidu-api-key <your-key>
git-ai-commit --siliconflow-api-key <your-key>
git-ai-commit --modelscope-api-key <your-key>
git-ai-commit --tencent-api-key <your-key>

# 指定提供商、模型和端点
git-ai-commit --ai-provider zhipu --api-key <key> --model GLM-4-Flash-250414 --endpoint https://open.bigmodel.cn/api/paas/v4/chat/completions
```

---

## 📋 环境变量

支持通过环境变量自动检测 API Key：

| 环境变量 | 对应提供商 |
|---------|-----------|
| `ZHIPU_API_KEY` | 智谱 GLM |
| `BAIDU_API_KEY` | 百度千帆 |
| `SILICONFLOW_API_KEY` | SiliconFlow |
| `MODELSCOPE_API_KEY` | ModelScope |
| `DEEPSEEK_API_KEY` | DeepSeek |
| `TENCENT_API_KEY` | 腾讯 |
| `VOLCENGINE_API_KEY` | 火山引擎 |
| `QWEN_API_KEY` | 阿里 Qwen |

---

## 🔑 免费 API Key 获取指引

### 智谱 GLM

1. 访问 [智谱 AI 开放平台](https://open.bigmodel.cn/)
2. 注册并登录账号
3. 进入「API 密钥」页面
4. 点击「创建 API 密钥」获取免费额度

### 百度千帆

1. 访问 [百度智能云](https://cloud.baidu.com/)
2. 注册并登录账号
3. 搜索「千帆大模型平台」
4. 进入控制台，在「API 密钥管理」中获取免费 API Key

### SiliconFlow

1. 访问 [SiliconFlow](https://siliconflow.cn/)
2. 注册并登录账号
3. 进入「API 密钥」设置
4. 创建新的 API 密钥，新用户可获得免费调用额度

### ModelScope

1. 访问 [ModelScope 平台](https://www.modelscope.cn/)
2. 注册并登录账号
3. 进入个人中心的「API 密钥」页面
4. 生成并复制 API Key，新用户有免费额度

### 阿里 Qwen

1. 访问 [阿里云 DashScope](https://dashscope.aliyun.com/)
2. 注册并登录阿里云账号
3. 进入「API 密钥」管理
4. 生成新的 API Key，新用户可获得免费调用次数

### DeepSeek

1. 访问 [DeepSeek 开放平台](https://platform.deepseek.com/)
2. 注册并登录账号
3. 进入「API 密钥」设置
4. 创建新的 API 密钥，新用户有免费使用额度

### 腾讯

1. 访问 [腾讯云 AI 平台](https://cloud.tencent.com/product/ai)
2. 注册并登录腾讯云账号
3. 搜索「智能对话」或「大语言模型」服务
4. 进入控制台，在「API 密钥管理」中获取免费 API Key

### 火山引擎

1. 访问 [火山引擎官网](https://www.volcengine.com/)
2. 注册并登录火山引擎账号
3. 搜索「大语言模型」或「智能对话」服务
4. 进入控制台，在「API 密钥管理」中获取免费 API Key

---

## ☁️ 支持的 AI 提供商

| 提供商 | 默认端点 | 默认模型 |
|-------|---------|---------|
| `zhipu` | `https://open.bigmodel.cn/api/paas/v4/chat/completions` | GLM-4-Flash-250414 |
| `baidu` | `https://qianfan.baidubce.com/v2/chat/completions` | ernie-4.5-0.3b |
| `siliconflow` | `https://api.siliconflow.cn/v1/chat/completions` | THUDM/GLM-4-9B-0414 |
| `modelscope` | `https://api-inference.modelscope.cn/v1/chat/completions` | deepseek-ai/DeepSeek-V4-Pro |
| `qwen` | `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions` | qwen-coder-plus-latest |
| `deepseek` | `https://api.deepseek.com/chat/completions` | deepseek-chat |
| `tencent` | `https://api.tencentcloud.com` | hunyuan-lite |
| `volcengine` | `https://ark.cn-beijing.volces.com/api/v3/chat/completions` | doubao-1-5-lite-32k-250115 |

---

## ⚙️ 工作原理

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  git diff   │───▶│  智能过滤   │───▶│  压缩摘要   │───▶│  AI 生成   │───▶│  输出结果   │
│ --cached   │    │  锁/二进制  │    │  diff      │    │  提交信息   │    │  CLI/Hook  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

1. **获取变更**：通过 `git diff --cached` 获取暂存区的代码变更
2. **智能过滤**：自动过滤锁文件、二进制文件、可执行文件等
3. **压缩摘要**：对 diff 进行压缩，保留关键结构信息
4. **生成提交**：调用 AI 模型生成符合规范的提交信息
5. **输出结果**：CLI 模式直接输出，Hook 模式写入提交信息文件

---

## 🔍 过滤规则

以下文件类型会被自动过滤，**不会**发送给 AI：

| 类型 | 文件/目录 |
|------|----------|
| 依赖锁文件 | `go.mod`, `go.sum`, `package-lock.json`, `pnpm-lock.yaml` |
| 压缩资源 | `dist/`, `build/`, `node_modules/`, `vendor/` |
| 二进制文件 | 图片、字体、视频、压缩包等 |
| 文档文件 | `.csv`, `.txt`, `.log` |

---

## 📖 命令行参数

| 参数 | 简写 | 说明 |
|-----|-----|------|
| `--api-key` | `-k` | API Key |
| `--ai-provider` | `-p` | AI 提供商 |
| `--endpoint` | `-e` | API 端点 |
| `--model` | `-m` | 模型名称 |
| `--zhipu-api-key` | - | 智谱 API Key |
| `--qwen-api-key` | - | 阿里 Qwen API Key |
| `--baidu-api-key` | - | 百度 API Key |
| `--siliconflow-api-key` | - | SiliconFlow API Key |
| `--modelscope-api-key` | - | ModelScope API Key |
| `--tencent-api-key` | - | 腾讯 API Key |
| `--volcengine-api-key` | - | 火山引擎 API Key |

---

## 💖 支持作者

<div align="center">

如果您觉得这个工具对您有帮助，可以通过以下方式支持作者：

![微信支付](wechat.png) ![支付宝](aliyun.png)

</div>

---

## 📬 定制需求

如果您有定制需求，可以联系作者获取更多支持。
