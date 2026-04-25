package main

import (
	"fmt"
	"os"
	"strings"
)

// 解析命令行参数
var (
	ApiKey     string
	Endpoint   string
	Model      string
	AiProvider string
)

// checkRequiredParams 检查必要的参数是否设置
// 返回:
//   - 是否所有必要参数都已设置
//   - 错误信息
func checkRequiredParams() (bool, error) {
	if AiProvider == "" {
		return false, fmt.Errorf("错误: AI_PROVIDER 是必需的")
	}
	if ApiKey == "" {
		return false, fmt.Errorf("错误: API_KEY 是必需的")
	}
	if Endpoint == "" {
		return false, fmt.Errorf("错误: ENDPOINT 是必需的")
	}
	if Model == "" {
		return false, fmt.Errorf("错误: MODEL 是必需的")
	}
	return true, nil
}

// printError 打印错误信息并退出
// 参数:
//   - err: 错误信息
func printError(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

// validateParams 验证参数格式
// 返回:
//   - 是否所有参数格式正确
//   - 错误信息
func validateParams() (bool, error) {
	// 验证 API Key
	if ApiKey != "" && len(ApiKey) < 10 {
		return false, fmt.Errorf("错误: API Key 格式无效")
	}

	// 验证端点 URL
	if Endpoint != "" {
		if !strings.HasPrefix(Endpoint, "https://") {
			return false, fmt.Errorf("错误: 端点 URL 必须使用 HTTPS 协议")
		}
	}

	// 验证模型名称
	if Model != "" && len(Model) < 3 {
		return false, fmt.Errorf("错误: 模型名称格式无效")
	}

	return true, nil
}
