"""WAP 端 e2e 共享 fixture。

复用父级 tests/e2e/conftest.py 的 session 级 browser（系统 Chrome），
这里只补移动端 context 和 WAP 专属的登录态。
"""
import pytest

# WAP 挂在 traefik PathPrefix(`/wap/`) 下，vite BASE_URL = /wap/
WAP_URL = "http://localhost/wap"

MOBILE_UA = (
    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) "
    "AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
)


@pytest.fixture(scope="session")
def wap_url() -> str:
    return WAP_URL


def _new_mobile_context(browser, base_url: str):
    ctx = browser.new_context(
        viewport={"width": 375, "height": 667},
        is_mobile=True,
        has_touch=True,
        device_scale_factor=2,
        user_agent=MOBILE_UA,
        base_url=base_url,
    )
    return ctx


@pytest.fixture()
def mobile_page(browser, wap_url):
    """未登录的移动端 page，每个用例一个干净 context。"""
    ctx = _new_mobile_context(browser, wap_url)
    page = ctx.new_page()
    page.set_default_timeout(15000)
    yield page
    ctx.close()


@pytest.fixture()
def wap_session(browser, wap_url):
    """
    已登录 string/123456 的 WAP page。

    登录态 = HttpOnly cookie（凭证） + localStorage `wap:user`（展示信息，
    同时是前端判断登录态的唯一信号，见 mobile/wap/src/utils/session.ts）。
    用例结束后清掉 localStorage，避免污染后续用例。
    """
    ctx = _new_mobile_context(browser, wap_url)
    page = ctx.new_page()
    page.set_default_timeout(15000)

    resp = page.request.post(
        "http://localhost/api/v1/user/login",
        data={"username": "string", "password": "123456"},
    )
    assert resp.status == 200, f"login failed: {resp.status}"
    body = resp.json()
    assert body.get("code") == 200, f"login error: {body}"
    data = body["data"]

    page.goto(f"{WAP_URL}/m/index", wait_until="domcontentloaded")
    page.evaluate(
        "(u) => localStorage.setItem('wap:user', JSON.stringify(u))",
        {
            "id": data["id"],
            "nickname": data["nickname"],
            "username": "string",
            "avatar": data.get("avatar", ""),
        },
    )
    page.reload(wait_until="domcontentloaded")

    yield page

    try:
        page.evaluate("() => localStorage.removeItem('wap:user')")
    except Exception:
        pass
    ctx.close()


@pytest.fixture()
def clean_history(mobile_page):
    """
    搜索历史是纯 localStorage（键 `wap:search:history`，见 session/storage_layer），
    这个 fixture 在用例前后清干净，保证断言确定性。
    必须先导航到同源页面——about:blank 上访问 localStorage 会抛 SecurityError。
    只给真正断言历史的用例显式引用，不做成 autouse——否则所有 wap 用例都会
    被迫多建一个 page。
    """
    mobile_page.goto(f"{WAP_URL}/m/search", wait_until="domcontentloaded")
    mobile_page.evaluate("() => localStorage.removeItem('wap:search:history')")
    yield mobile_page
    try:
        mobile_page.evaluate("() => localStorage.removeItem('wap:search:history')")
    except Exception:
        pass
