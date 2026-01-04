package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNoSession      = errors.New("no session")
	ErrInvalidSession = errors.New("invalid session")
)

const cookieName = "ops_local_session"

// Manager 会话管理器，负责创建、验证和清除会话。
type Manager struct {
	secret     []byte
	cookieName string
	ttl        time.Duration
}

// Data 会话数据结构，存储在 Cookie 中的实际内容。
type Data struct {
	UserJSON  string `json:"user_json"`
	CreatedAt int64  `json:"created_at"`
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		cookieName: cookieName,
		ttl:        24 * time.Hour,
	}
}

// Create 创建会话并设置 Cookie。
func (m *Manager) Create(w http.ResponseWriter, r *http.Request, userJSON string) error {
	// 验证密钥是否存在
	if len(m.secret) == 0 {
		return errors.New("session secret is empty")
	}

	// 构建会话数据（包含用户信息和创建时间）
	data := Data{
		UserJSON:  userJSON,
		CreatedAt: time.Now().Unix(),
	}

	// 序列化并编码数据
	payloadBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// 使用 HMAC-SHA256 生成签名以防止数据篡改
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	sig := mac.Sum(nil)

	// 设置 HTTP Cookie，包含编码后的会话数据和签名
	cookieVal := payload + "." + hex.EncodeToString(sig)
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    cookieVal,
		Path:     "/",
		Domain:   cookieDomain(r.Host),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(m.ttl),
	}

	http.SetCookie(w, cookie)
	return nil
}

func (m *Manager) Get(r *http.Request) (*Data, error) {
	// 从请求中获取会话 Cookie
	c, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, ErrNoSession
		}
		return nil, err
	}

	// 验证 Cookie 值格式是否正确, 解析 Cookie 值（分离数据和签名）
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidSession
	}
	payload, sigHex := parts[0], parts[1]

	// 验证签名是否有效
	expectedSig, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, ErrInvalidSession
	}

	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	if !hmac.Equal(mac.Sum(nil), expectedSig) {
		return nil, ErrInvalidSession
	}

	// 解码并反序列化会话数据
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrInvalidSession
	}

	var data Data
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return nil, ErrInvalidSession
	}

	// 检查会话是否过期
	if m.ttl > 0 {
		if time.Now().Unix()-data.CreatedAt > int64(m.ttl.Seconds()) {
			return nil, ErrInvalidSession
		}
	}

	return &data, nil
}

// Clear 清除会话 Cookie，创建一个过期的 Cookie 覆盖原有 Cookie
func (m *Manager) Clear(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain(r.Host),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(w, cookie)
}

// cookieDomain 从主机名中提取 Cookie 域名（移除端口）
func cookieDomain(host string) string {
	if host == "" {
		return ""
	}

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}
