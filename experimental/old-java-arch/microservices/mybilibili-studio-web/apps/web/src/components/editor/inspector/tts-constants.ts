import type { ElevenLabsModel, Voice } from "./tts-types";

export const TTS_PROVIDERS = [
  { id: "piper" as const, label: "Piper（免费）", description: "内置开源语音服务" },
  { id: "elevenlabs" as const, label: "ElevenLabs", description: "高质量 AI 语音服务" },
];

export const FALLBACK_MODELS: ElevenLabsModel[] = [
  { model_id: "eleven_v3", name: "Eleven v3", description: "ElevenLabs 最新模型", can_do_text_to_speech: true, languages: [] },
];

export const PIPER_VOICES: Voice[] = [
  { id: "amy", name: "艾米", gender: "female", language: "en-US" },
  { id: "ryan", name: "瑞安", gender: "male", language: "en-US" },
];

export const ENHANCE_SYSTEM_PROMPT = `You are a professional voice director transforming text into expressive, emotionally rich scripts for ElevenLabs v3 TTS. Your goal is to turn narration into performance.

Analyze the input for speaker intent, emotional arc, subtext, physical state, relationship dynamics, pacing needs, and environmental context.

Use the 4-Layer System: Delivery (HOW), Tone (emotional color), Texture (voice quality), Subtext (what's beneath). Layer 1-3 tags per emotional beat.

Available tags include: [authoritatively], [hesitantly], [breathlessly], [whispered], [softly], [sighs], [chuckles], [laughs], [sobs], [gasps], [pause:200ms], and many more.

Rules: Do not alter original text. Do not over-tag. Do not use conflicting tags. Output ONLY enhanced text with tags, no explanations.`;
