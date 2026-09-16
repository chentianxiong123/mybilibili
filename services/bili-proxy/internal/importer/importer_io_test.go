package importer

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/bili-proxy/internal/bilibili"
)

func newImpDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestNew_DefaultPlayPrefix(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{})
	require.NotNil(t, imp)
	assert.Equal(t, "/api/v1/bili/stream/", imp.client.PlayPrefix())
}

func TestNew_CustomPlayPrefix(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{Sessdata: "sess", PlayPrefix: "/custom/prefix/"})
	require.NotNil(t, imp)
	assert.Equal(t, "/custom/prefix/", imp.client.PlayPrefix())
}

func TestSha256Hex(t *testing.T) {
	// sha256("abc")
	assert.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", sha256hex("abc"))
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sha256hex(""))
}

func TestRandomStr(t *testing.T) {
	s := randomStr(12)
	assert.Len(t, s, 12)
	for _, c := range s {
		assert.Contains(t, "abcdefghijklmnopqrstuvwxyz0123456789", string(c))
	}
	assert.Equal(t, "", randomStr(0))
}

func TestRandomPassword(t *testing.T) {
	p1 := randomPassword()
	p2 := randomPassword()
	assert.Len(t, p1, 64)
	assert.Len(t, p2, 64)
	assert.NotEqual(t, p1, p2)
}

func TestBvExists_True(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	imp := New(db, Config{})
	ok, err := imp.bvExists("BV1xx411c7mD")
	require.NoError(t, err)
	assert.True(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBvExists_False(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	imp := New(db, Config{})
	ok, err := imp.bvExists("BV1xx411c7mD")
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBvExists_Error(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).WillReturnError(errors.New("db down"))
	imp := New(db, Config{})
	_, err := imp.bvExists("BV1xx411c7mD")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== ensureUser ====================

func TestEnsureUser_ZeroMid(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{})
	id, err := imp.ensureUser(bilibili.ViewOwner{Mid: 0})
	require.NoError(t, err)
	assert.Zero(t, id)
}

func TestEnsureUser_Existing(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	imp := New(db, Config{})
	id, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100, Name: "UP"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureUser_QueryError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(errors.New("db down"))
	imp := New(db, Config{})
	_, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100, Name: "UP"})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureUser_InsertOk(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	imp := New(db, Config{})
	id, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100, Name: "UP"})
	require.NoError(t, err)
	assert.Equal(t, int64(6), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureUser_InsertOk_EmptyName(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	imp := New(db, Config{})
	id, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(6), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureUser_InsertConflictReselect(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	imp := New(db, Config{})
	id, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100, Name: "UP"})
	require.NoError(t, err)
	assert.Equal(t, int64(7), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureUser_InsertError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(errors.New("constraint"))
	imp := New(db, Config{})
	_, err := imp.ensureUser(bilibili.ViewOwner{Mid: 100, Name: "UP"})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== importOne ====================

func TestImportOne_ZeroAIDCID(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{})
	err := imp.importOne(&bilibili.ViewInfo{AID: 0, CID: 0})
	require.NoError(t, err)
}

func TestImportOne_EnsureUserError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(errors.New("db down"))
	imp := New(db, Config{})
	err := imp.importOne(&bilibili.ViewInfo{AID: 1, CID: 2, Owner: bilibili.ViewOwner{Mid: 3}})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportOne_ManuscriptInsertError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO manuscripts`).WillReturnError(errors.New("insert fail"))
	imp := New(db, Config{})
	err := imp.importOne(&bilibili.ViewInfo{AID: 1, CID: 2, Owner: bilibili.ViewOwner{Mid: 3, Name: "UP"}})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportOne_VideoInsertError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id FROM users WHERE username`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnError(errors.New("insert fail"))
	imp := New(db, Config{})
	err := imp.importOne(&bilibili.ViewInfo{AID: 1, CID: 2, Owner: bilibili.ViewOwner{Mid: 3, Name: "UP"}})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportOne_Success(t *testing.T) {
	db, mock := newImpDB(t)
	// owner.Mid==0 → ensureUser 直接返回 0 → uploaderID 兜底为 1（无 users 查询）
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(`INSERT INTO videos`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(22))
	mock.ExpectExec(`UPDATE videos SET`).WillReturnResult(sqlmock.NewResult(0, 1))

	imp := New(db, Config{})
	info := &bilibili.ViewInfo{
		BVID:     "BV1xx411c7mD",
		AID:      1,
		CID:      2,
		Title:    "测试视频",
		Desc:     "描述",
		Pic:      "http://i0.hdslb.com/a.jpg",
		Duration: 3661,
		Pubdate:  1600000000,
		TName:    "科技",
		Stat:     bilibili.ViewStat{View: 10, Like: 2, Coin: 1, Favorite: 3, Share: 1, Reply: 2, Danmaku: 5},
	}
	err := imp.importOne(info)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportOne_Success_CoverNoScheme(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(`INSERT INTO videos`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(22))
	mock.ExpectExec(`UPDATE videos SET`).WillReturnResult(sqlmock.NewResult(0, 1))

	imp := New(db, Config{})
	info := &bilibili.ViewInfo{
		BVID:     "BV1xx411c7mD",
		AID:      1,
		CID:      2,
		Title:    "无协议封面",
		Desc:     "",
		Pic:      "//i0.hdslb.com/b.jpg",
		Duration: 60,
		Pubdate:  1600000000,
		TName:    "",
		Stat:     bilibili.ViewStat{View: 1, Like: 0},
	}
	err := imp.importOne(info)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Run ====================

func TestRun_EmptyBVs(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{})
	imported, skipped, err := imp.Run(nil)
	require.NoError(t, err)
	assert.Zero(t, imported)
	assert.Zero(t, skipped)
}

func TestRun_BvExistsError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).WillReturnError(errors.New("db down"))
	imp := New(db, Config{})
	imported, skipped, err := imp.Run([]string{"BV1xx411c7mD"})
	assert.Error(t, err)
	assert.Zero(t, imported)
	assert.Zero(t, skipped)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRun_SkipsExisting(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	imp := New(db, Config{})
	imported, skipped, err := imp.Run([]string{"BV1xx411c7mD", "BV1yy441c7mE"})
	require.NoError(t, err)
	assert.Zero(t, imported)
	assert.Equal(t, 2, skipped)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRun_TargetRespectedOnSkips(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	imp := New(db, Config{Target: 1})
	imported, skipped, err := imp.Run([]string{"BV1xx411c7mD"})
	require.NoError(t, err)
	assert.Zero(t, imported)
	assert.Equal(t, 1, skipped)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== ReclassifyAll ====================

func TestReclassifyAll_QueryError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id, title FROM manuscripts WHERE source_type`).WillReturnError(errors.New("db down"))
	imp := New(db, Config{})
	n, err := imp.ReclassifyAll()
	assert.Error(t, err)
	assert.Zero(t, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReclassifyAll_ScanError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id, title FROM manuscripts WHERE source_type`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow("bad", "title"))
	imp := New(db, Config{})
	n, err := imp.ReclassifyAll()
	assert.Error(t, err)
	assert.Zero(t, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReclassifyAll_ExecError(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id, title FROM manuscripts WHERE source_type`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow(1, "AI 模型入门"))
	mock.ExpectExec(`UPDATE manuscripts SET category_id`).WillReturnError(errors.New("update fail"))
	imp := New(db, Config{})
	_, err := imp.ReclassifyAll()
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReclassifyAll_Success(t *testing.T) {
	db, mock := newImpDB(t)
	mock.ExpectQuery(`SELECT id, title FROM manuscripts WHERE source_type`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow(1, "AI 大模型实战").
			AddRow(2, "随便聊聊"))
	mock.ExpectExec(`UPDATE manuscripts SET category_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET category_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	imp := New(db, Config{})
	n, err := imp.ReclassifyAll()
	require.NoError(t, err)
	// "AI 大模型实战" 映射到 1（非 11），"随便聊聊" 兜底 11
	assert.Equal(t, 1, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== mapCategory partition switch ====================

func TestMapCategory_GamePartitions(t *testing.T) {
	// "电子竞技" 因含关键词 "电子" 走规则 2；"游戏"/"单机游戏" 走分区名兜底
	for _, tname := range []string{"游戏", "单机游戏"} {
		assert.Equal(t, 11, mapCategory(tname, ""), "tname=%s", tname)
	}
	assert.Equal(t, 2, mapCategory("电子竞技", ""))
}

// ==================== CollectBVs (真实文件 I/O + exec) ====================

// writeFakeChrome 在 t.TempDir() 写一个假 chrome 脚本：输出含 BV 号的 DOM；URL 含 "fail" 时退出码 1。
func writeFakeChrome(t *testing.T, failOn string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "chrome")
	body := `#!/bin/sh
for a in "$@"; do
  case "$a" in
    *FAILTOKEN*) exit 1;;
  esac
done
printf '%s\n' '<html><div class="bili-video-card">https://www.bilibili.com/video/BV1xx411c7mD</div>'
printf '%s\n' '<div>https://www.bilibili.com/video/BV1yy441c7mE</div></html>'
`
	body = strings.ReplaceAll(body, "FAILTOKEN", failOn)
	require.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script
}

func TestCollectBVs_DedupAcrossPartitions(t *testing.T) {
	db, _ := newImpDB(t)
	chrome := writeFakeChrome(t, "never-match")
	imp := New(db, Config{
		ChromePath: chrome,
		Partitions: []string{"http://p1", "http://p2"},
	})
	bvs, err := imp.CollectBVs()
	require.NoError(t, err)
	// 两个分区输出相同 BV 集合，去重保序
	assert.Equal(t, []string{"BV1xx411c7mD", "BV1yy441c7mE"}, bvs)
}

func TestCollectBVs_PartialFailure(t *testing.T) {
	db, _ := newImpDB(t)
	chrome := writeFakeChrome(t, "http://bad")
	imp := New(db, Config{
		ChromePath: chrome,
		Partitions: []string{"http://ok", "http://bad", "http://ok"},
	})
	bvs, err := imp.CollectBVs()
	require.NoError(t, err)
	assert.Equal(t, []string{"BV1xx411c7mD", "BV1yy441c7mE"}, bvs)
}

func TestCollectBVs_NonexistentChrome(t *testing.T) {
	db, _ := newImpDB(t)
	imp := New(db, Config{
		ChromePath: filepath.Join(t.TempDir(), "no-such-chrome"),
		Partitions: []string{"http://p1"},
	})
	bvs, err := imp.CollectBVs()
	require.NoError(t, err)
	assert.Empty(t, bvs)
}
