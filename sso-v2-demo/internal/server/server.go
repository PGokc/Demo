package server

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ops_local_demo/internal/config"
	"ops_local_demo/internal/session"
)

type Server struct {
	cfg      config.Config
	client   *http.Client
	sessions *session.Manager
	verifier *rsa.PublicKey
}

func New(cfg config.Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		sessions: session.NewManager(cfg.SessionSecret),
	}

	if cfg.DeliveryPublicKeyPath != "" {
		pub, err := loadRSAPublicKey(cfg.DeliveryPublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load delivery public key: %w", err)
		}
		s.verifier = pub
		log.Printf("step=load_public_key path=%s result=ok", cfg.DeliveryPublicKeyPath)
	} else {
		log.Printf("step=load_public_key result=skipped reason=DELIVERY_PUBLIC_KEY_PATH_not_set")
	}

	if cfg.KA == "" {
		log.Printf("step=config_warning field=KA message=KA environment variable is not set; /login will fail")
	}
	if cfg.SessionSecret == "" {
		log.Printf("step=config_warning field=SESSION_SECRET message=session cookies will not be created until SESSION_SECRET is set")
	}
	if cfg.OpsHost != "" {
		log.Printf("step=config_info ops_host=%s", cfg.OpsHost)
	}

	return s, nil
}

// Routes 设置 HTTP 路由
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/login/sign/byted/callback", s.handleBytedCallback)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/health", s.handleHealth)
	return mux
}

// handleIndex 处理根路径请求，检查会话状态并渲染页面
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 检查会话 Cookie 是否存在
	data, err := s.sessions.Get(r)
	if err != nil {
		if !errors.Is(err, session.ErrNoSession) && !errors.Is(err, session.ErrInvalidSession) {
			log.Printf("step=index_get_session result=error error=%v", err)
		}
		// 会话不存在或无效，渲染未登录页面
		s.renderIndexLoggedOut(w, r)
		return
	}

	// 会话存在且有效，渲染已登录页面
	s.renderIndexLoggedIn(w, r, data)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.cfg.KA == "" {
		s.renderError(w, r, http.StatusInternalServerError, "KA 未配置", "请先设置 KA 环境变量")
		return
	}

	proxyURL := s.cfg.DeliveryBase + "/api/sign/proxy"
	u, err := url.Parse(proxyURL)
	if err != nil {
		s.renderError(w, r, http.StatusInternalServerError, "内部错误", "无法解析 Delivery 代理地址")
		return
	}

	q := u.Query()
	q.Set("ka", s.cfg.KA)
	q.Set("redirect_uri", s.cfg.RedirectPath)
	u.RawQuery = q.Encode()

	log.Printf("step=redirect_to_delivery ka=%s redirect_uri=%s proxy_url=%s", s.cfg.KA, s.cfg.RedirectPath, u.String())
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (s *Server) handleBytedCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	code := q.Get("code")
	ka := q.Get("ka")
	redirectURI := q.Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = s.cfg.RedirectPath
	}
	if !strings.HasPrefix(redirectURI, "/") {
		redirectURI = "/" + redirectURI
	}

	log.Printf("step=callback_received path=%s code=%s ka=%s redirect_uri=%s", r.URL.Path, code, ka, redirectURI)

	if code == "" {
		s.renderError(w, r, http.StatusBadRequest, "缺少 code 参数", "delivery 回调中未携带 byted code，无法继续登录")
		return
	}

	userURL := s.cfg.DeliveryBase + "/api/sign/byted/user"
	u, err := url.Parse(userURL)
	if err != nil {
		s.renderError(w, r, http.StatusInternalServerError, "内部错误", "无法解析 Delivery 用户信息地址")
		return
	}

	uq := u.Query()
	uq.Set("code", code)
	u.RawQuery = uq.Encode()

	log.Printf("step=call_delivery_user_api url=%s", u.String())

	resp, err := s.client.Get(u.String())
	if err != nil {
		log.Printf("step=call_delivery_user_api result=error error=%v", err)
		s.renderError(w, r, http.StatusBadGateway, "调用 Delivery 失败", "请检查网络连通性和 KA 配置")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("step=read_delivery_response result=error error=%v", err)
		s.renderError(w, r, http.StatusBadGateway, "读取 Delivery 响应失败", "")
		return
	}

	log.Printf("step=delivery_user_api_response status=%d body_len=%d", resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		log.Printf("step=delivery_user_api_response status=%d body=%s", resp.StatusCode, string(body))
		s.renderError(w, r, http.StatusBadGateway, "Delivery 返回错误", fmt.Sprintf("status=%d", resp.StatusCode))
		return
	}

	content, signVal, err := extractContentAndSign(body)
	if err != nil {
		log.Printf("step=parse_content_sign result=error error=%v body=%s", err, string(body))
		s.renderError(w, r, http.StatusBadGateway, "解析 Delivery 响应失败", err.Error())
		return
	}

	log.Printf("step=parse_content_sign result=ok content_len=%d", len(content))

	if s.verifier != nil {
		if err := verifySignature(s.verifier, content, signVal); err != nil {
			log.Printf("step=verify_signature result=failed error=%v", err)
			s.renderError(w, r, http.StatusInternalServerError, "验签失败", "请检查 DELIVERY_PUBLIC_KEY_PATH 配置以及签名算法")
			return
		}
		log.Printf("step=verify_signature result=ok")
	} else {
		log.Printf("step=verify_signature result=skipped reason=no_public_key_configured")
	}

	userJSON, decoded := decodeUserJSON(content)
	log.Printf("step=decode_user_content decoded=%t user_json_len=%d", decoded, len(userJSON))

	if s.cfg.SessionSecret == "" {
		log.Printf("step=set_session result=skipped reason=session_secret_missing")
	} else {
		if err := s.sessions.Create(w, r, userJSON); err != nil {
			log.Printf("step=set_session result=error error=%v", err)
			s.renderError(w, r, http.StatusInternalServerError, "写入本地会话失败", "")
			return
		}
		log.Printf("step=set_session result=ok")
	}

	log.Printf("step=redirect_to_final redirect_uri=%s", redirectURI)
	http.Redirect(w, r, redirectURI, http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.sessions.Clear(w, r)
	log.Printf("step=logout_cleared_session path=%s", r.URL.Path)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) renderIndexLoggedOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, "<html><head><title>ops_local_demo</title></head><body>")
	_, _ = fmt.Fprintf(w, "<h1>ops_local_demo</h1>")
	_, _ = fmt.Fprintf(w, "<p>当前未登录。</p>")
	_, _ = fmt.Fprintf(w, "<p><a href=\"/login\">使用 Delivery 字节登录</a></p>")
	_, _ = fmt.Fprintf(w, "<hr>")
	_, _ = fmt.Fprintf(w, "<p>Host: %s</p>", html.EscapeString(r.Host))
	_, _ = fmt.Fprintf(w, "<p>KA: %s</p>", html.EscapeString(s.cfg.KA))
	_, _ = fmt.Fprintf(w, "<p>Delivery base: %s</p>", html.EscapeString(s.cfg.DeliveryBase))
	_, _ = fmt.Fprintf(w, "<p>Redirect path: %s</p>", html.EscapeString(s.cfg.RedirectPath))
	_, _ = fmt.Fprintf(w, "</body></html>")
}

func (s *Server) renderIndexLoggedIn(w http.ResponseWriter, r *http.Request, data *session.Data) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, "<html><head><title>ops_local_demo</title></head><body>")
	_, _ = fmt.Fprintf(w, "<h1>ops_local_demo</h1>")
	_, _ = fmt.Fprintf(w, "<p>已登录。</p>")

	pretty := data.UserJSON
	if data.UserJSON != "" {
		var obj interface{}
		if err := json.Unmarshal([]byte(data.UserJSON), &obj); err == nil {
			if b, err := json.MarshalIndent(obj, "", "  "); err == nil {
				pretty = string(b)
			}
		}
	}

	_, _ = fmt.Fprintf(w, "<h2>用户信息 (来自 delivery content)</h2>")
	_, _ = fmt.Fprintf(w, "<pre>%s</pre>", html.EscapeString(pretty))
	_, _ = fmt.Fprintf(w, "<p><a href=\"/logout\">退出登录</a></p>")
	_, _ = fmt.Fprintf(w, "</body></html>")
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, status int, msg, detail string) {
	log.Printf("step=render_error status=%d msg=%s detail=%s path=%s", status, msg, detail, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "<html><head><title>Error</title></head><body>")
	_, _ = fmt.Fprintf(w, "<h1>Error</h1>")
	_, _ = fmt.Fprintf(w, "<p>Status: %d</p>", status)
	_, _ = fmt.Fprintf(w, "<p>%s</p>", html.EscapeString(msg))
	if detail != "" {
		_, _ = fmt.Fprintf(w, "<p><small>%s</small></p>", html.EscapeString(detail))
	}
	_, _ = fmt.Fprintf(w, "<p><a href=\"/\">返回首页</a></p>")
	_, _ = fmt.Fprintf(w, "</body></html>")
}

func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block found in public key file")
	}

	// Try PKIX first
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		if pub, ok := pubAny.(*rsa.PublicKey); ok {
			return pub, nil
		}
		return nil, errors.New("public key is not RSA")
	}

	// Fallback: maybe it's a certificate containing the public key
	if cert, err2 := x509.ParseCertificate(block.Bytes); err2 == nil {
		if pub, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return pub, nil
		}
	}

	return nil, err
}

func extractContentAndSign(body []byte) (string, string, error) {
	var top map[string]interface{}
	if err := json.Unmarshal(body, &top); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	getFromMap := func(m map[string]interface{}) (string, string, bool) {
		c, okC := m["content"]
		s, okS := m["sign"]
		if !okC || !okS {
			return "", "", false
		}
		cs, okC := c.(string)
		ss, okS := s.(string)
		if !okC || !okS {
			return "", "", false
		}
		return cs, ss, true
	}

	if c, s, ok := getFromMap(top); ok {
		return c, s, nil
	}

	if dataVal, ok := top["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if c, s, ok := getFromMap(dataMap); ok {
				return c, s, nil
			}
		}
	}

	return "", "", errors.New("could not find content/sign fields in response")
}

func decodeUserJSON(content string) (string, bool) {
	encoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}

	for _, enc := range encoders {
		b, err := enc.DecodeString(content)
		if err == nil {
			return string(b), true
		}
	}

	// Fallback: treat content itself as JSON
	return content, false
}

func verifySignature(pub *rsa.PublicKey, content, signatureB64 string) error {
	if signatureB64 == "" {
		return errors.New("empty signature")
	}

	decoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}

	var sig []byte
	var lastErr error
	for _, dec := range decoders {
		b, err := dec.DecodeString(signatureB64)
		if err == nil {
			sig = b
			lastErr = nil
			break
		}
		lastErr = err
	}
	if lastErr != nil {
		return fmt.Errorf("decode signature: %w", lastErr)
	}

	h := sha256.New()
	h.Write([]byte(content))
	digest := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest, sig); err != nil {
		return err
	}
	return nil
}
