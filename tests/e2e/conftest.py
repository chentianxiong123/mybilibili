"""共享 pytest fixture：复用系统 Chrome（与 MCP playwright 共用同一份浏览器二进制）。"""
import pytest
from playwright.sync_api import sync_playwright


@pytest.fixture(scope="session")
def base_url() -> str:
    return "http://localhost"


def pytest_configure(config):
    """告诉 pytest-playwright 用系统 Chrome 而不是下载 chromium。"""
    # pytest-playwright 默认会试图启动自带的 chromium。
    # 通过 conftest 配置 channel="chrome" 强制走系统 Chrome，避免下载。
    pass


@pytest.fixture(scope="session")
def browser():
    """全局共享一个 Chrome 实例，所有用例复用。"""
    with sync_playwright() as p:
        browser = p.chromium.launch(
            channel="chrome",
            headless=True,
            args=["--no-sandbox", "--disable-dev-shm-usage"],
        )
        yield browser
        browser.close()


@pytest.fixture()
def page(browser, base_url):
    """每个用例一个干净 page（已登录态由各用例自行处理）。"""
    ctx = browser.new_context(viewport={"width": 1280, "height": 800})
    page = ctx.new_page()
    page.set_default_timeout(10000)
    yield page
    ctx.close()


@pytest.fixture(scope="session")
def logged_in_page(browser, base_url):
    """已登录 string/123456 的 page fixture —— pytest-playwright 默认不提供 session-scoped page，所以单独写。"""
    ctx = browser.new_context(viewport={"width": 1280, "height": 800})
    page = ctx.new_page()
    page.set_default_timeout(10000)

    # 直接走 /api/v1/user/login 拿 token，写入 localStorage，触发一次刷新让 store 加载
    resp = page.request.post(
        f"{base_url}/api/v1/user/login",
        data={"username": "string", "password": "123456"},
    )
    assert resp.status == 200, f"login failed: {resp.status}"
    body = resp.json()
    assert body.get("code") == 200, f"login error: {body}"
    token = body["data"]["token"]
    user = {
        "id": body["data"]["id"],
        "username": "string",
        "nickname": body["data"]["nickname"],
        "avatar": body["data"]["avatar"],
    }

    page.goto(base_url)
    page.evaluate(
        "(args) => { localStorage.setItem('teri_token', args.t); localStorage.setItem('user', JSON.stringify(args.u)); }",
        {"t": token, "u": user},
    )
    page.goto(base_url)
    page.wait_for_load_state("networkidle")

    yield page
    ctx.close()