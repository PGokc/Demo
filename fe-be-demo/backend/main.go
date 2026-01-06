package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// APIResponse 定义统一的返回结构
// code: 0 表示成功，非 0 表示错误
// message: 提示信息
// data: 具体业务数据
//
// 示例：
// {"code":0,"message":"ok","data":{"status":"ok"}}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// CreateTodoRequest 用于解析创建 todo 的请求体

type CreateTodoRequest struct {
	Title string `json:"title"`
	Done  *bool  `json:"done,omitempty"`
}

var (
	todos  []Todo
	nextID = 1
	mu     sync.Mutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/echo", echoHandler)
	mux.HandleFunc("/api/todos", todosHandler)

	handler := withLogging(withCORS(mux))

	addr := ":8080"
	log.Printf("HTTP server starting on %s...", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// withLogging 简单日志中间件

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("%s %s done in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// withCORS 提供基础 CORS 支持
// 允许：http://localhost:5173 和本地 file:// 页面（Origin 通常为 "null" 或空）

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		switch origin {
		case "http://localhost:5173", "http://localhost:63342":
			w.Header().Set("Access-Control-Allow-Origin", origin)
		case "null":
			// 一般代表本地 file:// 页面
			w.Header().Set("Access-Control-Allow-Origin", "*")
		case "":
			// 无 Origin：可能是同源请求或 CLI 工具，不强制设置
		default:
			// 其他域名默认不允许跨域
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			// 预检请求直接返回
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthHandler: GET /api/health
// 返回服务健康状态

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r, []string{http.MethodGet})
		return
	}

	resp := APIResponse{
		Code:    0,
		Message: "ok",
		Data: map[string]string{
			"status": "ok",
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

// echoHandler: POST /api/echo
// 原样返回请求 JSON，并附加服务器时间戳

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, r, []string{http.MethodPost})
		return
	}

	if r.Header.Get("Content-Type") == "" {
		// 允许未显式声明 Content-Type，但推荐 application/json
		log.Printf("warning: /api/echo request without Content-Type header")
	}

	var payload map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("/api/echo decode error: %v", err)
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "invalid JSON body",
			Data: map[string]string{
				"detail": err.Error(),
			},
		})
		return
	}

	payload["server_time"] = time.Now().Format(time.RFC3339)

	resp := APIResponse{
		Code:    0,
		Message: "ok",
		Data:    payload,
	}
	writeJSON(w, http.StatusOK, resp)
}

// todosHandler: GET/POST /api/todos
// GET: 返回 todo 列表
// POST: 创建新的 todo

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTodos(w, r)
	case http.MethodPost:
		handleCreateTodo(w, r)
	default:
		methodNotAllowed(w, r, []string{http.MethodGet, http.MethodPost})
	}
}

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// 为避免并发读写问题，可复制一份
	list := make([]Todo, len(todos))
	copy(list, todos)

	resp := APIResponse{
		Code:    0,
		Message: "ok",
		Data:    list,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "" {
		log.Printf("warning: /api/todos POST without Content-Type header")
	}

	var req CreateTodoRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("/api/todos decode error: %v", err)
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "invalid JSON body",
			Data: map[string]string{
				"detail": err.Error(),
			},
		})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "title is required",
		})
		return
	}

	done := false
	if req.Done != nil {
		done = *req.Done
	}

	mu.Lock()
	todo := Todo{
		ID:    nextID,
		Title: req.Title,
		Done:  done,
	}
	nextID++
	todos = append(todos, todo)
	mu.Unlock()

	resp := APIResponse{
		Code:    0,
		Message: "ok",
		Data:    todo,
	}
	writeJSON(w, http.StatusOK, resp)
}

// methodNotAllowed 统一处理不支持的方法

func methodNotAllowed(w http.ResponseWriter, r *http.Request, allowed []string) {
	w.Header().Set("Allow", joinMethods(allowed))
	writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
		Code:    405,
		Message: "method not allowed",
		Data: map[string]interface{}{
			"allowed": allowed,
		},
	})
}

func joinMethods(methods []string) string {
	if len(methods) == 0 {
		return ""
	}
	if len(methods) == 1 {
		return methods[0]
	}

	// 简单拼接，避免引入 strings 包（也可以使用 strings.Join）
	s := methods[0]
	for i := 1; i < len(methods); i++ {
		s += ", " + methods[i]
	}
	return s
}

// writeJSON 统一输出 JSON 响应

func writeJSON(w http.ResponseWriter, statusCode int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
