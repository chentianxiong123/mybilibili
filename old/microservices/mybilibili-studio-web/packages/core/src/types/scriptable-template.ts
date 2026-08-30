import type { ProjectSettings } from "./project";
import type { Timeline } from "./timeline";
import type {
  TemplateCategory,
  TemplatePlaceholder,
  PlaceholderConstraints,
} from "./template";

export type ExtendedPlaceholderType =
  | "text"
  | "media"
  | "subtitle"
  | "shape"
  | "effect"
  | "transform"
  | "keyframe"
  | "color"
  | "number"
  | "boolean"
  | "audio"
  | "style"
  | "font"
  | "animation";

export interface PlaceholderTarget {
  readonly clipId?: string;
  readonly trackId?: string;
  readonly effectId?: string;
  readonly keyframeId?: string;
  readonly property: string;
}

export interface PlaceholderUIHints {
  readonly inputType:
    | "text"
    | "textarea"
    | "slider"
    | "color"
    | "select"
    | "toggle"
    | "media-picker"
    | "font-picker"
    | "animation-picker";
  readonly group?: string;
  readonly order?: number;
  readonly advanced?: boolean;
  readonly previewable?: boolean;
  readonly options?: Array<{ value: string; label: string }>;
}

export interface ExtendedPlaceholderConstraints extends PlaceholderConstraints {
  readonly min?: number;
  readonly max?: number;
  readonly step?: number;
  readonly pattern?: string;
  readonly allowedValues?: string[];
  readonly allowedFonts?: string[];
  readonly allowedAnimations?: string[];
}

export interface ExtendedPlaceholder {
  readonly id: string;
  readonly type: ExtendedPlaceholderType;
  readonly label: string;
  readonly description?: string;
  readonly required: boolean;
  readonly defaultValue: unknown;
  readonly targets: PlaceholderTarget[];
  readonly constraints?: ExtendedPlaceholderConstraints;
  readonly uiHints?: PlaceholderUIHints;
}

export type SocialMediaCategory =
  | "tiktok"
  | "instagram-reels"
  | "instagram-stories"
  | "instagram-post"
  | "youtube-shorts"
  | "youtube-video"
  | "facebook"
  | "twitter"
  | "linkedin"
  | "pinterest"
  | "intro"
  | "outro"
  | "promo"
  | "lower-third"
  | "slideshow"
  | "custom";

export interface SocialMediaPreset {
  readonly width: number;
  readonly height: number;
  readonly frameRate?: number;
  readonly maxDuration?: number;
  readonly recommendedDuration?: number;
  readonly safeZone?: {
    readonly top: number;
    readonly bottom: number;
    readonly left: number;
    readonly right: number;
  };
}

export const SOCIAL_MEDIA_PRESETS: Record<
  SocialMediaCategory,
  SocialMediaPreset
> = {
  tiktok: {
    width: 1080,
    height: 1920,
    frameRate: 30,
    maxDuration: 180,
    recommendedDuration: 15,
    safeZone: { top: 150, bottom: 300, left: 40, right: 40 },
  },
  "instagram-reels": {
    width: 1080,
    height: 1920,
    frameRate: 30,
    maxDuration: 90,
    recommendedDuration: 30,
    safeZone: { top: 200, bottom: 280, left: 40, right: 40 },
  },
  "instagram-stories": {
    width: 1080,
    height: 1920,
    frameRate: 30,
    maxDuration: 15,
    recommendedDuration: 10,
    safeZone: { top: 200, bottom: 200, left: 40, right: 40 },
  },
  "instagram-post": {
    width: 1080,
    height: 1080,
    frameRate: 30,
    maxDuration: 60,
    recommendedDuration: 30,
  },
  "youtube-shorts": {
    width: 1080,
    height: 1920,
    frameRate: 30,
    maxDuration: 60,
    recommendedDuration: 30,
    safeZone: { top: 100, bottom: 200, left: 40, right: 40 },
  },
  "youtube-video": {
    width: 1920,
    height: 1080,
    frameRate: 30,
  },
  facebook: {
    width: 1080,
    height: 1080,
    frameRate: 30,
    maxDuration: 240,
    recommendedDuration: 60,
  },
  twitter: {
    width: 1280,
    height: 720,
    frameRate: 30,
    maxDuration: 140,
    recommendedDuration: 45,
  },
  linkedin: {
    width: 1920,
    height: 1080,
    frameRate: 30,
    maxDuration: 600,
    recommendedDuration: 90,
  },
  pinterest: {
    width: 1000,
    height: 1500,
    frameRate: 30,
    maxDuration: 60,
    recommendedDuration: 15,
  },
  intro: {
    width: 1920,
    height: 1080,
    frameRate: 30,
    recommendedDuration: 5,
  },
  outro: {
    width: 1920,
    height: 1080,
    frameRate: 30,
    recommendedDuration: 10,
  },
  promo: {
    width: 1920,
    height: 1080,
    frameRate: 30,
    recommendedDuration: 30,
  },
  "lower-third": {
    width: 1920,
    height: 1080,
    frameRate: 30,
    recommendedDuration: 5,
  },
  slideshow: {
    width: 1920,
    height: 1080,
    frameRate: 30,
  },
  custom: {
    width: 1920,
    height: 1080,
    frameRate: 30,
  },
};

export interface TemplateScene {
  readonly id: string;
  readonly label: string;
  readonly startTime: number;
  readonly endTime: number;
  readonly color?: string;
}

export interface ScriptableTemplate {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly category: TemplateCategory;
  readonly socialCategory?: SocialMediaCategory;
  readonly thumbnailUrl: string | null;
  readonly previewUrl: string | null;
  readonly previewVideoUrl?: string | null;
  readonly createdAt: number;
  readonly modifiedAt: number;
  readonly settings: ProjectSettings;
  readonly timeline: Timeline;
  readonly placeholders: ExtendedPlaceholder[];
  readonly scenes?: TemplateScene[];
  readonly tags: string[];
  readonly author?: string;
  readonly version: string;
  readonly featured?: boolean;
  readonly premium?: boolean;
}

export interface ExtendedPlaceholderReplacement {
  readonly type: ExtendedPlaceholderType;
  readonly value: unknown;
  readonly mediaBlob?: Blob;
}

export interface ScriptableTemplateReplacements {
  readonly [placeholderId: string]: ExtendedPlaceholderReplacement;
}

export interface TemplateValidationError {
  readonly placeholderId: string;
  readonly message: string;
  readonly type: "missing" | "invalid" | "constraint";
}

export interface TemplateApplicationResult {
  readonly success: boolean;
  readonly errors: TemplateValidationError[];
  readonly warnings: string[];
}

export interface PlaceholderGroup {
  readonly id: string;
  readonly label: string;
  readonly description?: string;
  readonly placeholderIds: string[];
  readonly collapsed?: boolean;
}

export function isExtendedPlaceholder(
  placeholder: TemplatePlaceholder | ExtendedPlaceholder,
): placeholder is ExtendedPlaceholder {
  return (
    "targets" in placeholder &&
    Array.isArray((placeholder as ExtendedPlaceholder).targets)
  );
}

export function convertLegacyPlaceholder(
  placeholder: TemplatePlaceholder,
): ExtendedPlaceholder {
  return {
    id: placeholder.id,
    type: placeholder.type,
    label: placeholder.label,
    description: placeholder.description,
    required: placeholder.required,
    defaultValue: placeholder.defaultValue,
    targets: [],
    constraints: placeholder.constraints,
    uiHints: {
      inputType:
        placeholder.type === "media"
          ? "media-picker"
          : placeholder.type === "text"
            ? "textarea"
            : "text",
    },
  };
}

export function getPresetForCategory(
  category: SocialMediaCategory,
): SocialMediaPreset {
  return SOCIAL_MEDIA_PRESETS[category] || SOCIAL_MEDIA_PRESETS.custom;
}

export function createProjectSettingsFromPreset(
  preset: SocialMediaPreset,
): ProjectSettings {
  return {
    width: preset.width,
    height: preset.height,
    frameRate: preset.frameRate || 30,
    sampleRate: 48000,
    channels: 2,
  };
}

export const SOCIAL_MEDIA_CATEGORY_INFO: Array<{
  id: SocialMediaCategory;
  name: string;
  icon: string;
  platform: string;
}> = [
  { id: "tiktok", name: "短视频竖屏", icon: "smartphone", platform: "竖屏" },
  { id: "instagram-reels", name: "竖屏片段", icon: "film", platform: "竖屏" },
  {
    id: "instagram-stories",
    name: "限时竖屏",
    icon: "clock",
    platform: "竖屏",
  },
  { id: "instagram-post", name: "方形图文", icon: "square", platform: "方形" },
  {
    id: "youtube-shorts",
    name: "竖屏短片",
    icon: "smartphone",
    platform: "竖屏",
  },
  { id: "youtube-video", name: "横屏视频", icon: "monitor", platform: "横屏" },
  { id: "facebook", name: "方形动态", icon: "users", platform: "方形" },
  { id: "twitter", name: "横屏动态", icon: "at-sign", platform: "横屏" },
  { id: "linkedin", name: "横屏演示", icon: "briefcase", platform: "横屏" },
  {
    id: "pinterest",
    name: "长图",
    icon: "bookmark",
    platform: "竖屏",
  },
  { id: "intro", name: "片头", icon: "play", platform: "通用" },
  { id: "outro", name: "片尾", icon: "square", platform: "通用" },
  { id: "promo", name: "宣传片", icon: "megaphone", platform: "通用" },
  { id: "lower-third", name: "字幕条", icon: "type", platform: "通用" },
  { id: "slideshow", name: "幻灯片", icon: "images", platform: "通用" },
  { id: "custom", name: "自定义", icon: "settings", platform: "自定义" },
];
