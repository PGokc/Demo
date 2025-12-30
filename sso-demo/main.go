package main

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	clientID     = "dd1lnuls0vaf8qyp7zi2"
	clientSecret = "9th9ia72fl24y0exoix2kddr19ispx3rsu7qwgjk"

	ssoDomain   = "test-sso.bytedance.net"
	authURL     = "https://" + ssoDomain + "/oauth2/authorize"
	tokenURL    = "https://" + ssoDomain + "/oauth2/access_token"
	userInfoURL = "https://" + ssoDomain + "/oauth2/userinfo"

	appURL          = "http://localhost:8080"
	redirectURI     = "http://localhost:8080/auth/callback"
	stateCookieName = "oauth_state"
	sessionCookie   = "app_session_id"
	serverListen    = ":8080"
)

// ====== 类型定义 ======

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

type UserInfoResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	NameZh   string `json:"name_zh"`
	NameEn   string `json:"name_en"`
	Avatar   string `json:"avatar"`
	Employee string `json:"employee_id"`
}

// 本地 Session 结构
type Session struct {
	User        UserInfoResponse
	AccessToken string
	ExpireAt    time.Time
}

var (
	sessionStore = make(map[string]Session)
	sessionMu    sync.RWMutex
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth/callback", handleCallback)
	http.HandleFunc("/me", handleMe)
	http.HandleFunc("/logout", handleLogout)

	log.Printf("Server listening on %s\n", serverListen)
	log.Fatal(http.ListenAndServe(serverListen, nil))
}

// ====== Handler ======

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 检查本地登录态
	sess, ok := getSessionFromRequest(r)
	if !ok {
		fmt.Fprintf(w, `
		<html>
		<head><title>SSO Demo</title></head>
		<body>
			<h1>ByteDance SSO Demo</h1>
			<p>当前未登录</p>
			<a href="/login">使用字节账号登录</a>
		</body>
		</html>
		`)
		return
	}

	fmt.Fprintf(w, `
	<html>
	<head><title>SSO Demo</title></head>
	<body>
		<h1>ByteDance SSO Demo</h1>
		<p>已登录用户：</p>
		<ul>
			<li>UserID: %s</li>
			<li>姓名(中文): %s</li>
			<li>姓名(英文): %s</li>
			<li>Email: %s</li>
			<li>EmployeeID: %s</li>
		</ul>
		<p><a href="/me">查看当前用户(JSON)</a></p>
		<p><a href="/logout">退出登录</a></p>
	</body>
	</html>
	`, sess.User.UserID, sess.User.NameZh, sess.User.NameEn, sess.User.Email, sess.User.Employee)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomString(32)

	// state 放 Cookie，用于回调时校验
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
	})

	u, err := url.Parse(authURL)
	if err != nil {
		http.Error(w, "failed to parse auth url", http.StatusInternalServerError)
		return
	}

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("access_type", "online")
	q.Set("scope", "read")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")

	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// 校验 state，防 CSRF
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value == "" || state != stateCookie.Value {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// 换 token
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("exchange token error: %v\n", err)
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// 获取用户信息
	userInfo, err := fetchUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("fetch user info error: %v\n", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// 创建本地 session
	sessionID := generateRandomString(40)
	sessionMu.Lock()
	sessionStore[sessionID] = Session{
		User:        userInfo,
		AccessToken: tokenResp.AccessToken,
		ExpireAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}
	sessionMu.Unlock()

	// 设置本地 session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})

	// 清掉 state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// 跳回首页
	http.Redirect(w, r, appURL, http.StatusFound)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	sess, ok := getSessionFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(sess)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// 删除本地 session
	if cookie, err := r.Cookie(sessionCookie); err == nil && cookie.Value != "" {
		sessionMu.Lock()
		delete(sessionStore, cookie.Value)
		sessionMu.Unlock()
	}
	// 清 Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, appURL, http.StatusFound)
}

// ====== SSO 相关调用 ======

func exchangeCodeForToken(code string) (*AccessTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 直接读取 resp.Body 所有内容
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			// 处理读取 body 失败的错误
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		// 将 []byte 转为字符串
		bodyStr := string(bodyBytes)
		return nil, fmt.Errorf("token endpoint status %d, body: %s", resp.StatusCode, bodyStr)
	}

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func fetchUserInfo(accessToken string) (UserInfoResponse, error) {
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return UserInfoResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return UserInfoResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 直接读取 resp.Body 所有内容
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			// 处理读取 body 失败的错误
			return UserInfoResponse{}, fmt.Errorf("failed to read response body: %w", err)
		}
		// 将 []byte 转为字符串
		bodyStr := string(bodyBytes)
		return UserInfoResponse{}, fmt.Errorf("token endpoint status %d, body: %s", resp.StatusCode, bodyStr)
	}

	var info UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return UserInfoResponse{}, err
	}
	return info, nil
}

// ====== 工具函数 ======

func getSessionFromRequest(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return Session{}, false
	}
	sessionID := cookie.Value

	sessionMu.RLock()
	sess, ok := sessionStore[sessionID]
	sessionMu.RUnlock()
	if !ok {
		return Session{}, false
	}
	// 简单过期检查
	if time.Now().After(sess.ExpireAt) {
		sessionMu.Lock()
		delete(sessionStore, sessionID)
		sessionMu.Unlock()
		return Session{}, false
	}
	return sess, true
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	max := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := cryptoRand.Int(cryptoRand.Reader, max)
		if err != nil {
			return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
