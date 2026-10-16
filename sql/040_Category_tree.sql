-- =====================================================================
-- 040_Category_tree.sql
-- categories 表从扁平「学科分类」重构为真 B 站分区树（一级 + 二级）。
--   * id 直接使用真 B 站全局唯一 tid（一级/二级 tid 不重叠）
--   * parent_id=0 为一级分区；否则指向一级分区
--   * 现有 manuscripts.category_id 从旧学科 id 重映射到新分区
--   * creator_settings.default_category_id 归零（旧学科 id 语义已失效）
-- 幂等：基于「categories 表还是旧扁平结构」的前提执行，勿重复运行。
-- =====================================================================

-- 0) 记录旧分类列表备份（撤销用）
CREATE TABLE IF NOT EXISTS _categories_backup_table AS SELECT * FROM categories;

-- 1) 先重映射稿件：旧学科 id → 真 B 站 tid
--    关键词判定优先（标题），其余按旧学科就近归到一级分区
UPDATE manuscripts SET category_id = 95   WHERE category_id = 1;   -- 人工智能 → 科技/数码
UPDATE manuscripts SET category_id = 95   WHERE category_id = 2;   -- 电子     → 科技/数码
UPDATE manuscripts SET category_id = 201  WHERE category_id = 3;   -- 数学     → 知识/科学科普
UPDATE manuscripts SET category_id = 208  WHERE category_id = 4;   -- 英语     → 知识/校园学习
UPDATE manuscripts SET category_id = 234  WHERE category_id = 5;   -- 运动     → 运动(主)
UPDATE manuscripts SET category_id = 124  WHERE category_id = 6;   -- 心理学   → 知识/社科·法律·心理
UPDATE manuscripts SET category_id = 231  WHERE category_id = 7;   -- 软件     → 科技/计算机技术
UPDATE manuscripts SET category_id = 95   WHERE category_id = 8;   -- 硬件     → 科技/数码
UPDATE manuscripts SET category_id = 201  WHERE category_id = 9;   -- 物理     → 知识/科学科普
UPDATE manuscripts SET category_id = 232  WHERE category_id = 10;  -- 机械     → 科技/科工机械
UPDATE manuscripts SET category_id = 188  WHERE category_id = 11;  -- 科技     → 科技(主)
UPDATE manuscripts SET category_id = 124  WHERE category_id = 12;  -- 政治     → 知识/社科
UPDATE manuscripts SET category_id = 228  WHERE category_id = 13;  -- 历史     → 知识/人文历史
UPDATE manuscripts SET category_id = 207  WHERE category_id = 14;  -- 经济     → 知识/财经商业
UPDATE manuscripts SET category_id = 228  WHERE category_id = 15;  -- 文学     → 知识/人文历史
UPDATE manuscripts SET category_id = 124  WHERE category_id = 16;  -- 哲学     → 知识/社科
UPDATE manuscripts SET category_id = 208  WHERE category_id = 17;  -- 教育学   → 知识/校园学习
UPDATE manuscripts SET category_id = 201  WHERE category_id = 18;  -- 医学     → 知识/科学科普
UPDATE manuscripts SET category_id = 209  WHERE category_id = 19;  -- 管理学   → 知识/职业职场
UPDATE manuscripts SET category_id = 229  WHERE category_id = 20;  -- 艺术     → 知识/设计·创意
UPDATE manuscripts SET category_id = 201  WHERE category_id = 21;  -- 地理     → 知识/科学科普
UPDATE manuscripts SET category_id = 208  WHERE category_id = 22;  -- 语言     → 知识/校园学习
UPDATE manuscripts SET category_id = 122  WHERE category_id = 23;  -- 测试     → 知识/野生技术协会

-- 2) 按标题关键词精修（覆盖旧学科粗映射，命中优先）
UPDATE manuscripts SET category_id = 231 WHERE category_id = 188 AND (title LIKE '%程序%' OR title LIKE '%代码%' OR title LIKE '%写一个%' OR title LIKE '%Linux%' OR title LIKE '%Docker%' OR title LIKE '%MySQL%' OR title LIKE '%正则%' OR title LIKE '%命令行%' OR title LIKE '%脚本%' OR title LIKE '%GitHub%' OR title LIKE '%部署%');
UPDATE manuscripts SET category_id = 230 WHERE category_id = 188 AND title LIKE '%工具%';
UPDATE manuscripts SET category_id = 164 WHERE category_id = 188 AND (title LIKE '%跑步%' OR title LIKE '%热身%');
UPDATE manuscripts SET category_id = 208 WHERE category_id = 188 AND title LIKE '%试卷%';

-- 3) 创作者默认分区归零（旧 id 语义失效，避免外键悬挂）
UPDATE creator_settings SET default_category_id = NULL WHERE default_category_id IS NOT NULL AND default_category_id <= 23;

-- 3.1) TRUNCATE 时若有 FK 引用则需先解除（creator_settings 引用 categories，default 已置 NULL）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'creator_settings_ibfk_2') THEN
        ALTER TABLE creator_settings DROP CONSTRAINT IF EXISTS creator_settings_ibfk_2;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_creator_settings_categories') THEN
        ALTER TABLE creator_settings DROP CONSTRAINT IF EXISTS fk_creator_settings_categories;
    END IF;
END $$;

-- 4）数据结构修正（先加列再插数据）
ALTER TABLE categories ADD COLUMN IF NOT EXISTS parent_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);
-- 真 B 站允许同一二级分区名出现在不同一级分区下（如番剧/国创的「资讯」），去掉 name 全局唯一约束
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_name;
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_name_key;

-- 5) 清空并重建 categories 为真 B 站分区树
TRUNCATE TABLE categories RESTART IDENTITY;

INSERT INTO categories (id, name, parent_id, sort_order, created_at, updated_at) VALUES
-- ===== 一级分区（parent_id = 0）=====
(1,    '动画',      0, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(13,   '番剧',      0, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(167,  '国创',      0, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3,    '音乐',      0, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(129,  '舞蹈',      0, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4,    '游戏',      0, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(36,   '知识',      0, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(188,  '科技',      0, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(234,  '运动',      0, 9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(223,  '汽车',      0, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(160,  '生活',      0, 11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(211,  '美食',      0, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(217,  '动物圈',    0, 13, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(119,  '鬼畜',      0, 14, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(155,  '时尚',      0, 15, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(202,  '资讯',      0, 16, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(5,    '娱乐',      0, 17, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(181,  '影视',      0, 18, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(177,  '纪录片',    0, 19, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(23,   '电影',      0, 20, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(11,   '电视剧',    0, 21, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 动画(1) =====
(24, 'MAD·AMV', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(25, 'MMD·3D', 1, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(47, '短片·手书', 1, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(257,'配音', 1, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(210,'手办·模玩', 1, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(86, '特摄', 1, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(253,'动漫杂谈', 1, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(27, '综合', 1, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 番剧(13) =====
(51, '资讯', 13, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(152,'官方延伸', 13, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(32, '完结动画', 13, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(33, '连载动画', 13, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 国创(167) =====
(153,'国产动画', 167, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(168,'国产原创相关', 167, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(169,'布袋戏', 167, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(170,'资讯', 167, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(195,'动态漫·广播剧', 167, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 音乐(3) =====
(28, '原创音乐', 3, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(31, '翻唱', 3, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(30, 'VOCALOID·UTAU', 3, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(59, '演奏', 3, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(193,'MV', 3, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(29, '音乐现场', 3, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(130,'音乐综合', 3, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(243,'乐评盘点', 3, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(244,'音乐教学', 3, 9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 舞蹈(129) =====
(20, '宅舞', 129, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(154,'舞蹈综合', 129, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(156,'舞蹈教程', 129, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(198,'街舞', 129, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(199,'明星舞蹈', 129, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(200,'国风舞蹈', 129, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(255,'手势·网红舞', 129, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 游戏(4) =====
(17, '单机游戏', 4, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(171,'电子竞技', 4, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(172,'手机游戏', 4, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(65, '网络游戏', 4, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(173,'桌游棋牌', 4, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(121,'GMV', 4, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(136,'音游', 4, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(19, 'Mugen', 4, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 知识(36) =====
(201,'科学科普', 36, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(124,'社科·法律·心理', 36, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(228,'人文历史', 36, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(207,'财经商业', 36, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(208,'校园学习', 36, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(209,'职业职场', 36, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(229,'设计·创意', 36, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(122,'野生技术协会', 36, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 科技(188) =====
(95, '数码', 188, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(230,'软件应用', 188, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(231,'计算机技术', 188, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(232,'科工机械', 188, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(233,'极客DIY', 188, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 运动(234) =====
(235,'篮球·足球', 234, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(249,'足球', 234, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(164,'健身', 234, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(236,'竞技体育', 234, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(237,'运动文化', 234, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(238,'运动综合', 234, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 汽车(223) =====
(258,'汽车知识科普', 223, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(245,'赛车', 223, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(246,'改装玩车', 223, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(247,'新能源车', 223, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(248,'房车', 223, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(240,'摩托车', 223, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(227,'购车攻略', 223, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(176,'汽车生活', 223, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 生活(160) =====
(138,'搞笑', 160, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(250,'出行', 160, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(251,'三农', 160, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(239,'家居房产', 160, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(161,'手工', 160, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(162,'绘画', 160, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(21, '日常', 160, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(254,'亲子', 160, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 美食(211) =====
(76, '美食制作', 211, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(212,'美食侦探', 211, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(213,'美食测评', 211, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(214,'田园美食', 211, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(215,'美食记录', 211, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 动物圈(217) =====
(218,'喵星人', 217, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(219,'汪星人', 217, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(220,'动物二创', 217, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(221,'野生动物', 217, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(222,'小宠异宠', 217, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(75, '动物综合', 217, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 鬼畜(119) =====
(22, '鬼畜调教', 119, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(26, '音MAD', 119, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(126,'人力VOCALOID', 119, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(216,'鬼畜剧场', 119, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(127,'教程演示', 119, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 时尚(155) =====
(157,'美妆护肤', 155, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(252,'仿妆cos', 155, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(158,'穿搭', 155, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(159,'时尚潮流', 155, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 资讯(202) =====
(203,'热点', 202, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(204,'环球', 202, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(205,'社会', 202, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(206,'综合', 202, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 娱乐(5) =====
(71, '综艺', 5, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(241,'娱乐杂谈', 5, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(242,'粉丝创作', 5, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(137,'明星综合', 5, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 影视(181) =====
(182,'影视杂谈', 181, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(183,'影视剪辑', 181, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(85, '小剧场', 181, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(184,'预告·资讯', 181, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(256,'短片', 181, 5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 纪录片(177) =====
(37, '人文·历史', 177, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(178,'科学·探索·自然', 177, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(179,'军事', 177, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(180,'社会·美食·旅行', 177, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 电影(23) =====
(147,'华语电影', 23, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(145,'欧美电影', 23, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(146,'日本电影', 23, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(83, '其他国家', 23, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ===== 电视剧(11) =====
(185,'国产剧', 11, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(187,'海外剧', 11, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 5) 外部文件（000_MySQL_data_pg.sql）的 CREATE TABLE 也同步（见该文件）
--    这里仅对已初始化的库做结构修正：
ALTER TABLE categories ADD COLUMN IF NOT EXISTS parent_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);

-- 6) TRUNCATE ... RESTART IDENTITY 会重置 id 序列；而新分区使用真 B 站 tid 作显式 id，
--    恢复序列到最大 tid 之上，避免后续 CreateCategory 生成的 id 与分区 id 冲突。
SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT MAX(id) FROM categories));

-- 7) 恢复 creator_settings 对 categories 的外键引用
ALTER TABLE creator_settings
    ADD CONSTRAINT fk_creator_settings_categories
    FOREIGN KEY (default_category_id) REFERENCES categories(id) ON DELETE SET NULL;