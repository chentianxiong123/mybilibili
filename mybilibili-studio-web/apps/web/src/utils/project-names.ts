const CITY_NAMES = [
  "投稿", "剪辑", "混剪", "片头", "片尾",
  "直播切片", "教程", "日常", "旅行", "游戏",
  "开箱", "测评", "讲解", "配音", "字幕",
  "封面", "预告", "合集", "素材", "草稿",
  "复盘", "记录", "动态", "短片", "长视频",
];

const ADJECTIVES = [
  "新的", "快速", "精选", "清爽", "高能",
  "正式", "测试", "重制", "精剪", "临时",
  "完整", "竖屏", "横屏", "方形", "发布用",
];

export function generateProjectName(): string {
  const adjective = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)];
  const city = CITY_NAMES[Math.floor(Math.random() * CITY_NAMES.length)];
  return `${adjective} ${city}`;
}

export function generateSimpleProjectName(): string {
  return CITY_NAMES[Math.floor(Math.random() * CITY_NAMES.length)];
}
