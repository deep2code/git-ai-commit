package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ================= 数据结构 =================

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`  // 新增：限制最大 token 数
	Temperature float32   `json:"temperature,omitempty"` // 新增：降低随机性，更听话
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ================= 多模型适配逻辑 =================

func detectByEnv() {
	// 定义提供商和环境变量的映射，按优先级顺序排列
	providers := []struct {
		name       string
		envVarName string
	}{
		{"zhipu", "ZHIPU_API_KEY"},
		{"baidu", "BAIDU_API_KEY"},
		{"siliconflow", "SILICONFLOW_API_KEY"},
		{"modelscope", "MODELSCOPE_API_KEY"},
		{"qwen", "QWEN_API_KEY"},
		{"deepseek", "DEEPSEEK_API_KEY"},
		{"tencent", "TENCENT_API_KEY"},
		{"volcengine", "VOLCENGINE_API_KEY"},
	}

	// 遍历提供商，检查环境变量
	for _, provider := range providers {
		apiKey := os.Getenv(provider.envVarName)
		if apiKey != "" {
			if AiProvider == "" {
				AiProvider = provider.name
			}
			if ApiKey == "" {
				ApiKey = apiKey
			}
			return
		}
	}
}

func defaultModel() {
	switch AiProvider {
	case "tencent":
		if Endpoint == "" {
			Endpoint = "https://api.hunyuan.cloud.tencent.com/v1/chat/completions"
		}
		if Model == "" {
			Model = "hunyuan-lite"
		}

	case "zhipu":
		if Endpoint == "" {
			Endpoint = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		}
		if Model == "" {
			Model = "GLM-4-Flash-250414"
			// 永久免费模型,但经常不能使用
			// Model = "glm-4-flash"
		}

	case "baidu":
		if Endpoint == "" {
			Endpoint = "https://qianfan.baidubce.com/v2/chat/completions"
		}
		if Model == "" {
			Model = "ernie-4.5-0.3b"
		}

	case "siliconflow":
		if Endpoint == "" {
			Endpoint = "https://api.siliconflow.cn/v1/chat/completions"
		}
		if Model == "" {
			Model = "THUDM/GLM-4-9B-0414"
		}

	case "modelscope":
		if Endpoint == "" {
			Endpoint = "https://api-inference.modelscope.cn/v1/chat/completions"
		}
		if Model == "" {
			Model = "deepseek-ai/DeepSeek-V4-Pro"
		}

	case "volcengine":
		if Endpoint == "" {
			Endpoint = "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
		}
		if Model == "" {
			Model = "doubao-1-5-lite-32k-250115"
		}

	case "qwen":
		if Endpoint == "" {
			Endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		}

		if Model == "" {
			Model = "qwen-coder-plus-latest"
		}

	case "deepseek":
		if Endpoint == "" {
			Endpoint = "https://api.deepseek.com/chat/completions"
		}
		if Model == "" {
			Model = "deepseek-chat"
		}
	}
}

// generateCommitMessageByProvider 根据AI提供商生成提交信息
// 参数:
//   - diff: Git变更的Diff内容
//
// 返回:
//   - 生成的提交信息
//   - 错误信息
func generateCommitMessageByProvider(diff string) (string, error) {
	if AiProvider == "zhipu" {
		token, err := generateZhipuToken(ApiKey)
		if err != nil {
			return "", fmt.Errorf("生成智谱认证令牌失败: %v", err)
		}
		return callOpenAICompatible(token, Endpoint, Model, diff)

	}

	return callOpenAICompatible(ApiKey, Endpoint, Model, diff)
}

// callOpenAICompatible 调用兼容OpenAI API格式的AI服务
// 参数:
//   - apiKey: API密钥
//   - endpoint: API端点URL
//   - model: 模型名称
//   - diff: Git变更的Diff内容
//
// 返回:
//   - 生成的提交信息
//   - 错误信息
func callOpenAICompatible(apiKey, endpoint, model, diff string) (string, error) {
	if apiKey == "" {
		return "", errors.New("API密钥不能为空")
	}
	if endpoint == "" {
		return "", errors.New("API端点不能为空")
	}
	if model == "" {
		return "", errors.New("模型名称不能为空")
	}
	if diff == "" {
		return "", errors.New("Diff内容不能为空")
	}

	prompt := buildPrompt(diff)

	// 构造请求，增加长度限制参数
	payload := ChatRequest{
		Model:       model,
		Messages:    []Message{{Role: "user", Content: prompt}},
		MaxTokens:   50,  // 强制硬限制：约等于 30-50 个汉字
		Temperature: 0.1, // 降低随机性，减少废话
	}
	return sendRequest(endpoint, apiKey, payload)
}

// sendRequest 发送HTTP请求到AI服务并处理响应
// 参数:
//   - endpoint: API端点URL
//   - bearerToken: 认证令牌
//   - payload: 请求体
//
// 返回:
//   - 生成的提交信息
//   - 错误信息
func sendRequest(endpoint, bearerToken string, payload interface{}) (string, error) {
	// 确保使用 HTTPS 连接
	if !strings.HasPrefix(endpoint, "https://") {
		return "", fmt.Errorf("错误: API 端点必须使用 HTTPS 协议")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	// 创建安全的 HTTP 客户端
	client := &http.Client{
		Timeout: time.Second * 30,
		// 只允许 TLS 1.2 及以上版本
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败,状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应数据失败: %v, 响应内容: %s", err, string(body))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API错误: %s (类型: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API未返回任何结果, 响应: %s", string(body))
	}

	if len(chatResp.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("API返回空内容, 响应: %s", string(body))
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// buildPrompt 构造极简短 Prompt
// 参数:
//   - diff: Git变更的Diff内容
//
// 返回:
//   - 构造好的Prompt字符串
func buildPrompt(diff string) string {
	return fmt.Sprintf(`分析以下代码变更，生成一条极简的 Commit Message。
核心要求：
1. 使用中文。
2. 必须遵循 Conventional Commits 格式 (例如: feat: ..., fix: ...)。
3. 【严格限制】输出内容总长度（包括冒号和空格）绝对不能超过 30 个汉字。
4. 只输出标题，不要输出正文，不要包含任何解释。
5. 宁可少字，不可多字。

变更内容:
%s`, diff)
}

// generateZhipuToken 生成智谱API的认证令牌
// 参数:
//   - apiKey: 智谱API Key
//
// 返回:
//   - 生成的认证令牌
//   - 错误信息
func generateZhipuToken(apiKey string) (string, error) {
	parts := strings.Split(apiKey, ".")
	if len(parts) != 2 {
		return "", errors.New("无效的智谱API Key格式，应为: id.secret")
	}
	id, secret := parts[0], parts[1]
	header := map[string]string{"alg": "HS256", "sign_type": "SIGN"}
	hJson, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("序列化header失败: %v", err)
	}
	hEnc := base64.RawURLEncoding.EncodeToString(hJson)
	now := time.Now().Unix()
	payload := map[string]interface{}{"api_key": id, "exp": now + 3600, "timestamp": now}
	pJson, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化payload失败: %v", err)
	}
	pEnc := base64.RawURLEncoding.EncodeToString(pJson)
	signStr := hEnc + "." + pEnc
	h := hmac.New(sha256.New, []byte(secret))
	_, err = h.Write([]byte(signStr))
	if err != nil {
		return "", fmt.Errorf("计算签名失败: %v", err)
	}
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return signStr + "." + signature, nil
}
