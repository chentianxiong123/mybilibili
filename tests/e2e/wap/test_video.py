"""WAP 视频详情页：播放器 / UP 主信息 / 简介与评论 Tab 切换。"""


def test_video_detail_renders_player_and_up_info(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/video/11", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".video-detail-page", timeout=15000)

    assert mobile_page.locator(".player-wrap").count() >= 1
    assert mobile_page.locator(".top-nav-bar .back-icon").count() == 1
    assert mobile_page.locator(".top-nav-bar .home-icon").count() == 1

    # UP 主信息卡片
    up_name = mobile_page.locator(".up-info-row .up-name")
    assert up_name.count() == 1
    assert up_name.text_content().strip(), "UP name should not be empty"
    assert mobile_page.locator(".up-info-row .up-avatar").count() == 1

    # 关注按钮
    follow = mobile_page.locator(".follow-up-btn")
    assert follow.count() == 1
    assert follow.text_content().strip() in ("+ 关注", "已关注")


def test_video_detail_has_desc_and_comment_tabs(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/video/11", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".tab-title", timeout=15000)

    tabs = mobile_page.locator(".tab-title")
    assert tabs.count() == 2
    labels = [t.strip() for t in tabs.all_text_contents()]
    assert any("简介" in l for l in labels), f"missing 简介 tab: {labels}"
    assert any("评论" in l for l in labels), f"missing 评论 tab: {labels}"

    # 默认简介态
    assert mobile_page.locator(".desc-content-wrapper").count() == 1
    assert mobile_page.locator(".comments-content-wrapper").count() == 0


def test_video_detail_tab_switch_toggles_content(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/video/11", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".desc-content-wrapper", timeout=15000)

    tabs = mobile_page.locator(".tab-title")
    # 第二个 tab 是评论
    tabs.nth(1).tap()
    mobile_page.wait_for_function(
        "() => document.querySelectorAll('.comments-content-wrapper').length > 0",
        timeout=10000,
    )

    assert mobile_page.locator(".comments-content-wrapper").count() >= 1
    assert mobile_page.locator(".desc-content-wrapper").count() == 0
    assert "active" in (tabs.nth(1).get_attribute("class") or "")

    # 切回简介
    tabs.nth(0).tap()
    mobile_page.wait_for_function(
        "() => document.querySelectorAll('.desc-content-wrapper').length > 0",
        timeout=10000,
    )
    assert mobile_page.locator(".desc-content-wrapper").count() == 1
    assert mobile_page.locator(".comments-content-wrapper").count() == 0


def test_video_detail_home_icon_navigates_back_to_index(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/video/11", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".top-nav-bar .home-icon", timeout=15000)

    mobile_page.locator(".top-nav-bar .home-icon").tap()
    mobile_page.wait_for_url("**/m/index**", timeout=10000)
    assert "/m/index" in mobile_page.url


def test_video_detail_vertical_entry_present(mobile_page, wap_url):
    """竖屏入口是 WAP 特有的（/m/vertical/:aId），播放器区应渲染出来。"""
    mobile_page.goto(f"{wap_url}/m/video/11", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".player-wrap", timeout=15000)

    entry = mobile_page.locator(".player-wrap .vertical-entry")
    assert entry.count() == 1
    assert "竖屏" in entry.text_content()
