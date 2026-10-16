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

	"mybilibili/bili-proxy/internal/bilibili"
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
	// B 站未登录/低等级账号最多拿到 720P + 360P 两档。
	// 高 qn 会被静默降级，所以不写 play_url_hd，避免前端展示一个实际拿不到的"1080P"档。
	_, err = i.db.Exec(`UPDATE videos SET
		play_url_sd = $1, play_url_ld = $2
		WHERE id = $3`,
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
		res, err := i.db.Exec(`UPDATE manuscripts SET category_id = $1, updated_at = NOW() WHERE id = $2 AND category_id <> $1`, cat, it.id)
		if err != nil {
			return n, err
		}
		if rows, _ := res.RowsAffected(); rows > 0 {
			n++
		}
	}
	return n, nil
}

// tidByName 真 B 站分区名 → tid（一级分区 + 常用二级分区名）。
var tidByName = map[string]int{
	// 一级分区
	"动画": 1, "番剧": 13, "国创": 167, "音乐": 3, "舞蹈": 129, "游戏": 4,
	"知识": 36, "科技": 188, "运动": 234, "汽车": 223, "生活": 160, "美食": 211,
	"动物圈": 217, "鬼畜": 119, "时尚": 155, "资讯": 202, "娱乐": 5, "影视": 181,
	"纪录片": 177, "电影": 23, "电视剧": 11,
	// 二级分区（按知名度与常见命中优先）
	"MAD·AMV": 24, "MMD·3D": 25, "短片·手书": 47, "配音": 257, "手办·模玩": 210,
	"特摄": 86, "动漫杂谈": 253, "综合": 27,
	"完结动画": 32, "连载动画": 33, "官方延伸": 152,
	"国产动画": 153, "国产原创相关": 168, "布袋戏": 169, "动态漫·广播剧": 195,
	"原创音乐": 28, "翻唱": 31, "VOCALOID·UTAU": 30, "演奏": 59, "MV": 193,
	"音乐现场": 29, "音乐综合": 130, "乐评盘点": 243, "音乐教学": 244,
	"宅舞": 20, "舞蹈综合": 154, "舞蹈教程": 156, "街舞": 198, "明星舞蹈": 199,
	"国风舞蹈": 200, "手势·网红舞": 255,
	"单机游戏": 17, "电子竞技": 171, "手机游戏": 172, "网络游戏": 65, "桌游棋牌": 173,
	"GMV": 121, "音游": 136, "Mugen": 19,
	"科学科普": 201, "社科·法律·心理": 124, "社科人文": 124, "人文历史": 228,
	"财经商业": 207, "校园学习": 208, "职业职场": 209, "设计·创意": 229,
	"野生技术协会": 122, "野生技能协会": 122,
	"数码": 95, "软件应用": 230, "计算机技术": 231, "科工机械": 232, "极客DIY": 233,
	"篮球·足球": 235, "篮球": 235, "足球": 249, "健身": 164, "竞技体育": 236,
	"运动文化": 237, "运动综合": 238,
	"汽车知识科普": 258, "赛车": 245, "改装玩车": 246, "新能源车": 247, "房车": 248,
	"摩托车": 240, "购车攻略": 227, "汽车生活": 176,
	"搞笑": 138, "出行": 250, "三农": 251, "家居房产": 239, "手工": 161, "绘画": 162,
	"日常": 21, "亲子": 254,
	"美食制作": 76, "美食侦探": 212, "美食测评": 213, "田园美食": 214, "美食记录": 215,
	"喵星人": 218, "汪星人": 219, "动物二创": 220, "野生动物": 221, "小宠异宠": 222,
	"动物综合": 75,
	"鬼畜调教": 22, "音MAD": 26, "人力VOCALOID": 126, "鬼畜剧场": 216, "教程演示": 127,
	"美妆护肤": 157, "仿妆cos": 252, "穿搭": 158, "时尚潮流": 159,
	"热点": 203, "环球": 204, "社会": 205,
	"综艺": 71, "娱乐杂谈": 241, "粉丝创作": 242, "明星综合": 137,
	"影视杂谈": 182, "影视剪辑": 183, "小剧场": 85, "预告·资讯": 184, "短片": 256,
	"人文·历史": 37, "科学·探索·自然": 178, "军事": 179, "社会·美食·旅行": 180,
	"华语电影": 147, "欧美电影": 145, "日本电影": 146, "其他国家": 83,
	"国产剧": 185, "海外剧": 187,
}

// mapCategory 将 B 站分区名 + 标题关键词映射到真 B 站 tid 分区。
// 顺序：分区名精确匹配(二级/一级) → 标题关键词规则 → 兜底 科技(188)。
func mapCategory(tname, title string) int {
	t := strings.TrimSpace(tname)
	if id, ok := tidByName[t]; ok {
		return id
	}
	tt := strings.ToLower(title)
	type rule struct{ cat int; kws []string }
	rules := []rule{
		{231, []string{"ai", "人工智能", "chatgpt", "大模型", "模型", "智能", "token", "openclaw", "agent", "mcp", "编程", "代码", "linux", "python", "docker", "mysql", "正则", "命令行", "github", "部署", "软件", "开发", "系统"}},
		{230, []string{"工具", "插件"}},
		{95, []string{"数码", "手机", "评测", "电脑", "5g", "硬件", "芯片", "电路", "开发板", "单片机"}},
		{232, []string{"机械", "发动机", "工程"}},
		{201, []string{"数学", "函数", "几何", "微积分", "代数", "物理", "量子", "力学", "原子", "化学"}},
		{208, []string{"英语", "english", "新概念", "雅思", "托福", "四级", "六级", "学习", "考试", "数学", "物理", "试卷", "启蒙"}},
		{164, []string{"运动", "健身", "体育", "跑步", "篮球", "足球", "热身", "训练"}},
		{124, []string{"心理", "精神", "情绪", "焦虑", "抑郁", "基因", "童年"}},
		{228, []string{"历史", "古代", "王朝", "历史人文", "文学"}},
		{207, []string{"经济", "财经", "金融", "股票", "房价", "投资"}},
		{203, []string{"政治", "时政", "美国", "以色列", "国际", "选举", "特朗普"}},
	}
	hay := tt + " " + t
	for _, r := range rules {
		for _, kw := range r.kws {
			if strings.Contains(hay, kw) {
				return r.cat
			}
		}
	}
	return 188
}