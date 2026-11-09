export interface TourStep {
  id: string;
  target: string | null;
  title: string;
  description: string;
  tips?: string[];
  position: "center" | "top" | "bottom" | "left" | "right";
}

export const TOUR_STEPS: TourStep[] = [
  {
    id: "welcome",
    target: null,
    title: "欢迎使用 mybilibili 剪辑工作室",
    description: "快速熟悉剪辑器的主要区域",
    position: "center",
  },
  {
    id: "assets",
    target: "[data-tour='assets']",
    title: "素材面板",
    description: "导入素材、生成 AI 内容、添加形状、贴纸与自定义 SVG。",
    tips: [
      "拖拽导入视频、音频、图片",
      "AI 生成页签可生成图片与背景",
      "支持形状与自定义 SVG 导入",
      "支持贴纸、背景与叠加层",
    ],
    position: "right",
  },
  {
    id: "timeline",
    target: "[data-tour='timeline']",
    title: "时间线",
    description: "排列并剪辑片段。拖动片段可移动，拖动边缘可裁剪。",
    tips: ["按 S 分割片段", "空格播放或暂停", "滚轮缩放时间线"],
    position: "top",
  },
  {
    id: "preview",
    target: "[data-tour='preview']",
    title: "预览",
    description: "剪辑时实时查看画面效果。",
    tips: [
      "方向键逐帧移动",
      "点击可定位播放位置",
      "支持全屏预览",
    ],
    position: "left",
  },
  {
    id: "inspector",
    target: "[data-tour='inspector']",
    title: "检查器",
    description:
      "选择片段后查看属性，可添加效果、调色与动画。",
    tips: [
      "变换、效果与调色",
      "任意属性可添加关键帧",
      "AI 辅助工具",
    ],
    position: "left",
  },
  {
    id: "complete",
    target: null,
    title: "准备完成",
    description: "开始剪辑吧。随时按 ? 查看快捷键。",
    position: "center",
  },
];

export const ONBOARDING_KEY = "openreel-onboarding-complete";
