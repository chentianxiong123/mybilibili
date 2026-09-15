package importer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mybilibili/bili-proxy-service/internal/bilibili"
)

var bvRe = regexp.MustCompile(`BV[0-9A-Za-z]{10}`)

// Config 导入配置。
type Config struct {
	ChromePath string
	Partitions []string // 分区 URL 列表
	Target     int      // 目标导入条数
	Sessdata   string
	PlayPrefix string
}

type Importer struct {
	db     *sql.DB
	client *bilibili.Client
	cfg    Config
}

func New(db *sql.DB, cfg Config) *Importer {
	c := bilibili.NewClient(cfg.Sessdata)
	if cfg.PlayPrefix != "" {
		c.SetPlayPrefix(cfg.PlayPrefix)
	}
	return &Importer{db: db, client: c, cfg: cfg}
}

// CollectBVs 用 chrome headless dump 分区页 DOM 提取 BV 号（去重保序）。
func (i *Importer) CollectBVs() ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, u := range i.cfg.Partitions {
		f, err := os.CreateTemp("", "bili_dom_*.html")
		if err != nil {
			return nil, err
		}
		name := f.Name()

		cmd := exec.Command(i.cfg.ChromePath, "--headless=new", "--disable-gpu", "--no-sandbox",
			"--virtual-time-budget=8000", "--dump-dom", u)
		cmd.Stdout = f
		cmd.Stderr = io.Discard
		err = cmd.Run()
		f.Close()
		if err != nil {
			log.Printf("chrome dump %s: %v", u, err)
			os.Remove(name)
			continue
		}
		body, err := os.ReadFile(name)
		os.Remove(name)
		if err != nil {
			continue
		}
		bvs := bvRe.FindAllString(string(body), -1)
		for _, bv := range bvs {
			if !seen[bv] {
				seen[bv] = true
				out = append(out, bv)
			}
		}
		log.Printf("dump %s -> %d bvs (total %d)", u, len(bvs), len(out))
	}
	return out, nil
}

// Run 执行导入（幂等：bvid 已存在则跳过）。
func (i *Importer) Run(bvs []string) (int, int, error) {
	imported, skipped := 0, 0
	for _, bv := range bvs {
		if i.cfg.Target > 0 && imported >= i.cfg.Target {
			break
		}
		exists, err := i.bvExists(bv)
		if err != nil {
			return imported, skipped, err
		}
		if exists {
			skipped++
			continue
		}
		info, err := i.client.GetView(bv)
		if err != nil {
			log.Printf("view %s: %v", bv, err)
			skipped++
			continue
		}
		if err := i.importOne(info); err != nil {
			log.Printf("import %s: %v", bv, err)
			skipped++
			continue
		}
		imported++
		log.Printf("imported %s %q", bv, info.Title)
	}
	return imported, skipped, nil
}

func (i *Importer) bvExists(bv string) (bool, error) {
	var n int
	err := i.db.QueryRow(`SELECT COUNT(*) FROM manuscripts WHERE bvid = $1`, bv).Scan(&n)
	return n > 0, err
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func randomPassword() string {
	return sha256hex(time.Now().String() + randomStr(12))
}

func randomStr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	seed := time.Now().UnixNano()
	for i := range b {
		seed = seed*6364136223846793005 + 1442695040888963407
		b[i] = letters[uint(seed>>33)%uint(len(letters))]
	}
	return string(b)
}

// ensureUser 为真实 B 站 UP 创建本地账号（幂等）。
func (i *Importer) ensureUser(owner bilibili.ViewOwner) (int64, error) {
	if owner.Mid == 0 {
		return 0, nil
	}
	username := "up_" + strconv.FormatInt(owner.Mid, 10)
	var id int64
	err := i.db.QueryRow(`SELECT id FROM users WHERE username = $1`, username).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	nickname := owner.Name
	if nickname == "" {
		nickname = username
	}
	err = i.db.QueryRow(`INSERT INTO users (username, password, nickname, email, avatar, level, status)
		VALUES ($1,$2,$3,$4, $5, 1, 1)
		ON CONFLICT (username) DO NOTHING RETURNING id`,
		username, randomPassword(), nickname, username+"@up.local", owner.Face).Scan(&id)
	if err == sql.ErrNoRows {
		_ = i.db.QueryRow(`SELECT id FROM users WHERE username = $1`, username).Scan(&id)
		return id, nil
	}
	return id, err
}

// importOne 导入单个稿件（首个分P）。
func (i *Importer) importOne(info *bilibili.ViewInfo) error {
	if info.AID == 0 || info.CID == 0 {
		return nil
	}
	categoryID := mapCategory(info.TName, info.Title)
	uploaderID, err := i.ensureUser(info.Owner)
	if err != nil {
		return err
	}
	if uploaderID == 0 {
		uploaderID = 1
	}
	coverURL := info.Pic
	if !strings.HasPrefix(coverURL, "http") {
		coverURL = "https:" + coverURL
	}
	coverURL = strings.Replace(coverURL, "http://i", "https://i", 1)
	var msID int64
	if err := i.db.QueryRow(`INSERT INTO manuscripts
		(title, description, cover_url, user_id, category_id, view_count, like_count,
		 coin_count, collect_count, share_count, comment_count, danmaku_count,
		 status, review_status, upload_time, updated_at, duration, duration_seconds,
		 source_type, bvid, origin_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, 3, 1,
		        to_timestamp($13), NOW(), $14, $15, 'bilibili', $16, $17)
		RETURNING id`,
		info.Title, info.Desc, coverURL, uploaderID, categoryID,
		info.Stat.View, info.Stat.Like, info.Stat.Coin, info.Stat.Favorite, info.Stat.Share,
		info.Stat.Reply, info.Stat.Danmaku,
		info.Pubdate, formatDuration(info.Duration), info.Duration,
		info.BVID, "https://www.bilibili.com/video/"+info.BVID,
	).Scan(&msID); err != nil {
		return err
	}
	var vid int64
	if err := i.db.QueryRow(`INSERT INTO videos
		(manuscript_id, video_order, title, upload_time, updated_at, process_status,
		 source_video_url, duration_seconds, cid)
		VALUES ($1,0,$2, NOW(), NOW(), 5, $3, $4, $5)
		RETURNING id`,
		msID, info.Title, "https://www.bilibili.com/video/"+info.BVID, info.Duration, info.CID,
	).Scan(&vid); err != nil {
		return err
	}
	prefix := strings.TrimRight(i.client.PlayPrefix(), "/")
	_, err = i.db.Exec(`UPDATE videos SET
		play_url_hd = $1, play_url_sd = $2, play_url_ld = $3
		WHERE id = $4`,
		prefix+"/"+strconv.FormatInt(vid, 10)+"?qn=80",
		prefix+"/"+strconv.FormatInt(vid, 10)+"?qn=64",
		prefix+"/"+strconv.FormatInt(vid, 10)+"?qn=32",
		vid)
	return err
}

func formatDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ReclassifyAll 对已导入的 bilibili 稿件按标题重新映射分类（幂等）。
func (i *Importer) ReclassifyAll() (int, error) {
	rows, err := i.db.Query(`SELECT id, title FROM manuscripts WHERE source_type = 'bilibili'`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	type item struct {
		id    int64
		title string
	}
	var list []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.title); err != nil {
			return 0, err
		}
		list = append(list, it)
	}
	n := 0
	for _, it := range list {
		cat := mapCategory("", it.title)
		if _, err := i.db.Exec(`UPDATE manuscripts SET category_id = $1, updated_at = NOW() WHERE id = $2 AND category_id <> $1`, cat, it.id); err != nil {
			return n, err
		}
		if cat != 11 {
			n++
		}
	}
	return n, nil
}

// mapCategory 将 B 站分区名 + 标题关键词映射到本地 category_id（兜底科技=11）。
func mapCategory(tname, title string) int {
	t := strings.ToLower(strings.TrimSpace(tname))
	tt := strings.ToLower(title)
	// 用关键词判定本地分类，标题优先、分区名兜底
	type rule struct{ cat int; kws []string }
	rules := []rule{
		{1, []string{"ai", "人工智能", "chatgpt", "大模型", "模型", "智能"}},
		{2, []string{"电子", "芯片", "电路", "电路板", "半导体"}},
		{3, []string{"数学", "函数", "几何", "微积分", "代数"}},
		{4, []string{"英语", "english", "新概念", "雅思", "托福", "四级", "六级"}},
		{5, []string{"运动", "健身", "体育", "跑步", "篮球", "足球", "篮球"}},
		{6, []string{"心理", "精神", "情绪", "焦虑", "抑郁"}},
		{7, []string{"软件", "编程", "开发", "代码", "程序员", "linux", "python", "前端", "后端", "系统"}},
		{8, []string{"硬件", "开发板", "单片机", "机械键盘", "显卡"}},
		{9, []string{"物理", "量子", "力学", "原子"}},
		{10, []string{"机械", "汽车", "发动机", "工程"}},
		{11, []string{"科技", "数码", "手机", "评测", "电脑", "5g"}},
		{12, []string{"政治", "时政", "美国", "以色列", "特朗普", "国际", "选举"}},
		{13, []string{"历史", "古代", "王朝", "历史人文", "毛泽东"}},
		{14, []string{"经济", "财经", "金融", "股票", "房价", "投资"}},
	}
	hay := tt + " " + t
	for _, r := range rules {
		for _, kw := range r.kws {
			if strings.Contains(hay, kw) {
				return r.cat
			}
		}
	}
	// 分区名精确匹配
	switch t {
	case "游戏", "单机游戏", "电子竞技":
		return 11
	case "动画", "鬼畜", "影视", "音乐", "舞蹈", "生活", "美食", "搞笑", "娱乐":
		return 11
	}
	return 11
}