import type { ProjectSettings } from "./project";
import type { Timeline, Track, Clip, Subtitle } from "./timeline";

export interface Template {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly category: TemplateCategory;
  readonly thumbnailUrl: string | null;
  readonly previewUrl: string | null;
  readonly createdAt: number;
  readonly modifiedAt: number;
  readonly settings: ProjectSettings;
  readonly timeline: TemplateTimeline;
  readonly placeholders: TemplatePlaceholder[];
  readonly tags: string[];
  readonly author?: string;
  readonly version: string;
}

export type TemplateCategory =
  | "social-media"
  | "youtube"
  | "tiktok"
  | "instagram"
  | "business"
  | "personal"
  | "slideshow"
  | "intro-outro"
  | "lower-third"
  | "custom";

export interface TemplateTimeline extends Omit<
  Timeline,
  "tracks" | "subtitles"
> {
  readonly tracks: TemplateTrack[];
  readonly subtitles: TemplateSubtitle[];
}

export interface TemplateTrack extends Omit<Track, "clips"> {
  readonly clips: TemplateClip[];
}

export interface TemplateClip extends Clip {
  readonly placeholderId?: string;
  readonly isPlaceholder: boolean;
}

export interface TemplateSubtitle extends Subtitle {
  readonly placeholderId?: string;
  readonly isPlaceholder: boolean;
}

export type PlaceholderType = "text" | "media" | "subtitle";

export interface TemplatePlaceholder {
  readonly id: string;
  readonly type: PlaceholderType;
  readonly label: string;
  readonly description?: string;
  readonly required: boolean;
  readonly defaultValue?: string;
  readonly constraints?: PlaceholderConstraints;
}

export interface PlaceholderConstraints {
  readonly minDuration?: number;
  readonly maxDuration?: number;
  readonly aspectRatio?: number;
  readonly mediaTypes?: Array<"video" | "audio" | "image">;
  readonly maxLength?: number;
}

export interface TemplateReplacements {
  readonly [placeholderId: string]: PlaceholderReplacement;
}

export interface PlaceholderReplacement {
  readonly type: PlaceholderType;
  readonly value: string;
  readonly mediaBlob?: Blob;
}

export interface TemplateSummary {
  readonly id: string;
  readonly name: string;
  readonly category: TemplateCategory;
  readonly thumbnailUrl: string | null;
  readonly placeholderCount: number;
  readonly duration: number;
}

export const TEMPLATE_CATEGORIES: Array<{
  id: TemplateCategory;
  name: string;
  icon: string;
}> = [
  { id: "social-media", name: "社交内容", icon: "share" },
  { id: "youtube", name: "横屏投稿", icon: "youtube" },
  { id: "tiktok", name: "竖屏短视频", icon: "smartphone" },
  { id: "instagram", name: "图文短片", icon: "instagram" },
  { id: "business", name: "商业展示", icon: "briefcase" },
  { id: "personal", name: "个人创作", icon: "user" },
  { id: "slideshow", name: "幻灯片", icon: "images" },
  { id: "intro-outro", name: "片头片尾", icon: "play" },
  { id: "lower-third", name: "字幕条", icon: "subtitles" },
  { id: "custom", name: "自定义", icon: "folder" },
];
