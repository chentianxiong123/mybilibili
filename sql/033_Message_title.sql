-- 系统通知广播标题：messages 加 title 列（正文之外的可选短标题，B站式"标题+时间 / 正文"）
-- 普通私信/回复/点赞不带标题，该列留 NULL；仅 msgType=5 系统通知广播使用。
ALTER TABLE messages ADD COLUMN IF NOT EXISTS title TEXT;
