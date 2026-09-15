package bilibili

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	Referer       = "https://www.bilibili.com/"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	minInterval   = 300 * time.Millisecond
	playURLPrefix = "/api/v1/bili/stream/"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Client 封装对 B 站 API 的调用（view / playurl）。
type Client struct {
	sessdata   string
	lastCall   time.Time
	playPrefix string
}

func NewClient(sessdata string) *Client {
	return &Client{sessdata: sessdata, playPrefix: playURLPrefix}
}

// SetPlayPrefix 设置回写给 DB 的 play_url 前缀（core 反代路径）。
func (c *Client) SetPlayPrefix(p string) { c.playPrefix = p }

func (c *Client) PlayPrefix() string { return c.playPrefix }

// throttle 简单的全局限流，避免触发风控。
func (c *Client) throttle() {
	elapsed := time.Since(c.lastCall)
	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}
	c.lastCall = time.Now()
}

func (c *Client) doGet(rawurl string) ([]byte, error) {
	c.throttle()
	req, err := http.NewRequest("GET", rawurl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", Referer)
	if c.sessdata != "" {
		req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: c.sessdata})
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bilibili http %d for %s", resp.StatusCode, rawurl)
	}
	return io.ReadAll(resp.Body)
}

type APIResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ViewInfo B 站 view 接口关键字段。
type ViewInfo struct {
	BVID       string          `json:"bvid"`
	AID        int64           `json:"aid"`
	CID        int64           `json:"cid"`
	Title      string          `json:"title"`
	Desc       string          `json:"desc"`
	Pic        string          `json:"pic"`
	Duration   int64           `json:"duration"`
	Pubdate    int64           `json:"pubdate"`
	TName      string          `json:"tname"`
	Owner      ViewOwner       `json:"owner"`
	Stat       ViewStat        `json:"stat"`
	Pages      []ViewPage      `json:"pages"`
	Tag        json.RawMessage `json:"tag"`
	Everything json.RawMessage `json:"-"`
}

type ViewOwner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type ViewStat struct {
	View     int64 `json:"view"`
	Danmaku  int64 `json:"danmaku"`
	Reply    int64 `json:"reply"`
	Favorite int64 `json:"favorite"`
	Coin     int64 `json:"coin"`
	Share    int64 `json:"share"`
	Like     int64 `json:"like"`
}

type ViewPage struct {
	CID        int64  `json:"cid"`
	Page       int    `json:"page"`
	Part       string `json:"part"`
	Duration   int64  `json:"duration"`
	FirstFrame string `json:"first_frame"`
}

// GetView 拉取视频元信息。
func (c *Client) GetView(bvid string) (*ViewInfo, error) {
	u := "https://api.bilibili.com/x/web-interface/view?bvid=" + url.QueryEscape(bvid)
	body, err := c.doGet(u)
	if err != nil {
		return nil, err
	}
	var v APIResp
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	if v.Code != 0 {
		return nil, fmt.Errorf("view api code=%d msg=%s", v.Code, v.Message)
	}
	var info ViewInfo
	if err := json.Unmarshal(v.Data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// DurlInfo playurl 的 durl 结构（mp4 流）。
type DurlInfo struct {
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Length   int64  `json:"length"`
}

type PlayURLData struct {
	Format       string      `json:"format"`
	Quality      int         `json:"quality"`
	AcceptQuality []int      `json:"accept_quality"`
	DURL         []DurlInfo  `json:"durl"`
	Dash         *DashResult `json:"dash"`
}

type DashResult struct {
	Video []DashStream `json:"video"`
	Audio []DashStream `json:"audio"`
}

type DashStream struct {
	ID       int      `json:"id"`
	BaseURL  string   `json:"base_url"`
	BaseURLC string   `json:"baseUrl"`
	Bandwidth int64   `json:"bandwidth"`
	MimeType string   `json:"mime_type"`
	Codecs   string   `json:"codecs"`
}

// GetPlayURL 以 html5 platform 获取 MP4 durl（未登录可拿到 720p，无需 WBI）。
func (c *Client) GetPlayURL(bvid string, cid int64, qn int) (*PlayURLData, error) {
	u := fmt.Sprintf("https://api.bilibili.com/x/player/playurl?bvid=%s&cid=%d&qn=%d&fnval=0&platform=html5",
		url.QueryEscape(bvid), cid, qn)
	body, err := c.doGet(u)
	if err != nil {
		return nil, err
	}
	var v APIResp
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	if v.Code != 0 {
		return nil, fmt.Errorf("playurl code=%d msg=%s", v.Code, v.Message)
	}
	var d PlayURLData
	if err := json.Unmarshal(v.Data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ResolveStream 解析可播放的 CDN 直链。优先 durl(mp4)，fallback dash video。
func (c *Client) ResolveStream(bvid string, cid int64, qn int) (string, string, error) {
	pu, err := c.GetPlayURL(bvid, cid, qn)
	if err != nil {
		return "", "", err
	}
	if len(pu.DURL) > 0 {
		return pu.DURL[0].URL, "video/mp4", nil
	}
	if pu.Dash != nil && len(pu.Dash.Video) > 0 {
		// dash 视频流基本是 mp4 codec，但浏览器无法直接播 dash；这里回退一个 url 供降级。
		return pu.Dash.Video[0].BaseURL, "video/mp4", nil
	}
	return "", "", fmt.Errorf("no playable stream for bvid=%s cid=%d qn=%d", bvid, cid, qn)
}