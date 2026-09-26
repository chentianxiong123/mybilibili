"""WAP 私信：会话列表 / 打开会话 / 发送。

这三处都是**静默**失败——不报错、不白屏，只是什么都不对：
- `getConversations` 不做 snake_case → camelCase 换算：列表的昵称、头像、
  最后一条消息全部渲染成空；`openConversation` 于是退到 `item.id`（会话 id）
  当 `targetUserId`，点进去打开的是另一场会话，表现为「空白聊天」。
- `Chat.send` 发的是 `receiver_id`，后端 `handleSend` 只认 `receiverId`
  → 400 被 try/catch 吞掉，`msg.id` 取不到就不 push，界面看着像「点了没反应」。
- `isMine` 读 `localStorage['user_id']`，而全项目从没写过这个键
  → 自己发的消息渲染成对方的气泡（左对齐 + 默认头像）。
"""
import time

CHAT_INPUT = "input[placeholder*='聊聊']"


def _first_conversation(page) -> dict:
    """以后端返回为准，避免和渲染层用同一份（可能有 bug 的）数据自我印证。"""
    resp = page.request.get("http://localhost/api/v1/message/conversations")
    assert resp.status == 200, f"conversations 请求失败: {resp.status}"
    data = resp.json().get("data") or []
    assert data, "测试账号需要至少一个会话"
    return data[0]


def test_conversation_list_renders_name_avatar_and_last_message(wap_session, wap_url):
    page = wap_session
    page.goto(f"{wap_url}/m/message", wait_until="networkidle")
    page.wait_for_selector(".chat-item", timeout=15000)

    expected = _first_conversation(page)
    first = page.locator(".chat-item").first

    name = first.locator(".name").inner_text().strip()
    last = first.locator(".last-msg").inner_text().strip()
    avatar = first.locator("img").first.get_attribute("src") or ""

    assert expected["target_user_name"] in name, (
        f"昵称渲染为空或不符: {name!r}（API 给的是 {expected['target_user_name']!r}，"
        "说明 snake_case 没换算成 camelCase）"
    )
    assert last and last != "暂无新消息", f"最后一条消息没渲染出来: {last!r}"
    assert expected["target_user_avatar"] in avatar, (
        f"头像回退到默认图: {avatar!r}（API 给的是 {expected['target_user_avatar']!r}）"
    )
    assert first.locator(".time").inner_text().strip(), "时间没渲染出来"


def test_conversation_click_opens_that_conversation(wap_session, wap_url):
    page = wap_session
    page.goto(f"{wap_url}/m/message", wait_until="networkidle")
    page.wait_for_selector(".chat-item", timeout=15000)

    expected = _first_conversation(page)
    target = expected["target_user_id"]

    page.locator(".chat-item").first.click()
    page.wait_for_timeout(1500)

    assert f"/m/message/chat/{target}" in page.url, (
        f"点开会话跳到了 {page.url}，应为 .../chat/{target}——"
        "targetUserId 取错会打开另一场会话"
    )
    title = page.locator(".chat-topbar h1").inner_text().strip()
    assert title and title != "私信", (
        f"聊天页标题退化成占位符 {title!r}，说明 query.name 没传过来"
    )
    # 页面应当加载完毕（loading 占位符消失）
    page.wait_for_selector(".loading", state="detached", timeout=15000)


def test_send_message_renders_as_own_and_persists(wap_session, wap_url):
    page = wap_session
    conv = _first_conversation(page)
    target = conv["target_user_id"]
    text = f"e2e-send-{int(time.time())}"

    page.goto(f"{wap_url}/m/message/chat/{target}", wait_until="networkidle")
    page.wait_for_selector(CHAT_INPUT, timeout=15000)

    box = page.locator(CHAT_INPUT)
    box.fill(text)
    box.press("Enter")

    try:
        # 发出去的消息必须立刻以「自己的气泡」出现（右对齐 .mine）
        page.locator(".message-row.mine", has_text=text).first.wait_for(
            timeout=10000
        )
        assert page.locator(".message-row.mine", has_text=text).count() >= 1, (
            "发送后消息没进自己的气泡——要么 receiverId 字段名不对，要么 isMine 判错"
        )
        assert page.locator(".message-row.other", has_text=text).count() == 0, (
            "自己发的消息被渲染成对方的气泡"
        )

        # 刷新后还在，才算真的写进了服务端
        page.reload(wait_until="networkidle")
        page.wait_for_timeout(1500)
        assert page.locator(".message-row.mine", has_text=text).count() >= 1, (
            "刷新后消息消失——400 被 catch 吞掉了，其实没发出去"
        )
    finally:
        _delete_messages_with_content(page, conv["id"], text)


def _delete_messages_with_content(page, conversation_id: int, text: str) -> None:
    """清掉用例自己造的私信，别把 e2e 数据留在库里。"""
    resp = page.request.get(
        f"http://localhost/api/v1/message/conversations/{conversation_id}/messages",
        params={"page": 1, "page_size": 50},
    )
    for m in resp.json().get("data") or []:
        if m.get("content") == text:
            page.request.delete(f"http://localhost/api/v1/message/{m['id']}")
