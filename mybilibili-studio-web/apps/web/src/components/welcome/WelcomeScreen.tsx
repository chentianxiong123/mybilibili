import { useState, useCallback, useEffect } from "react";
import {
  Clock,
  Layers,
  ArrowRight,
  Smartphone,
  Monitor,
  Square,
  FolderOpen,
  LogOut,
  UserRound,
} from "lucide-react";
import { Button, Switch, Label } from "@mybilibili-studio/ui";
import { useProjectStore } from "../../stores/project-store";
import { useUIStore } from "../../stores/ui-store";
import { SOCIAL_MEDIA_PRESETS, type SocialMediaCategory } from "@mybilibili-studio/core";
import { TemplateGallery } from "./TemplateGallery";
import { RecentProjects } from "./RecentProjects";
import { useRouter } from "../../hooks/use-router";
import { useEditorPreload } from "../../hooks/useEditorPreload";
import { useAnalytics, AnalyticsEvents } from "../../hooks/useAnalytics";
import { useAuthStore } from "../../stores/auth-store";

interface FormatOption {
  id: string;
  preset: SocialMediaCategory;
  label: string;
  description: string;
  dimensions: string;
  icon: React.ElementType;
}

const FORMAT_OPTIONS: FormatOption[] = [
  {
    id: "vertical",
    preset: "tiktok",
    label: "竖屏",
    description: "短视频、竖版内容",
    dimensions: "1080 × 1920",
    icon: Smartphone,
  },
  {
    id: "horizontal",
    preset: "youtube-video",
    label: "横屏",
    description: "投稿视频、网页视频",
    dimensions: "1920 × 1080",
    icon: Monitor,
  },
  {
    id: "square",
    preset: "instagram-post",
    label: "方形",
    description: "社交平台方形内容",
    dimensions: "1080 × 1080",
    icon: Square,
  },
];

const StudioLogo: React.FC<{ className?: string }> = ({ className = "" }) => (
  <svg
    viewBox="0 0 490 490"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <path
      d="M245 24.5C123.223 24.5 24.5 123.223 24.5 245s98.723 220.5 220.5 220.5 220.5-98.723 220.5-220.5S366.777 24.5 245 24.5Z"
      stroke="currentColor"
      strokeWidth="30.625"
    />
    <g>
      <path
        d="M245 98v73.5"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="M392 245h-73.5"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="M245 392v-73.5"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="M98 245h73.5"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="m348.941 141.059-51.965 51.965"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="m348.941 348.941-51.965-51.965"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="m141.059 348.941 51.965-51.965"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
      <path
        d="m141.059 141.059 51.965 51.965"
        stroke="currentColor"
        strokeWidth="24.5"
        strokeLinecap="round"
      />
    </g>
    <path
      d="M294 245a49 49 0 0 1-49 49 49 49 0 0 1-49-49 49 49 0 0 1 98 0"
      fill="currentColor"
    />
  </svg>
);

type ViewMode = "home" | "templates" | "recent";

interface WelcomeScreenProps {
  initialTab?: "templates" | "recent";
}

export const WelcomeScreen: React.FC<WelcomeScreenProps> = ({ initialTab }) => {
  const setSkipWelcomeScreen = useUIStore(
    (state) => state.setSkipWelcomeScreen,
  );
  const skipWelcomeScreen = useUIStore((state) => state.skipWelcomeScreen);
  const openModal = useUIStore((state) => state.openModal);
  const createNewProject = useProjectStore((state) => state.createNewProject);
  const authUser = useAuthStore((state) => state.user);
  const authStatus = useAuthStore((state) => state.status);
  const logout = useAuthStore((state) => state.logout);
  const { navigate } = useRouter();
  const { track } = useAnalytics();

  const [viewMode, setViewMode] = useState<ViewMode>(initialTab ?? "home");
  const [hoveredFormat, setHoveredFormat] = useState<string | null>(null);

  useEditorPreload(true);

  const handleCreateProject = useCallback(
    (option: FormatOption) => {
      const preset = SOCIAL_MEDIA_PRESETS[option.preset];
      createNewProject(`新建${option.label}项目`, {
        width: preset.width,
        height: preset.height,
        frameRate: preset.frameRate,
      });
      track(AnalyticsEvents.PROJECT_CREATED, {
        preset: option.preset,
        width: preset.width,
        height: preset.height,
        frameRate: preset.frameRate ?? 30,
        source: "quick_start",
      });
      navigate("editor");
    },
    [createNewProject, navigate, track],
  );

  const handleTemplateApplied = useCallback(() => {
    navigate("editor");
  }, [navigate]);

  const handleProjectSelected = useCallback(() => {
    navigate("editor");
  }, [navigate]);

  useEffect(() => {
    if (skipWelcomeScreen) {
      navigate("editor");
    }
  }, [skipWelcomeScreen, navigate]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (viewMode !== "home") {
          setViewMode("home");
        } else {
          navigate("editor");
        }
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [navigate, viewMode]);

  if (viewMode === "templates") {
    return (
      <div className="fixed inset-0 z-50 bg-background flex flex-col">
        <header className="flex items-center justify-between px-6 py-4 border-b border-border">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setViewMode("home")}
          >
            <ArrowRight className="rotate-180" size={16} />
            返回
          </Button>
          <h2 className="text-sm font-medium text-text-primary">模板</h2>
          <div className="w-16" />
        </header>
        <div className="flex-1 overflow-y-auto p-6">
          <TemplateGallery onTemplateApplied={handleTemplateApplied} />
        </div>
      </div>
    );
  }

  if (viewMode === "recent") {
    return (
      <div className="fixed inset-0 z-50 bg-background flex flex-col">
        <header className="flex items-center justify-between px-6 py-4 border-b border-border">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setViewMode("home")}
          >
            <ArrowRight className="rotate-180" size={16} />
            返回
          </Button>
          <h2 className="text-sm font-medium text-text-primary">
            最近项目
          </h2>
          <div className="w-16" />
        </header>
        <div className="flex-1 overflow-y-auto p-6">
          <RecentProjects onProjectSelected={handleProjectSelected} />
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-50 bg-background overflow-hidden">
      <div className="absolute right-6 top-6 z-10 flex items-center gap-2">
        {authStatus === "authenticated" && authUser ? (
          <>
            <div className="flex items-center gap-2 rounded-md border border-border bg-background-secondary px-3 py-2 text-sm text-text-primary">
              {authUser.avatar ? (
                <img
                  src={authUser.avatar}
                  alt="用户头像"
                  className="h-6 w-6 rounded-full object-cover"
                />
              ) : (
                <UserRound size={16} className="text-text-muted" />
              )}
              <span className="max-w-[140px] truncate">
                {authUser.nickname || authUser.username}
              </span>
            </div>
            <Button variant="outline" size="sm" onClick={logout}>
              <LogOut size={15} />
              退出
            </Button>
          </>
        ) : (
          <Button variant="outline" size="sm" onClick={() => openModal("auth")}>
            <UserRound size={15} />
            登录
          </Button>
        )}
      </div>

      <div className="relative h-full flex flex-col items-center justify-center px-6">
        <div className="w-full max-w-3xl">
          <div className="flex flex-col items-center text-center mb-10">
            <div className="flex items-center gap-3 mb-5">
              <div className="w-12 h-12 text-primary">
                <StudioLogo className="w-full h-full" />
              </div>
              <span className="text-xl font-semibold text-text-primary">
                mybilibili 剪辑工作室
              </span>
            </div>

            <h1 className="text-3xl sm:text-4xl font-bold text-text-primary mb-3">
              新建剪辑项目
            </h1>
            <p className="text-base text-text-secondary mb-2">
              选择画布比例后进入工作台。
            </p>
            <p className="text-sm text-text-muted max-w-md">
              模板和最近项目可在下方打开。
            </p>
          </div>

          <div className="grid grid-cols-3 gap-3 mb-8">
            {FORMAT_OPTIONS.map((option) => {
              const Icon = option.icon;
              const isHovered = hoveredFormat === option.id;

              return (
                <button
                  key={option.id}
                  onClick={() => handleCreateProject(option)}
                  onMouseEnter={() => setHoveredFormat(option.id)}
                  onMouseLeave={() => setHoveredFormat(null)}
                  className={`
                    group relative flex flex-col items-center p-5 rounded-md
                    bg-background-secondary border border-border
                    hover:border-primary/40 hover:bg-background-tertiary
                    transition-all duration-200
                    ${isHovered ? "scale-[1.02] shadow-lg shadow-primary/5" : ""}
                  `}
                >
                  <div className="relative z-10 flex flex-col items-center">
                    <div
                      className={`
                      w-14 h-14 mb-4 rounded-md flex items-center justify-center
                      bg-background-tertiary group-hover:bg-primary/10
                      transition-colors duration-200
                    `}
                    >
                      <Icon
                        size={28}
                        className="text-text-muted group-hover:text-primary transition-colors"
                      />
                    </div>

                    <h3 className="text-lg font-semibold text-text-primary mb-1">
                      {option.label}
                    </h3>
                    <p className="text-sm text-text-muted mb-3">
                      {option.description}
                    </p>
                    <span className="text-xs font-mono text-text-muted/70 bg-background-tertiary px-2 py-1 rounded">
                      {option.dimensions}
                    </span>
                  </div>

                  <div
                    className={`
                    absolute bottom-4 left-1/2 -translate-x-1/2
                    flex items-center gap-1 text-sm font-medium text-primary
                    opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0
                    transition-all duration-200
                  `}
                  >
                    开始创建
                    <ArrowRight size={14} />
                  </div>
                </button>
              );
            })}
          </div>

          <div className="flex items-center justify-center gap-3">
            <Button
              variant="outline"
              onClick={() => setViewMode("templates")}
              className="rounded-md"
            >
              <Layers size={16} />
              浏览模板
            </Button>
            <Button
              variant="outline"
              onClick={() => setViewMode("recent")}
              className="rounded-md"
            >
              <Clock size={16} />
              最近项目
            </Button>
            <Button
              variant="outline"
              onClick={() => navigate("editor")}
              className="rounded-md"
            >
              <FolderOpen size={16} />
              打开编辑器
            </Button>
          </div>
        </div>

        <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Switch
              id="skip-welcome"
              checked={skipWelcomeScreen}
              onCheckedChange={setSkipWelcomeScreen}
            />
            <Label
              htmlFor="skip-welcome"
              className="text-xs text-text-muted cursor-pointer"
            >
              启动时跳过
            </Label>
          </div>

          <span className="text-text-muted/30">·</span>

          <p className="text-xs text-text-muted/60">
            按{" "}
            <kbd className="px-1.5 py-0.5 bg-background-tertiary border border-border rounded text-text-muted font-mono text-[10px]">
              Esc
            </kbd>{" "}
            跳过
          </p>
        </div>
      </div>
    </div>
  );
};

export default WelcomeScreen;

