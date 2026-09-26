"""WAP 直播：直播间详情 / 直播列表。"""


def test_live_room_renders_without_pageerror(mobile_page, wap_url):
    """直播间必须能打开。

    Room.vue 是纯 JS 的 `<script setup>`（没有 lang="ts"），里面写了
    `ref<any>(null)` 会直接 ReferenceError，整个组件挂掉 → 只剩底部 tabbar 的白屏。
    """
    errors = []
    mobile_page.on("pageerror", lambda e: errors.append(str(e)))

    mobile_page.goto(f"{wap_url}/m/live/1", wait_until="domcontentloaded")
    mobile_page.wait_for_timeout(2500)

    assert not errors, f"/m/live/1 抛出未捕获异常: {errors}"
    body = mobile_page.locator("body").inner_text()
    assert "直播间" in body, f"直播间没渲染出来，页面文本: {body[:200]!r}"


def test_live_list_has_no_4xx(mobile_page, wap_url):
    """直播列表不能有 4xx 资源（封面兜底图 nocontent.png 曾被写成相对路径，
    动态 `:src` 不经 Vite 转换，拼出 /wap/assets/nocontent.png → 404）。"""
    bad = []
    mobile_page.on(
        "response", lambda r: bad.append(f"{r.status} {r.url}") if r.status >= 400 else None
    )

    mobile_page.goto(f"{wap_url}/m/live/list", wait_until="domcontentloaded")
    mobile_page.wait_for_timeout(3000)

    assert not bad, f"直播列表出现 4xx/5xx 资源: {sorted(set(bad))}"
