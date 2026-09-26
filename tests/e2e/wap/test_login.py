"""WAP 登录：表单校验 / 登录失败提示 / 登录成功后的会话落盘。"""


def test_login_page_shows_form_fields(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-form input", timeout=15000)

    inputs = mobile_page.locator(".login-form input")
    assert inputs.count() == 2
    assert inputs.nth(0).get_attribute("placeholder") == "用户名"
    assert inputs.nth(1).get_attribute("placeholder") == "密码"
    assert mobile_page.locator(".error-msg").count() == 0


def test_login_with_empty_fields_shows_validation_error(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-btn", timeout=15000)

    mobile_page.locator(".login-btn").tap()
    mobile_page.wait_for_selector(".error-msg", timeout=5000)

    assert "请输入用户名和密码" in mobile_page.locator(".error-msg").text_content()
    # 校验失败不应跳转
    assert "/m/login" in mobile_page.url


def test_login_with_wrong_password_shows_error(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-form input", timeout=15000)

    inputs = mobile_page.locator(".login-form input")
    inputs.nth(0).fill("string")
    inputs.nth(1).fill("definitely-wrong")
    mobile_page.locator(".login-btn").tap()

    mobile_page.wait_for_selector(".error-msg", timeout=10000)
    msg = mobile_page.locator(".error-msg").text_content().strip()
    assert msg and msg != "请输入用户名和密码", f"expected a login-failure message, got {msg!r}"
    assert "/m/login" in mobile_page.url


def test_login_success_redirects_and_persists_session(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-form input", timeout=15000)

    inputs = mobile_page.locator(".login-form input")
    inputs.nth(0).fill("string")
    inputs.nth(1).fill("123456")
    mobile_page.locator(".login-btn").tap()

    mobile_page.wait_for_url("**/m/index**", timeout=20000)
    assert "/m/index" in mobile_page.url

    # 会话必须写进 localStorage（前缀 wap:，键 user）
    raw = mobile_page.evaluate("() => localStorage.getItem('wap:user')")
    assert raw, "wap:user should be written after successful login"
    user = mobile_page.evaluate(
        "() => JSON.parse(localStorage.getItem('wap:user') || 'null')"
    )
    assert user, "wap:user should be valid JSON"
    assert user.get("id") == 4
    assert user.get("nickname"), "nickname should be stored for header rendering"


def test_login_success_shows_avatar_in_header(mobile_page, wap_url):
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-form input", timeout=15000)

    inputs = mobile_page.locator(".login-form input")
    inputs.nth(0).fill("string")
    inputs.nth(1).fill("123456")
    mobile_page.locator(".login-btn").tap()

    mobile_page.wait_for_url("**/m/index**", timeout=20000)
    mobile_page.wait_for_selector(".mobile-header .user-avatar", timeout=10000)

    avatar = mobile_page.locator(".mobile-header .user-avatar img")
    assert avatar.count() >= 1
    src = avatar.first.get_attribute("src") or ""
    assert src, "header avatar img should have a src"


def test_browse_without_login_keeps_no_session(mobile_page, wap_url):
    """「先逛逛」直接回首页，且不产生 wap:user。"""
    mobile_page.goto(f"{wap_url}/m/login", wait_until="domcontentloaded")
    mobile_page.wait_for_selector(".login-links a", timeout=15000)

    mobile_page.locator(".login-links a").tap()
    mobile_page.wait_for_url("**/m/index**", timeout=10000)

    assert "/m/index" in mobile_page.url
    assert mobile_page.evaluate("() => localStorage.getItem('wap:user')") is None
