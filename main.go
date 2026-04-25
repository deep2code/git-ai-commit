package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// main 程序主函数
// 功能：
// 1. 检测运行模式（CLI模式或Git Hook模式）
// 2. 解析命令行参数
// 3. 自动检测环境变量中的API Key
// 4. 设置默认模型和端点
// 5. 验证必要参数
// 6. 获取Git暂存区的变更
// 7. 调用AI生成提交信息
// 8. 输出结果
func main() {
	isHookMode := false
	msgFilePath := ""

	// 定义命令行参数
	var apiKey string
	var aiProvider string
	var zhipuAPIKey string
	var qwenAPIKey string
	var baiduAPIKey string
	var siliconflowAPIKey string
	var modelscopeAPIKey string
	var deepseekAPIKey string
	var tencentAPIKey string
	var volcengineAPIKey string
	var endpoint string
	var model string

	// 长选项
	flag.StringVar(&apiKey, "api-key", "", "API Key")
	flag.StringVar(&aiProvider, "ai-provider", "", "AI Provider")
	flag.StringVar(&zhipuAPIKey, "zhipu-api-key", "", "Zhipu API Key")
	flag.StringVar(&qwenAPIKey, "qwen-api-key", "", "Qwen API Key")
	flag.StringVar(&baiduAPIKey, "baidu-api-key", "", "Baidu API Key")
	flag.StringVar(&siliconflowAPIKey, "siliconflow-api-key", "", "SiliconFlow API Key")
	flag.StringVar(&modelscopeAPIKey, "modelscope-api-key", "", "ModelScope API Key")
	flag.StringVar(&deepseekAPIKey, "deepseek-api-key", "", "DeepSeek API Key")
	flag.StringVar(&tencentAPIKey, "tencent-api-key", "", "Tencent API Key")
	flag.StringVar(&volcengineAPIKey, "volcengine-api-key", "", "Volcengine API Key")
	flag.StringVar(&endpoint, "endpoint", "", "API Endpoint")
	flag.StringVar(&model, "model", "", "Model Name")

	// 短选项
	flag.StringVar(&apiKey, "k", "", "API Key (short)")
	flag.StringVar(&aiProvider, "p", "", "AI Provider (short)")
	flag.StringVar(&endpoint, "e", "", "API Endpoint (short)")
	flag.StringVar(&model, "m", "", "Model Name (short)")

	if len(os.Args) > 1 {
		msgFilePath = os.Args[1]
		if msgFilePath == ".git/COMMIT_EDITMSG" {
			isHookMode = true
			if len(os.Args) > 2 {
				source := os.Args[2]
				if source == "merge" || source == "squash" {
					// 合并或压缩提交时不需要生成提交信息
					os.Exit(0)
				}
			}
		} else {
			// CLI 模式：解析命令行参数
			flag.Parse()
		}
	} else {
		// CLI 模式：解析命令行参数
		flag.Parse()
	}

	// 设置参数值
	if apiKey != "" {
		ApiKey = apiKey
	}
	if aiProvider != "" {
		AiProvider = aiProvider
	}
	if endpoint != "" {
		Endpoint = endpoint
	}
	if model != "" {
		Model = model
	}

	if AiProvider == "" || ApiKey == "" {
		// 处理特定提供商的 API Key
		providers := []struct {
			name     string
			flagName string
		}{
			{"zhipu", "zhipu-api-key"},
			{"tencent", "tencent-api-key"},
			{"baidu", "baidu-api-key"},
			{"siliconflow", "siliconflow-api-key"},
			{"modelscope", "modelscope-api-key"},
			{"qwen", "qwen-api-key"},
			{"deepseek", "deepseek-api-key"},
			{"volcengine", "volcengine-api-key"},
		}

		for _, provider := range providers {
			flagValue := flag.Lookup(provider.flagName)
			if flagValue != nil && flagValue.Value.String() != "" {
				ApiKey = flagValue.Value.String()
				AiProvider = provider.name
				break
			}
		}
	}

	if AiProvider == "" || ApiKey == "" {
		detectByEnv()
	}

	if Endpoint == "" || Model == "" {
		defaultModel()
	}

	// 验证参数格式
	if ok, err := validateParams(); !ok {
		printError(err)
	}

	// 验证必要参数
	if ok, err := checkRequiredParams(); !ok {
		printError(err)
	}

	// 1. 获取 Diff
	diffStr, err := getStagedChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取暂存区变更失败: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(diffStr) == "" {
		fmt.Println("未检测到暂存区变更")
		os.Exit(0)
	}

	// 2. 生成 Commit Message
	commitMsg, err := generateCommitMessageByProvider(diffStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI生成失败 (%s): %v\n", AiProvider, err)
		os.Exit(1)
	}

	// 3. 输出结果
	if isHookMode {
		err = os.WriteFile(msgFilePath, []byte(commitMsg), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "写入提交信息文件失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(commitMsg)
	}
}
