-- 032_System_sender.sql
-- 系统通知需要一个真实 sender：messages.sender_id 有外键指向 users(id)，
-- 管理员广播原先传 sender_id=0，插入被外键拒绝且错误被吞掉，
-- 接口回 broadcast_sent 但一条都没发出去。
-- 建一个固定系统账号（username 唯一，重复执行安全），广播统一用它当 sender。

INSERT INTO users (username, nickname, password, level)
VALUES ('system', '系统通知', '', 6)
ON CONFLICT (username) DO NOTHING;