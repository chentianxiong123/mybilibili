"""视频页点赞 + dislike 折叠回归测试

回归目标：
  1. 已登录用户点 ❤ → POST /api/v1/manuscript/{id}/like，DB likeCount +1，
     UI 计数同步 +1 且按钮高亮；测试结束 DELETE 还原。
  2. 点 👎 → 评论区折叠（纯前端 localStorage，不发任何 HTTP），
     刷新保持折叠；再点 👎 或点折叠条 → 展开。
"""


LIKE_BUTTON_JS = r"""
() => {
  // 找带 icon-dianzan 的 toolbar 项
  const items = document.querySelectorAll('.video-toolbar-left-item');
  for (const it of items) {
    if (it.querySelector('.icon-dianzan')) return it;
  }
  return null;
}
"""

DISLIKE_BUTTON_JS = r"""
() => {
  const items = document.querySelectorAll('.video-toolbar-left-item');
  for (const it of items) {
    if (it.querySelector('.icon-diancai')) return it;
  }
  return null;
}
"""

COMMENT_FOLDED_JS = r"""
() => Boolean(document.querySelector('.comment-folded'))
"""

COMMENT_TREE_JS = r"""
() => Boolean(document.querySelector('.comment-container'))
"""

LIKE_COUNT_TEXT_JS = r"""
() => {
  const items = document.querySelectorAll('.video-toolbar-left-item');
  for (const it of items) {
    if (it.querySelector('.icon-dianzan')) {
      return it.querySelector('.video-toolbar-item-text')?.textContent?.trim() || '';
    }
  }
  return '';
}
"""


class TestVideoLove:
    """点击 ❤ 真实打后端 /manuscript/{id}/like"""

    def test_love_toggles_like_count(self, logged_in_page, base_url):
        """点 ❤：后端 likeCount +1；清理回原状。"""
        page = logged_in_page
        token = page.evaluate("() => localStorage.getItem('teri_token')")
        hdrs = {"Authorization": f"Bearer {token}"}

        # 先归零：确保处于未点赞状态（避免上一轮残留导致幂等不计数）
        page.request.delete(f"{base_url}/api/v1/manuscript/10/like", headers=hdrs)

        # 基线 likeCount
        resp = page.request.get(f"{base_url}/api/v1/manuscript/detail/10")
        assert resp.status == 200
        before = resp.json()["data"]["likeCount"]

        # 进视频页（vid 25 → manuscript 10）
        page.goto(f"{base_url}/video/25", wait_until="networkidle")
        ui_before = page.evaluate(LIKE_COUNT_TEXT_JS)

        # 点 ❤
        like_btn = page.evaluate_handle(LIKE_BUTTON_JS).as_element()
        assert like_btn is not None, "找不到点赞按钮"
        like_btn.click()

        # 等前端把请求打完
        page.wait_for_timeout(1200)

        # UI 上的计数也应 +1（handleNum 对小数字原样输出）
        ui_after = page.evaluate(LIKE_COUNT_TEXT_JS)
        assert ui_after == str(before + 1), f"UI 计数应变为 {before + 1}，实际 {ui_after!r}（原 {ui_before!r}）"
        assert page.locator(".video-toolbar-left-item.on .icon-dianzan").count() == 1, "点赞按钮应高亮"

        # after：直接查 detail（likeCount 由 LikeManuscript RPC 落库）
        after = page.request.get(
            f"{base_url}/api/v1/manuscript/detail/10"
        ).json()["data"]["likeCount"]
        assert after == before + 1, f"点赞后 likeCount 应 {before + 1}，实际 {after}"

        # 清理：取消点赞恢复原状
        page.request.delete(f"{base_url}/api/v1/manuscript/10/like", headers=hdrs)
        restored = page.request.get(
            f"{base_url}/api/v1/manuscript/detail/10"
        ).json()["data"]["likeCount"]
        assert restored == before, f"清理后应回到 {before}，实际 {restored}"


class TestVideoDislike:
    """点 👎 纯前端折叠评论区，不发任何 HTTP"""

    def test_dislike_folds_comment_without_network(self, logged_in_page, base_url):
        """点 👎：评论区折叠，无后端请求。"""
        page = logged_in_page
        # 归零：清掉上一轮残留的 dislike localStorage
        page.evaluate("() => localStorage.removeItem('mybilibili_disliked_vids')")

        # 监听网络：禁 manuscript/*/like 请求
        bad_calls = []
        def on_request(req):
            if "/like" in req.url and "manuscript" in req.url:
                bad_calls.append(req.url)
        page.on("request", on_request)

        page.goto(f"{base_url}/video/25", wait_until="networkidle")
        # 确保未折叠
        assert not page.evaluate(COMMENT_FOLDED_JS), "初始状态评论区不应折叠"

        # 点 👎
        dislike_btn = page.evaluate_handle(DISLIKE_BUTTON_JS).as_element()
        assert dislike_btn is not None, "找不到不喜欢按钮"
        dislike_btn.click()
        page.wait_for_timeout(500)

        # 折叠出现，原评论区消失
        assert page.evaluate(COMMENT_FOLDED_JS), "点不喜欢后应折叠"
        assert not page.evaluate(COMMENT_TREE_JS), "折叠后不应有评论区 DOM"

        # 不应发 manuscript/like 请求
        assert bad_calls == [], f"dislike 不应触发 HTTP：{bad_calls}"

        # 刷新后仍保持折叠（localStorage 持久化）
        page.reload(wait_until="networkidle")
        assert page.evaluate(COMMENT_FOLDED_JS), "刷新后仍应折叠"
        assert "不喜欢" in page.locator(".comment-folded").inner_text(), "折叠条应提示不喜欢"

        # 展开：点折叠条
        page.locator(".comment-folded").click()
        page.wait_for_timeout(300)
        assert not page.evaluate(COMMENT_FOLDED_JS), "点折叠条应展开"
        assert page.evaluate(COMMENT_TREE_JS), "展开后评论区应回来"

        # 再点 👎 → 折叠；再点 👎 → 展开
        page.evaluate_handle(DISLIKE_BUTTON_JS).as_element().click()
        page.wait_for_timeout(300)
        assert page.evaluate(COMMENT_FOLDED_JS), "再次点不喜欢应折叠"
        page.evaluate_handle(DISLIKE_BUTTON_JS).as_element().click()
        page.wait_for_timeout(300)
        assert not page.evaluate(COMMENT_FOLDED_JS), "再次点应展开"

        # 清理 localStorage，避免影响后续用例
        page.evaluate("() => localStorage.removeItem('mybilibili_disliked_vids')")