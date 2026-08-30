import { useState, useCallback } from "react";
import {
  Smartphone,
  Monitor,
  Square,
  ChevronRight,
  Check,
  Info,
} from "lucide-react";
import { Button, Input, Label } from "@mybilibili-studio/ui";
import { useProjectStore } from "../../stores/project-store";
import { useAnalytics, AnalyticsEvents } from "../../hooks/useAnalytics";
import {
  SOCIAL_MEDIA_PRESETS,
  SOCIAL_MEDIA_CATEGORY_INFO,
  createProjectSettingsFromPreset,
  type SocialMediaCategory,
} from "@mybilibili-studio/core";

interface StartFromScratchProps {
  onProjectCreated?: () => void;
}

interface PresetGroup {
  platform: string;
  presets: SocialMediaCategory[];
}

const PRESET_GROUPS: PresetGroup[] = [
  {
    platform: "竖屏 (9:16)",
    presets: [
      "tiktok",
      "instagram-reels",
      "instagram-stories",
      "youtube-shorts",
    ],
  },
  {
    platform: "方形 (1:1)",
    presets: ["instagram-post", "facebook"],
  },
  {
    platform: "横屏 (16:9)",
    presets: ["youtube-video", "twitter", "linkedin"],
  },
  {
    platform: "其他",
    presets: ["pinterest", "custom"],
  },
];

const PRESET_ICONS: Record<string, React.ElementType> = {
  "竖屏 (9:16)": Smartphone,
  "方形 (1:1)": Square,
  "横屏 (16:9)": Monitor,
  其他: Square,
};

export const StartFromScratch: React.FC<StartFromScratchProps> = ({
  onProjectCreated,
}) => {
  const createNewProject = useProjectStore((state) => state.createNewProject);
  const updateSettings = useProjectStore((state) => state.updateSettings);
  const { track } = useAnalytics();
  const [selectedPreset, setSelectedPreset] =
    useState<SocialMediaCategory>("youtube-video");
  const [projectName, setProjectName] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  const preset = SOCIAL_MEDIA_PRESETS[selectedPreset];
  const info = SOCIAL_MEDIA_CATEGORY_INFO.find((c) => c.id === selectedPreset);

  const handleCreate = useCallback(async () => {
    setIsCreating(true);

    const settings = createProjectSettingsFromPreset(preset);
    createNewProject(projectName.trim() || `${info?.name || "新建"}项目`);
    await updateSettings(settings);

    track(AnalyticsEvents.PROJECT_CREATED, {
      preset: selectedPreset,
      width: preset.width,
      height: preset.height,
      frameRate: preset.frameRate || 30,
      source: "start_from_scratch",
    });

    setTimeout(() => {
      setIsCreating(false);
      onProjectCreated?.();
    }, 100);
  }, [
    createNewProject,
    updateSettings,
    preset,
    projectName,
    info,
    onProjectCreated,
    track,
    selectedPreset,
  ]);

  return (
    <div className="space-y-6">
      <div>
        <Label className="text-sm font-medium text-text-primary mb-2 block">
          项目名称
        </Label>
        <Input
          type="text"
          value={projectName}
          onChange={(e) => setProjectName(e.target.value)}
          placeholder="我的新视频"
          className="max-w-md bg-background-tertiary border-border text-text-primary"
        />
      </div>

      <div>
        <h3 className="text-sm font-medium text-text-primary mb-4">
          选择画幅
        </h3>

        <div className="grid md:grid-cols-2 gap-6">
          {PRESET_GROUPS.map((group) => {
            const GroupIcon = PRESET_ICONS[group.platform] || Square;

            return (
              <div key={group.platform} className="space-y-3">
                <div className="flex items-center gap-2 text-xs text-text-muted font-medium">
                  <GroupIcon size={14} />
                  <span>{group.platform}</span>
                </div>

                <div className="grid grid-cols-2 gap-2">
                  {group.presets.map((presetId) => {
                    const presetInfo = SOCIAL_MEDIA_CATEGORY_INFO.find(
                      (c) => c.id === presetId,
                    );
                    const presetData = SOCIAL_MEDIA_PRESETS[presetId];
                    const isSelected = selectedPreset === presetId;

                    return (
                      <button
                        key={presetId}
                        onClick={() => setSelectedPreset(presetId)}
                        className={`flex items-center gap-3 p-3 rounded-lg border transition-all duration-200 text-left ${
                          isSelected
                            ? "border-primary bg-primary/10"
                            : "border-border bg-background hover:border-border-hover hover:bg-background-tertiary"
                        }`}
                      >
                        <div
                          className={`w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors ${
                            isSelected
                              ? "border-primary bg-primary"
                              : "border-border"
                          }`}
                        >
                          {isSelected && (
                            <Check size={12} className="text-black" />
                          )}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-xs font-medium text-text-primary truncate">
                            {presetInfo?.name || presetId}
                          </p>
                          <p className="text-[10px] text-text-muted">
                            {presetData.width}×{presetData.height}
                          </p>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="flex items-start gap-3 p-4 bg-background-tertiary rounded-xl border border-border">
        <Info size={16} className="text-primary flex-shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-medium text-text-primary">
            {info?.name || selectedPreset} 画幅
          </p>
          <p className="text-xs text-text-muted mt-1">
            {preset.width}×{preset.height}px • {preset.frameRate || 30}fps
            {preset.maxDuration && ` • 最长 ${preset.maxDuration}s`}
            {preset.recommendedDuration &&
              ` • 推荐 ${preset.recommendedDuration}s`}
          </p>
          {preset.safeZone && (
            <p className="text-xs text-text-muted mt-0.5">
              安全区：上 {preset.safeZone.top}px，下 {preset.safeZone.bottom}px
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center justify-end gap-3">
        <Button
          onClick={handleCreate}
          disabled={isCreating}
          className="shadow-glow"
        >
          {isCreating ? (
            <>
              <div className="w-4 h-4 border-2 border-black/30 border-t-black rounded-full animate-spin" />
              创建中...
            </>
          ) : (
            <>
              创建项目
              <ChevronRight size={16} />
            </>
          )}
        </Button>
      </div>
    </div>
  );
};

export default StartFromScratch;

