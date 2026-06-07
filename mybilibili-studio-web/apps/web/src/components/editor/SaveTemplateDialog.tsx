import { useState, useCallback } from "react";
import { Upload, Cloud, HardDrive, Check, AlertCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  Button,
  Input,
  Label,
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@mybilibili-studio/ui";
import { useProjectStore } from "../../stores/project-store";
import { useEngineStore } from "../../stores/engine-store";
import {
  TEMPLATE_CATEGORIES,
  type TemplateCategory,
  type TemplatePlaceholder,
  type Template,
  type ShapeClip,
  type SVGClip,
  type StickerClip,
} from "@mybilibili-studio/core";
import { templateCloudService } from "../../services/template-cloud-service";

interface TemplateWithGraphics extends Template {
  timeline: Template["timeline"] & {
    graphics?: {
      shapes: ShapeClip[];
      svgs: SVGClip[];
      stickers: StickerClip[];
    };
  };
}

interface SaveTemplateDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export const SaveTemplateDialog: React.FC<SaveTemplateDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const { project } = useProjectStore();
  const getTemplateEngine = useEngineStore((state) => state.getTemplateEngine);
  const getGraphicsEngine = useEngineStore((state) => state.getGraphicsEngine);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState<TemplateCategory>("custom");
  const [tags, setTags] = useState("");
  const [author, setAuthor] = useState("");
  const [saveLocation, setSaveLocation] = useState<"local" | "cloud">("cloud");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSave = useCallback(async () => {
    if (!name.trim()) {
      setError("请填写模板名称");
      return;
    }

    if (!description.trim()) {
      setError("请填写模板说明");
      return;
    }

    setIsSaving(true);
    setError(null);

    try {
      const templateEngine = await getTemplateEngine();
      const graphicsEngine = getGraphicsEngine();

      await templateEngine.initialize();

      const tagArray = tags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);

      const placeholders: TemplatePlaceholder[] = [];

      const template = templateEngine.createFromProject(project, {
        name: name.trim(),
        description: description.trim(),
        category,
        placeholders,
        tags: tagArray,
      });

      const templateWithMeta = {
        ...template,
        author: author.trim() || "Anonymous",
      };

      if (graphicsEngine) {
        const shapes = graphicsEngine.getAllShapeClips();
        const svgs = graphicsEngine.getAllSVGClips();
        const stickers = graphicsEngine.getAllStickerClips();

        if (shapes.length > 0 || svgs.length > 0 || stickers.length > 0) {
          (templateWithMeta as TemplateWithGraphics).timeline.graphics = {
            shapes,
            svgs,
            stickers,
          };
        }
      }

      if (saveLocation === "cloud") {
        const result =
          await templateCloudService.uploadTemplate(templateWithMeta);
        if (!result.success) {
          throw new Error(result.error || "上传到云端失败");
        }
      } else {
        await templateEngine.saveTemplate(templateWithMeta);
      }

      setSuccess(true);
      setTimeout(() => {
        onClose();
        setSuccess(false);
        setName("");
        setDescription("");
        setTags("");
        setAuthor("");
        setCategory("custom");
      }, 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存模板失败");
    } finally {
      setIsSaving(false);
    }
  }, [
    name,
    description,
    category,
    tags,
    author,
    saveLocation,
    project,
    getTemplateEngine,
    getGraphicsEngine,
    onClose,
  ]);

  if (!isOpen) return null;

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg p-0 gap-0 bg-background border-border overflow-hidden">
        <DialogHeader className="p-4 border-b border-border space-y-0">
          <DialogTitle className="text-lg font-semibold text-text-primary">
            保存为模板
          </DialogTitle>
        </DialogHeader>

        <div className="p-4 space-y-4 max-h-[70vh] overflow-y-auto">
          {success && (
            <div className="flex items-center gap-2 p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
              <Check size={16} className="text-green-400" />
              <span className="text-sm text-green-400">
                模板保存成功！
              </span>
            </div>
          )}

          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
              <AlertCircle size={16} className="text-red-400" />
              <span className="text-sm text-red-400">{error}</span>
            </div>
          )}

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              模板名称 <span className="text-red-400">*</span>
            </Label>
            <Input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="我的模板"
              className="bg-background-secondary border-border text-text-primary"
              maxLength={50}
            />
            <p className="text-[10px] text-text-muted">
              {name.length}/50 字
            </p>
          </div>

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              模板说明 <span className="text-red-400">*</span>
            </Label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="说明这个模板的用途和使用方式..."
              className="w-full px-3 py-2 text-sm bg-background-secondary border border-border rounded-lg text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
              rows={4}
              maxLength={500}
            />
            <p className="text-[10px] text-text-muted">
              {description.length}/500 字
            </p>
          </div>

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              分类
            </Label>
            <Select value={category} onValueChange={(value) => setCategory(value as TemplateCategory)}>
              <SelectTrigger className="w-full bg-background-secondary border-border text-text-primary">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-background-secondary border-border">
                {TEMPLATE_CATEGORIES.map((cat) => (
                  <SelectItem key={cat.id} value={cat.id}>
                    {cat.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              标签（用逗号分隔）
            </Label>
            <Input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="片头, 动画, 投稿"
              className="bg-background-secondary border-border text-text-primary"
            />
          </div>

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              作者名称
            </Label>
            <Input
              type="text"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              placeholder="你的昵称或用户名"
              className="bg-background-secondary border-border text-text-primary"
            />
          </div>

          <div className="space-y-2">
            <Label className="text-xs font-medium text-text-secondary">
              保存位置
            </Label>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => setSaveLocation("cloud")}
                className={`flex items-center justify-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
                  saveLocation === "cloud"
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-background-secondary text-text-secondary hover:border-primary/50"
                }`}
              >
                <Cloud size={16} />
                <span className="text-sm font-medium">云端</span>
              </button>
              <button
                onClick={() => setSaveLocation("local")}
                className={`flex items-center justify-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
                  saveLocation === "local"
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-background-secondary text-text-secondary hover:border-primary/50"
                }`}
              >
                <HardDrive size={16} />
                <span className="text-sm font-medium">本地</span>
              </button>
            </div>
            <p className="text-[10px] text-text-muted">
              {saveLocation === "cloud"
                ? "保存到云端，可在其他设备访问"
                : "保存到当前浏览器本地存储"}
            </p>
          </div>
        </div>

        <div className="flex items-center justify-end gap-2 p-4 border-t border-border">
          <Button variant="ghost" onClick={onClose} disabled={isSaving}>
            取消
          </Button>
          <Button
            onClick={handleSave}
            disabled={isSaving || !name.trim() || !description.trim()}
          >
            {isSaving ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent" />
                保存中...
              </>
            ) : (
              <>
                <Upload size={16} />
                保存模板
              </>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

