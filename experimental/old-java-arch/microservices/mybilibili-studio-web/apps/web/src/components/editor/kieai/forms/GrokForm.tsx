import { Button } from "@mybilibili-studio/ui";
import type { GrokInput } from "../../../../services/kieai/image-generation";

interface Props {
  value: GrokInput;
  onChange: (v: GrokInput) => void;
  onSubmit: () => void;
  isLoading: boolean;
}

export function GrokForm({ value, onChange, onSubmit, isLoading }: Props) {
  return (
    <div className="space-y-4">
      <div className="p-2 rounded-lg bg-blue-500/10 border border-blue-500/30 text-xs text-blue-400">
        Grok Imagine 会参考源图片的风格和构图，也可以通过提示词控制变化方向。
      </div>

      <div className="space-y-1.5">
        <label className="text-xs font-medium text-text-secondary">提示词（可选）</label>
        <textarea
          value={value.prompt ?? ""}
          onChange={(e) => onChange({ ...value, prompt: e.target.value || undefined })}
          placeholder="可选：描述你想调整或强调的内容..."
          maxLength={1000}
          rows={3}
          className="w-full rounded-lg border border-border bg-background-elevated px-3 py-2 text-sm text-text-primary placeholder:text-text-muted resize-none outline-none focus:border-primary"
        />
        <p className="text-[10px] text-text-muted text-right">{(value.prompt ?? "").length}/1000</p>
      </div>

      <Button onClick={onSubmit} disabled={isLoading} className="w-full">
        {isLoading ? "生成中..." : "使用 Grok Imagine 生成"}
      </Button>
    </div>
  );
}

