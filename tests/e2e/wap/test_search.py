"""WAP 搜索：热搜榜 / 搜索流程 / 搜索历史的持久化与清空。"""
import json


def test_search_page_renders_topbar_and_hotwords(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".search-topbar input", timeout=15000)

    assert mobile_page.locator(".search-topbar").count() == 1
    assert mobile_page.locator(".search-topbar input").count() == 1
    assert mobile_page.locator(".search-topbar .search-btn").count() == 1
    assert mobile_page.locator(".search-topbar .back-btn").count() == 1

    # 热搜榜
    assert mobile_page.locator(".hot-section h2").text_content().strip() == "bilibili热搜"
    hot_btns = mobile_page.locator(".hot-grid button")
    # 热搜榜最多取前 10 条，实际条数取决于 Redis 里有多少有效关键词
    assert 1 <= hot_btns.count() <= 10


def test_hotword_buttons_render_plain_text_not_raw_json(mobile_page, wap_url):
    """
    热搜条目必须是纯关键词文本。

    api/search.ts 的 getHotwords 用 `item.keyword || item` 兜底，当后端某条
    返回 keyword 为空串时会退化成渲染整个对象（表现为 JSON 文本）。这条用例
    锁住该行为。
    """
    mobile_page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".hot-grid button", timeout=15000)

    texts = [t.strip() for t in mobile_page.locator(".hot-grid button").all_text_contents()]
    assert 1 <= len(texts) <= 10
    for i, t in enumerate(texts):
        assert t, f"hotword #{i} is empty"
        assert "{" not in t and "}" not in t, (
            f"hotword #{i} rendered as JSON object instead of keyword: {t!r}"
        )
        # 尾部可能带「新/热/独家」角标，去掉后关键词仍非空
        keyword = t.rstrip("新热独家").strip()
        assert keyword, f"hotword #{i} keyword empty after stripping badge: {t!r}"


def test_search_history_starts_empty(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".history-section", timeout=15000)

    assert mobile_page.locator(".empty-text").count() == 1
    assert "暂无搜索历史" in mobile_page.locator(".empty-text").text_content()


def test_search_flow_returns_video_results(clean_history, wap_url):
    page = clean_history
    page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    page.wait_for_selector(".search-topbar input", timeout=15000)

    page.fill(".search-topbar input", "AI")
    page.wait_for_timeout(600)
    page.press(".search-topbar input", "Enter")
    page.wait_for_selector(".video-item", timeout=15000)

    items = page.locator(".video-item")
    assert items.count() > 0, "search should return at least one video"

    first = items.first
    href = first.get_attribute("href")
    assert href and "/m/video/" in href, f"unexpected result href: {href}"
    assert first.locator(".title").text_content().strip()


def test_search_history_persists_after_search(clean_history, wap_url):
    page = clean_history
    page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    page.wait_for_selector(".search-topbar input", timeout=15000)

    page.fill(".search-topbar input", "OpenClaw")
    page.press(".search-topbar input", "Enter")
    page.wait_for_selector(".video-item", timeout=15000)

    # 退出结果态回到搜索首页，历史里应出现刚搜的词
    page.locator(".search-topbar .back-btn").tap()
    page.wait_for_selector(".history-section", timeout=10000)

    history_btns = page.locator(".history-section .pill-grid button")
    assert history_btns.count() >= 1, "history should contain the searched keyword"
    shown = [t.strip() for t in history_btns.all_text_contents()]
    assert "OpenClaw" in shown, f"history missing keyword, got {shown}"

    # 键名 = PREFIX('wap:') + K.searchHistory('search:history')，见 utils/storage_layer.ts
    raw = page.evaluate("() => localStorage.getItem('wap:search:history')")
    assert raw, "wap:search:history should be written"
    assert "OpenClaw" in json.loads(raw)


def test_search_history_clear_button_empties_it(clean_history, wap_url):
    page = clean_history
    page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    page.wait_for_selector(".search-topbar input", timeout=15000)

    page.fill(".search-topbar input", "test")
    page.press(".search-topbar input", "Enter")
    page.wait_for_selector(".video-item", timeout=15000)
    page.locator(".search-topbar .back-btn").tap()
    page.wait_for_selector(".history-section .pill-grid button", timeout=10000)

    page.locator(".history-section .icon-only").tap()
    page.wait_for_selector(".empty-text", timeout=5000)

    assert page.locator(".history-section .pill-grid").count() == 0
    assert page.evaluate("() => localStorage.getItem('wap:search:history')") in (None, "[]")


def test_search_hotword_click_triggers_search(clean_history, wap_url):
    page = clean_history
    page.goto(f"{wap_url}/m/search", wait_until="domcontentloaded")
    page.wait_for_selector(".hot-grid button", timeout=15000)

    page.locator(".hot-grid button").nth(1).tap()
    page.wait_for_timeout(2500)

    # 点热搜后进入结果态（keyword 非空 → 渲染 Result），或无结果也应离开搜索首页
    left_home = page.locator(".search-home").count() == 0
    has_results = page.locator(".video-item").count() > 0
    assert left_home or has_results, "clicking a hotword should start a search"
