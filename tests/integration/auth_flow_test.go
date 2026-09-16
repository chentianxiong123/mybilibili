//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 完整认证链路：注册 → 登录 → 获取个人信息 → token刷新
func TestAuthFlow_Complete(t *testing.T) {
	username := randomUsername()

	// Step 1: 注册
	resp, body := doPost(t, coreURL+"/api/v1/user/register", map[string]string{
		"username": username,
		"password": "Test1234",
		"nickname": username,
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, body.Data)

	// Step 2: 登录
	resp, body = doPost(t, coreURL+"/api/v1/user/login", map[string]string{
		"username": username,
		"password": "Test1234",
	})
	assert.Equal(t, 200, resp.StatusCode)
	var loginData struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		ID           int64  `json:"id"`
	}
	json.Unmarshal(body.Data, &loginData)
	assert.NotEmpty(t, loginData.Token)
	assert.NotEmpty(t, loginData.RefreshToken)

	// Step 3: 获取当前用户
	resp, body = doGet(t, coreURL+"/api/v1/user/me", loginData.Token)
	assert.Equal(t, 200, resp.StatusCode)

	// Step 4: 刷新 token
	resp, body = doPost(t, coreURL+"/api/v1/user/token/refresh", map[string]string{
		"refreshToken": loginData.RefreshToken,
	})
	assert.Equal(t, 200, resp.StatusCode)

	// Step 5: 用新 token 访问
	var refreshData struct {
		Token string `json:"token"`
	}
	json.Unmarshal(body.Data, &refreshData)
	assert.NotEmpty(t, refreshData.Token)
	resp, _ = doGet(t, coreURL+"/api/v1/user/me", refreshData.Token)
	assert.Equal(t, 200, resp.StatusCode)
}

// 错误密码登录
func TestAuthFlow_WrongPassword(t *testing.T) {
	resp, _ := doPost(t, coreURL+"/api/v1/user/login", map[string]string{
		"username": "nonexistent",
		"password": "wrong",
	})
	assert.Contains(t, []int{401, 404}, resp.StatusCode)
}

// 重复注册
func TestAuthFlow_DuplicateRegister(t *testing.T) {
	username := randomUsername()
	body := map[string]string{"username": username, "password": "Test1234"}
	doPost(t, coreURL+"/api/v1/user/register", body)
	resp, _ := doPost(t, coreURL+"/api/v1/user/register", body)
	assert.Contains(t, []int{400, 409}, resp.StatusCode)
}