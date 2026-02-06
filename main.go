// main.go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Backend struct {
	Name         string   `yaml:"name"`
	URL          string   `yaml:"url"`
	APIKey       string   `yaml:"api_key"`
	Weight       int      `yaml:"weight"`
	DefaultModel string   `yaml:"default_model"`
	Models       []string `yaml:"models"`
}

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Timeout  time.Duration `yaml:"timeout"`
	Retry    int           `yaml:"retry"`
	Mode     string        `yaml:"mode"`
	Backends []Backend     `yaml:"backends"`
}

type ChatRequest struct {
	Model    string          `json:"model"`
	Messages json.RawMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

var (
	config        Config
	roundRobinIdx int
	rrMutex       sync.Mutex
)

func main() {
	// 支持 -c 参数指定配置文件
	configFile := "config.yaml"
	for i, arg := range os.Args {
		if arg == "-c" && i+1 < len(os.Args) {
			configFile = os.Args[i+1]
		}
	}

	loadConfig(configFile)

	http.HandleFunc("/v1/chat/completions", handleChat)
	http.HandleFunc("/v1/models", handleModels)
	http.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("🚀 代理启动在 %s (模式: %s, 后端数: %d)", addr, config.Mode, len(config.Backends))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Timeout == 0 {
		config.Timeout = 180 * time.Second
	}
	if config.Retry == 0 {
		config.Retry = 3
	}
	if config.Mode == "" {
		config.Mode = "random"
	}

	// 检查后端配置
	if len(config.Backends) == 0 {
		log.Fatalf("错误: 没有配置任何后端")
	}

	// 打印后端信息
	for i, b := range config.Backends {
		log.Printf("后端 %d: %s (%s) 模型: %v", i+1, b.Name, b.URL, b.Models)
	}
}

// 根据模型名找后端，找不到就随机选一个
func findBackend(model string) (*Backend, string) {
	if len(config.Backends) == 0 {
		return nil, ""
	}

	// 先精确匹配
	for i := range config.Backends {
		b := &config.Backends[i]
		for _, m := range b.Models {
			if m == model {
				return b, model
			}
		}
	}

	// 找不到，随机选一个后端，用它的默认模型
	idx := rand.Intn(len(config.Backends))
	b := &config.Backends[idx]
	useModel := b.DefaultModel
	if useModel == "" && len(b.Models) > 0 {
		useModel = b.Models[0]
	}
	log.Printf("模型 %s 未找到，随机使用 %s/%s", model, b.Name, useModel)
	return b, useModel
}

// 获取下一个后端（轮询或随机）
func nextBackend() *Backend {
	if len(config.Backends) == 0 {
		return nil
	}

	if config.Mode == "round-robin" {
		rrMutex.Lock()
		idx := roundRobinIdx % len(config.Backends)
		roundRobinIdx++
		rrMutex.Unlock()
		return &config.Backends[idx]
	}

	// 加权随机
	totalWeight := 0
	for _, b := range config.Backends {
		w := b.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	r := rand.Intn(totalWeight)
	for i := range config.Backends {
		w := config.Backends[i].Weight
		if w <= 0 {
			w = 1
		}
		r -= w
		if r < 0 {
			return &config.Backends[i]
		}
	}
	return &config.Backends[0]
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// 处理 CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查后端
	if len(config.Backends) == 0 {
		http.Error(w, "没有可用的后端", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求失败", http.StatusBadRequest)
		return
	}

	var req ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 选后端和模型
	var backend *Backend
	var useModel string

	if req.Model == "" || req.Model == "auto" {
		backend = nextBackend()
		if backend == nil {
			http.Error(w, "没有可用的后端", http.StatusServiceUnavailable)
			return
		}
		useModel = backend.DefaultModel
		if useModel == "" && len(backend.Models) > 0 {
			useModel = backend.Models[0]
		}
	} else {
		backend, useModel = findBackend(req.Model)
		if backend == nil {
			http.Error(w, "没有可用的后端", http.StatusServiceUnavailable)
			return
		}
	}

	log.Printf("请求: model=%s -> %s/%s stream=%v", req.Model, backend.Name, useModel, req.Stream)

	// 构造请求体，保留原始请求中的其他字段
	var reqMap map[string]interface{}
	json.Unmarshal(body, &reqMap)
	reqMap["model"] = useModel
	reqData, _ := json.Marshal(reqMap)

	// 重试逻辑
	var lastErr error
	tried := make(map[string]bool)

	for i := 0; i < config.Retry; i++ {
		if i > 0 {
			// 换一个后端重试
			for j := 0; j < len(config.Backends); j++ {
				b := nextBackend()
				if b != nil && !tried[b.Name] {
					backend = b
					useModel = b.DefaultModel
					if useModel == "" && len(b.Models) > 0 {
						useModel = b.Models[0]
					}
					reqMap["model"] = useModel
					reqData, _ = json.Marshal(reqMap)
					break
				}
			}
			log.Printf("重试 %d: %s/%s", i, backend.Name, useModel)
		}
		tried[backend.Name] = true

		err := doRequest(w, backend, reqData, req.Stream)
		if err == nil {
			return
		}
		lastErr = err
		log.Printf("后端 %s 失败: %v", backend.Name, err)
	}

	http.Error(w, fmt.Sprintf("所有后端失败: %v", lastErr), http.StatusBadGateway)
}

func doRequest(w http.ResponseWriter, backend *Backend, reqData []byte, stream bool) error {
	url := strings.TrimSuffix(backend.URL, "/") + "/chat/completions"

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+backend.APIKey)

	client := &http.Client{Timeout: config.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 流式响应
	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("不支持流式响应")
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "%s\n\n", line)
			flusher.Flush()
		}
		return scanner.Err()
	}

	// 非流式
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, resp.Body)
	return err
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var models []map[string]interface{}
	seen := make(map[string]bool)

	for _, b := range config.Backends {
		for _, m := range b.Models {
			if !seen[m] {
				seen[m] = true
				models = append(models, map[string]interface{}{
					"id":       m,
					"object":   "model",
					"owned_by": b.Name,
				})
			}
		}
	}

	resp := map[string]interface{}{
		"object": "list",
		"data":   models,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"backends": len(config.Backends),
		"mode":     config.Mode,
	})
}
