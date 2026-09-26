"""WAP 路由与页面：排行榜 / 个人空间 / 竖屏 / 未匹配路由 404。"""


def test_ranking_page_renders_rank_items(mobile_page, wap_url):
    # 路由是 /m/ranking/:rId，rId=0 表示「全站」分区
    mobile_page.goto(f"{wap_url}/m/ranking/0", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".ranking-page", timeout=15000)

    assert mobile_page.locator(".ranking-header .title").text_content().strip() == "排行榜"
    assert mobile_page.locator(".ranking-header .back-btn").count() == 1
    assert mobile_page.locator(".tab-bar-wrapper").count() == 1
    assert mobile_page.locator(".video-list .rank-item").count() > 0


def test_ranking_rank_item_click_navigates_to_detail(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/ranking/0", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-list .rank-item", timeout=15000)

    mobile_page.locator(".video-list .rank-item").first.tap()
    mobile_page.wait_for_url("**/m/video/**", timeout=15000)
    assert "/m/video/" in mobile_page.url


def test_ranking_back_btn_returns_to_index(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/ranking/0", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".ranking-header .back-btn", timeout=15000)

    mobile_page.locator(".ranking-header .back-btn").tap()
    mobile_page.wait_for_url("**/m/index**", timeout=10000)
    assert "/m/index" in mobile_page.url


def test_space_requires_login_and_redirects_to_login(mobile_page, wap_url):
    """未登录访问 /m/space 应被替换成登录页（Space.vue onMounted 的 isLogin 判定）。"""
    mobile_page.goto(f"{wap_url}/m/space", wait_until="domcontentloaded")
    mobile_page.wait_for_url("**/m/login**", timeout=15000)
    assert "/m/login" in mobile_page.url
    assert mobile_page.locator(".login-page").count() == 1


def test_space_renders_when_logged_in(wap_session, wap_url):
    wap_session.goto(f"{wap_url}/m/space", wait_until="domcontentloaded")
    wap_session.wait_for_selector(".space-page", timeout=15000)

    # Space.vue：.top-nav 是顶部图标栏，.user-info-section 由 userInfo 控制
    assert wap_session.locator(".space-page .top-nav").count() == 1
    assert wap_session.locator(".space-page .user-info-section").count() == 1
    assert wap_session.locator(".space-page .stats-row .stat-item").count() >= 3
    assert wap_session.locator(".space-page .common-tools-grid .tool-item").count() > 0
    nickname = wap_session.locator(".space-page .nickname").text_content().strip()
    assert nickname, "logged-in space should show nickname"


def test_vertical_page_route_responds(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/vertical/11", wait_until="domcontentloaded")
    mobile_page.wait_for_timeout(2500)

    # 竖屏是独立全屏页，至少不应是 404 页
    assert mobile_page.locator(".not-found").count() == 0
    assert "/m/vertical/" in mobile_page.url


def test_unknown_route_shows_not_found(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/definitely-not-a-page", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".not-found", timeout=15000)

    assert mobile_page.locator(".not-found .error-code").text_content().strip() == "404"
    assert "页面不存在" in mobile_page.locator(".not-found .message").text_content()
    assert mobile_page.locator(".not-found .back-btn").count() == 1


def test_not_found_back_btn_returns_home(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/definitely-not-a-page", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".not-found .back-btn", timeout=15000)

    mobile_page.locator(".not-found .back-btn").tap()
    mobile_page.wait_for_url("**/m/index**", timeout=10000)
    assert "/m/index" in mobile_page.url


def test_login_page_route_and_browse_link(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-page", timeout=15000)

    assert mobile_page.locator(".login-form").count() == 1
    assert mobile_page.locator(".login-form input").count() >= 2
    assert mobile_page.locator(".login-btn").count() == 1
    assert "先逛逛" in mobile_page.locator(".login-links a").text_content()
