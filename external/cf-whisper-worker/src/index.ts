/**
 * Cloudflare Workers AI - Whisper 语音转文字 API
 * OpenAI 兼容接口: POST /v1/audio/transcriptions
 * 简单接口: POST /
 */

interface Env {
  AI: any;
  WHISPER_MODEL: string;
}

const DEFAULT_MODEL = "@cf/openai/whisper";

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // 只允许 POST
    if (request.method !== "POST") {
      return new Response(JSON.stringify({ error: "Method not allowed. Use POST." }), {
        status: 405,
        headers: { "Content-Type": "application/json" },
      });
    }

    const url = new URL(request.url);
    const path = url.pathname;

    // OpenAI 兼容端点
    if (path === "/v1/audio/transcriptions") {
      return handleTranscription(request, env, url);
    }

    // 根路径也支持
    if (path === "/" || path === "/transcribe") {
      return handleTranscription(request, env, url);
    }

    return new Response(JSON.stringify({ error: "Not found" }), { status: 404 });
  },
};

async function handleTranscription(
  request: Request,
  env: Env,
  url: URL
): Promise<Response> {
  try {
    const contentType = request.headers.get("content-type") || "";
    let audioData: ArrayBuffer;
    let language: string | undefined;
    let prompt: string | undefined;
    let model: string = env.WHISPER_MODEL || DEFAULT_MODEL;

    // 查询参数
    const queryModel = url.searchParams.get("model");
    const queryLang = url.searchParams.get("language");
    const queryPrompt = url.searchParams.get("prompt");

    if (queryModel) model = queryModel;
    if (queryLang) language = queryLang;
    if (queryPrompt) prompt = queryPrompt;

    if (contentType.includes("multipart/form-data")) {
      const formData = await request.formData();
      const file = formData.get("file");

      if (!file || !(file instanceof Blob)) {
        return new Response(JSON.stringify({ error: "Missing 'file' field" }), {
          status: 400,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (formData.has("model")) {
        const m = formData.get("model");
        if (typeof m === "string") model = m;
      }
      if (formData.has("language")) {
        const l = formData.get("language");
        if (typeof l === "string") language = l;
      }
      if (formData.has("prompt")) {
        const p = formData.get("prompt");
        if (typeof p === "string") prompt = p;
      }

      audioData = await file.arrayBuffer();
    } else {
      audioData = await request.arrayBuffer();
    }

    const options: Record<string, any> = {};
    if (language) options.language = language;
    if (prompt) options.prompt = prompt;

    const result = await env.AI.run(model, {
      audio: audioData,
      ...options,
    });

    const format = url.searchParams.get("response_format") || "json";

    switch (format) {
      case "text":
        return new Response(result.text, {
          headers: { "Content-Type": "text/plain; charset=utf-8" },
        });

      case "srt":
        const srt = generateSRT(result.words || []);
        return new Response(srt, {
          headers: { "Content-Type": "text/srt; charset=utf-8" },
        });

      case "vtt":
        return new Response(result.vtt || "", {
          headers: { "Content-Type": "text/vtt; charset=utf-8" },
        });

      case "verbose_json":
        return new Response(JSON.stringify({
          text: result.text,
          language,
          word_count: result.word_count,
          words: result.words,
          vtt: result.vtt,
        }, null, 2), {
          headers: { "Content-Type": "application/json; charset=utf-8" },
        });

      default: // json
        return new Response(JSON.stringify({
          text: result.text,
          word_count: result.word_count,
          words: result.words,
          vtt: result.vtt,
        }, null, 2), {
          headers: { "Content-Type": "application/json; charset=utf-8" },
        });
    }
  } catch (error: any) {
    return new Response(JSON.stringify({
      error: error.message || "Transcription failed",
      detail: String(error),
    }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
}

function generateSRT(words: any[]): string {
  if (!words || words.length === 0) return "";
  const lines: string[] = [];
  let index = 1;
  const buffer: any[] = [];
  let bufStart = 0;

  function flush() {
    if (buffer.length === 0) return;
    const text = buffer.map(w => w.word).join(" ");
    const start = formatTimeSRT(bufStart);
    const end = formatTimeSRT(buffer[buffer.length - 1].end);
    lines.push(`${index++}\n${start} --> ${end}\n${text}`);
    buffer.length = 0;
  }

  for (const w of words) {
    if (buffer.length === 0) bufStart = w.start;
    buffer.push(w);
    if (w.end - bufStart > 2) {
      flush();
      bufStart = w.start;
      buffer.length = 1;
    }
  }
  flush();

  return lines.join("\n\n") + "\n";
}

function formatTimeSRT(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const ms = Math.floor((seconds % 1) * 1000);
  return `${pad(h, 2)}:${pad(m, 2)}:${pad(s, 2)},${pad(ms, 3)}`;
}

function pad(n: number, len: number): string {
  return String(n).padStart(len, "0");
}