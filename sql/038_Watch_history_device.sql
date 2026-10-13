-- 观看历史增加客户端类型（设备维度），用于账号诊断「观看途径」分布
-- 取值：mobile 移动端 / pc PC / h5 h5及小程序 / tv 云视听小电视 / other 其他
ALTER TABLE watch_history ADD COLUMN IF NOT EXISTS client_type VARCHAR(16) NOT NULL DEFAULT 'pc';