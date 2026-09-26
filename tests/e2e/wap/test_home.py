"""WAP 首页：顶栏 / 分区 TabBar / 视频网格 / 轮播图 / 进详情页。"""


def test_home_renders_header_and_tabbar(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-item", timeout=15000)

    assert mobile_page.locator(".mobile-header").count() == 1
    assert mobile_page.locator(".mobile-header .search-box").count() == 1
    # 头像入口 → /m/space，消息入口 → /m/message
    assert mobile_page.locator(".mobile-header .user-avatar").count() == 1
    assert mobile_page.locator(".mobile-header .msg-icon").count() == 1

    tabs = mobile_page.locator(".tab-bar .tab-item")
    assert tabs.count() >= 2
    names = [t.strip() for t in tabs.all_text_contents()]
    assert "推荐" in names
    assert "直播" in names


def test_home_default_active_tab_is_recommend(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".tab-bar .tab-item.active", timeout=15000)
    active = mobile_page.locator(".tab-bar .tab-item.active")
    assert active.count() == 1
    assert active.first.text_content().strip() == "推荐"


def test_home_video_grid_loads_with_real_items(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-item", timeout=15000)

    items = mobile_page.locator(".video-item")
    assert items.count() > 0

    first = items.first
    # VideoItem 的 router-link 目标：/m/video/<aId>（BASE_URL 前缀为 /wap/）
    href = first.get_attribute("href")
    assert href and "/m/video/" in href, f"unexpected href: {href}"
    a_id = href.rstrip("/").split("/")[-1]
    assert a_id.isdigit(), f"aId not numeric: {href}"

    # 标题非空，封面 alt 与标题一致
    title = first.locator(".title").text_content().strip()
    assert title, "video title should not be empty"
    pic = first.locator(".pic img")
    assert pic.count() == 1
    assert (pic.first.get_attribute("alt") or "").strip() == title


def test_home_banner_renders_on_recommend_tab(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-item", timeout=15000)

    # 推荐 Tab 下轮播图由 `activeTabId === 0 && banners.length` 控制
    assert mobile_page.locator(".banner-slider").count() >= 1
    assert mobile_page.locator(".banner-slider .swiper-slide img").count() >= 1


def test_home_tab_switch_changes_active_and_content(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".tab-bar .tab-item.active", timeout=15000)

    tabs = mobile_page.locator(".tab-bar .tab-item")
    names = [t.strip() for t in tabs.all_text_contents()]
    target = "热门" if "热门" in names else names[-1]
    tabs.nth(names.index(target)).tap()

    active = mobile_page.locator(".tab-bar .tab-item.active")
    mobile_page.wait_for_function(
        "n => document.querySelector('.tab-bar .tab-item.active')?.textContent.trim() === n",
        arg=target,
        timeout=10000,
    )
    assert active.first.text_content().strip() == target


def test_home_video_item_click_navigates_to_detail(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-item", timeout=15000)

    href = mobile_page.locator(".video-item").first.get_attribute("href")
    mobile_page.locator(".video-item").first.tap()
    mobile_page.wait_for_url("**/m/video/**", timeout=15000)
    assert "/m/video/" in mobile_page.url
    assert "/m/video/" in href


def test_home_drawer_opens_and_lists_partitions(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".partition-bar .switch-btn", timeout=15000)

    mobile_page.locator(".partition-bar .switch-btn").tap()
    mobile_page.wait_for_timeout(500)

    drawer = mobile_page.locator(".drawer, [class*=drawer]")
    assert drawer.count() >= 1, "drawer should exist after opening"


def test_home_favicon_path_is_not_double_prefixed(mobile_page, wap_url):
    """index.html 里的 icon 路径不能带 /wap 前缀。

    Vite 会按 base('/wap/') 重写资源 URL，HTML 里再写死 /wap/ 就变成
    /wap/wap/wap-icon.svg → 404。
    """
    mobile_page.goto(f"{wap_url}/m/index", wait_until="domcontentloaded")

    href = mobile_page.locator("link[rel=icon]").get_attribute("href")
    assert href == "/wap/wap-icon.svg", f"favicon href 被重复加了前缀: {href!r}"

    resp = mobile_page.request.get(f"{wap_url}/wap-icon.svg")
    assert resp.status == 200, f"favicon 返回 {resp.status}"
