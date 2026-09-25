"""Sanity test：确认 Python+Chrome+pytest 能跑通 playwright。

不验证任何业务 —— 仅验证：
  1. 启动 Chrome（系统 google-chrome 150.x）
  2. 访问 http://localhost/ 首页
  3. 标题是 "哔哩哔哩"
"""


def test_chrome_can_launch_and_reach_home(page, base_url):
    page.goto(base_url)
    assert "哔哩哔哩" in page.title()


def test_login_api_works(logged_in_page, base_url):
    """顺手验证 /api/v1/user/login 仍可用 —— 这是后续所有登录态 e2e 的前置。"""
    # 登录后访问个人空间，头像应渲染出来
    logged_in_page.goto(f"{base_url}/space/4")
    logged_in_page.wait_for_selector(".v-avatar img", timeout=5000)
    assert logged_in_page.locator(".v-avatar img").count() >= 1