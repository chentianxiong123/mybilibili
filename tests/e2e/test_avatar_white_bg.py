"""avatar 透明背景白底回归测试（commit 100a2673）

回归目标：
  所有用户头像 img 的 computed background-color === rgb(255, 255, 255)
  —— 不管是 VAvatar 组件、HeaderBar picture、还是 Element Plus el-avatar。

覆盖 3 个场景：
  1. 未登录访客观看 `/space/4`（VAvatar 组件路径）
  2. 已登录用户看 `/video/25` 评论区（el-avatar + 多种 class 路径）
  3. 已登录用户看 HeaderBar（picture > img，无 class）
"""


# 所有"头像 img"的判定：class 含 avatar，或父级 class 含 avatar，或 picture > img
AVATOR_JS = r"""
() => {
  const out = [];
  const all = document.querySelectorAll('img');
  all.forEach(img => {
    const cls = img.className || '';
    const parentCls = img.parentElement?.className || '';
    const grandparentCls = img.parentElement?.parentElement?.className || '';
    const grandgrandparentCls = img.parentElement?.parentElement?.parentElement?.className || '';
    const allCls = [cls, parentCls, grandparentCls, grandgrandparentCls]
      .filter(c => typeof c === 'string').join(' ');
    const isAvatarLike = /\bavatar\b/i.test(allCls);
    if (isAvatarLike) {
      const cs = getComputedStyle(img);
      out.push({
        cls,
        parentCls,
        src: img.src?.slice(-60) || '',
        bg: cs.backgroundColor,
      });
    }
  });
  return out;
}
"""


def _assert_all_avatars_white(page, scene: str):
    """公共断言：所有 avatar img 透明区域应为白色 #fff。"""
    avatars = page.evaluate(AVATOR_JS)
    assert avatars, f"[{scene}] 页面没有任何 avatar img —— 场景不对或头像没渲染"
    bad = [a for a in avatars if a["bg"] != "rgb(255, 255, 255)"]
    assert not bad, (
        f"[{scene}] {len(bad)}/{len(avatars)} 个头像透明区域不是白色:\n"
        + "\n".join(f"  cls={b['cls']!r} parent={b['parentCls']!r} bg={b['bg']} src={b['src']}" for b in bad)
    )


def test_space_page_avatar_white_bg(page, base_url):
    """场景1：未登录访问 /space/4 —— VAvatar 组件路径"""
    page.goto(f"{base_url}/space/4")
    page.wait_for_selector(".v-avatar img", timeout=5000)
    _assert_all_avatars_white(page, "space/4")


def test_video_comment_avatar_white_bg(logged_in_page, base_url):
    """场景2：登录后看 /video/25 评论区 —— el-avatar + 多种 class"""
    logged_in_page.goto(f"{base_url}/video/25")
    logged_in_page.wait_for_selector("img", timeout=5000)
    # 触发评论加载（如果有 lazy 加载）
    logged_in_page.wait_for_timeout(1000)
    _assert_all_avatars_white(logged_in_page, "video/25 comment")


def test_header_bar_avatar_white_bg(logged_in_page, base_url):
    """场景3：HeaderBar picture > img（无 class 的 img）"""
    logged_in_page.goto(f"{base_url}/")
    logged_in_page.wait_for_selector("picture.v-img img", timeout=5000)
    avatars = logged_in_page.evaluate(AVATOR_JS)
    header_avatars = [a for a in avatars if "v-img" in a["parentCls"]]
    assert header_avatars, "HeaderBar picture.v-img > img 没渲染"
    bad = [a for a in header_avatars if a["bg"] != "rgb(255, 255, 255)"]
    assert not bad, f"HeaderBar {len(bad)}/{len(header_avatars)} 个头像非白色"