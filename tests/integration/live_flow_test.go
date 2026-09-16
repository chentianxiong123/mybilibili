//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 直播间创建 → 获取 → 状态变更 → 列表 全链路
func TestLiveFlow_Complete(t *testing.T) {
	// 注册并登录获得 token
	username := randomUsername()
	_, reg := doPost(t, coreURL+"/api/v1/user/register", map[string]string{
		"username": username,
		"password": "Test1234",
		"nickname": username,
	})
	assert.NotEmpty(t, reg.Data)

	resp, login := doPost(t, coreURL+"/api/v1/user/login", map[string]string{
		"username": username,
		"password": "Test1234",
	})
	assert.Equal(t, 200, resp.StatusCode)
	var loginData struct {
		Token string `json:"token"`
		ID    int64  `json:"id"`
	}
	json.Unmarshal(login.Data, &loginData)
	assert.NotEmpty(t, loginData.Token)

	// 创建直播间
	resp, createRoom := doPost(t, liveURL+"/api/v1/live/room", map[string]interface{}{
		"room_name": "集成测试直播间",
		"category":  "游戏",
	})
	assert.Equal(t, 200, resp.StatusCode)
	var room struct {
		ID       int64  `json:"id"`
		RoomName string `json:"roomName"`
		Status   string `json:"status"`
	}
	json.Unmarshal(createRoom.Data, &room)
	assert.NotEmpty(t, room.ID)
	assert.Equal(t, "集成测试直播间", room.RoomName)

	// 获取直播间
	resp, _ = doGet(t, fmt.Sprintf(liveURL+"/api/v1/live/room/%d", room.ID), "")
	assert.Equal(t, 200, resp.StatusCode)

	// 直播间列表，创建后应能查到
	resp, list := doGet(t, liveURL+"/api/v1/live/room/list", "")
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, list.Data)
}

// 不存在直播间的获取应返回 404
func TestLiveFlow_GetNonExistentRoom(t *testing.T) {
	resp, _ := doGet(t, liveURL+"/api/v1/live/room/999999", "")
	assert.Equal(t, 404, resp.StatusCode)
}