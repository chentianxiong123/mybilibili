<template>
  <header class="h-topbar grid grid-cols-[1fr_auto_1fr] items-center gap-2.5 px-3 bg-bg border-b border-border shrink-0 z-30 relative text-[oklch(0.9_0_0)]">
    <!-- ─── Left: window dots + autosave ─────────────────────── -->
    <div class="flex items-center gap-2">
      <button
        @click="navigateToWelcome"
        class="flex items-center gap-1.5 pr-1.5"
        title="返回首页"
      >
        <span class="w-[11px] h-[11px] rounded-full bg-[oklch(0.7_0.18_25)]" />
        <span class="w-[11px] h-[11px] rounded-full bg-[oklch(0.78_0.14_80)]" />
        <span class="w-[11px] h-[11px] rounded-full bg-[oklch(0.7_0.15_145)]" />
      </button>

      <span class="text-[11px] text-fg-3 flex items-center gap-1.5">
        <span class="w-[5px] h-[5px] rounded-full bg-accent" />
        <template v-if="exportStateLocal.isExporting">
          导出中... {{ Math.round(exportStateLocal.progress) }}%
        </template>
        <template v-else>
          自动保存：{{ autosaveLabel }}
        </template>
      </span>
    </div>

    <!-- ─── Center: project name ────────────────────────────── -->
    <div class="flex items-center gap-1.5 text-[12.5px] font-medium">
      <input
        v-model="projectNameDraft"
        @blur="commitProjectName"
        @keydown.enter="blurInput"
        @keydown.esc="cancelRename"
        :size="Math.max(projectNameDraft.length, 6)"
        spellcheck="false"
        class="bg-transparent border-0 text-center font-medium text-[12.5px] text-fg px-2 py-0.5 rounded min-w-[60px] focus:bg-bg-2 focus:outline-none"
      />
      <!-- 使用 ReactAdapter 渲染 ProjectSwitcher React 组件 -->
      <ReactAdapter :component="ProjectSwitcherRaw" />
    </div>

    <!-- ─── Right: undo/redo, history, comments, settings, export ── -->
    <div class="flex items-center justify-end gap-1.5">
      <!-- Search (⌘K) -->
      <button
        @click="uiStore.openModal('search')"
        class="w-[26px] h-[26px] grid place-items-center rounded-md text-fg-2 hover:bg-hover hover:text-fg transition-colors"
        title="搜索工具、特效或询问 AI (⌘K)"
      >
        <SearchIcon :size="14" />
      </button>

      <!-- Undo -->
      <button
        @click="handleUndo"
        class="w-[26px] h-[26px] grid place-items-center rounded-md text-fg-2 hover:bg-hover hover:text-fg transition-colors"
        title="撤销 (⌘Z)"
      >
        <UndoIcon :size="14" />
      </button>

      <!-- Redo -->
      <button
        @click="handleRedo"
        class="w-[26px] h-[26px] grid place-items-center rounded-md text-fg-2 hover:bg-hover hover:text-fg transition-colors"
        title="重做 (⇧⌘Z)"
      >
        <RedoIcon :size="14" />
      </button>

      <div class="w-px h-4 bg-border mx-1" />

      <!-- History -->
      <button
        @click="isHistoryOpen = !isHistoryOpen"
        :class="[
          'w-[26px] h-[26px] grid place-items-center rounded-md transition-colors',
          isHistoryOpen ? 'bg-accent-soft text-accent' : 'text-fg-2 hover:bg-hover hover:text-fg'
        ]"
        title="操作历史"
      >
        <HistoryIcon :size="14" />
      </button>

      <!-- Keyframe editor -->
      <button
        @click="uiStore.toggleKeyframeEditor()"
        :class="[
          'w-[26px] h-[26px] grid place-items-center rounded-md transition-colors',
          uiState.keyframeEditorOpen ? 'bg-accent-soft text-accent' : 'text-fg-2 hover:bg-hover hover:text-fg'
        ]"
        title="关键帧编辑器"
      >
        <DiamondIcon :size="14" />
      </button>

      <!-- Audio mixer -->
      <button
        @click="uiStore.togglePanel('audioMixer')"
        :class="[
          'w-[26px] h-[26px] grid place-items-center rounded-md transition-colors',
          uiState.panels.audioMixer?.visible ? 'bg-accent-soft text-accent' : 'text-fg-2 hover:bg-hover hover:text-fg'
        ]"
        title="音频混音器"
      >
        <MusicIcon :size="14" />
      </button>

      <!-- Project JSON / Comments -->
      <button
        @click="uiStore.openModal('scriptView')"
        class="w-[26px] h-[26px] grid place-items-center rounded-md text-fg-2 hover:bg-hover hover:text-fg transition-colors"
        title="项目 JSON"
      >
        <MessageSquareIcon :size="14" />
      </button>

      <div class="w-px h-4 bg-border mx-1" />

      <!-- Settings Dropdown Button -->
      <div class="relative">
        <button
          @click="isMenuOpen = !isMenuOpen"
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[12px] font-medium text-fg-2 hover:bg-hover hover:text-fg transition-colors"
          title="更多工具"
        >
          <StarIcon :size="14" />
        </button>
        <div
          v-if="isMenuOpen"
          class="absolute right-0 mt-1 w-56 rounded-md bg-bg-1 border border-border shadow-lg z-50 p-1 space-y-0.5 text-xs text-fg-2"
        >
          <button @click="toggleTheme" class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2">
            <SunIcon v-if="themeState.mode === 'light'" :size="14" />
            <MoonIcon v-else-if="themeState.mode === 'dark'" :size="14" />
            <SunMoonIcon v-else :size="14" />
            <span>主题：{{ themeLabel }}</span>
          </button>
          <button @click="openSettings" class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2">
            <SettingsIcon :size="14" />
            <span>设置与接口密钥</span>
          </button>
          <button @click="isRecorderOpen = true; isMenuOpen = false" class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2">
            <CircleIcon :size="14" class="fill-current text-status-error" />
            <span>屏幕录制</span>
          </button>
          <div class="h-px bg-border my-1" />
          <button @click="handleStartTour" class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2">
            <PlayIcon :size="14" />
            <span>编辑器导览</span>
          </button>
          <button @click="handleStartMoGraphTour" class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2">
            <SparklesIcon :size="14" class="text-purple-400" />
            <span>动画导览</span>
          </button>
        </div>
      </div>

      <!-- Export Panel -->
      <template v-if="exportStateLocal.isExporting">
        <button
          @click="handleCancelExport"
          class="inline-flex items-center gap-1.5 px-3 py-1 rounded-md bg-accent-soft text-accent text-[12.5px] font-semibold"
        >
          <Loader2Icon :size="13" class="animate-spin" />
          <span>{{ Math.round(exportStateLocal.progress) }}%</span>
          <XIcon :size="11" class="ml-1 opacity-70" />
        </button>
      </template>
      <template v-else-if="exportStateLocal.error">
        <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-md border border-status-error/40 bg-status-error/10 text-status-error text-[11px]">
          <span class="max-w-[180px] truncate">{{ exportStateLocal.error }}</span>
          <button
            @click="exportStateLocal.error = null"
            class="opacity-70 hover:opacity-100"
          >
            <XIcon :size="11" />
          </button>
        </div>
      </template>
      <template v-else-if="exportStateLocal.complete">
        <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-md bg-accent-soft text-accent text-[12.5px]">
          <CheckIcon :size="13" />
          <span class="font-medium">已保存</span>
        </div>
      </template>
      <template v-else>
        <div class="relative">
          <button
            @click="isAccountOpen = !isAccountOpen"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[12px] font-medium text-fg-2 hover:bg-hover hover:text-fg transition-colors"
            title="账号"
          >
            <img
              v-if="authAvatar"
              :src="authAvatar"
              alt="用户头像"
              class="w-[18px] h-[18px] rounded-full object-cover"
            />
            <UserRoundIcon v-else :size="14" />
            <span class="max-w-[86px] truncate">{{ accountLabel }}</span>
          </button>
          <div
            v-if="isAccountOpen"
            class="absolute right-0 mt-1 w-48 rounded-md bg-bg-1 border border-border shadow-lg z-50 p-1 space-y-0.5 text-xs text-fg-2"
          >
            <div v-if="authState.status === 'authenticated'" class="px-2.5 py-2 border-b border-border">
              <div class="text-fg font-medium truncate">{{ accountLabel }}</div>
              <div class="text-[10px] text-fg-muted truncate">已登录 mybilibili</div>
            </div>
            <button
              v-if="authState.status === 'authenticated'"
              @click="handleLogout"
              class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2"
            >
              <LogOutIcon :size="14" />
              <span>退出登录</span>
            </button>
            <button
              v-else
              @click="openAuthDialog"
              class="w-full text-left px-2.5 py-1.5 hover:bg-hover hover:text-fg rounded flex items-center gap-2"
            >
              <UserRoundIcon :size="14" />
              <span>登录</span>
            </button>
          </div>
        </div>

        <div class="relative">
          <button
            @click="isExportOpen = !isExportOpen"
            class="relative inline-flex items-center gap-1.5 px-3.5 py-[5px] rounded-md bg-accent text-accent-fg font-semibold text-[12.5px] shadow-glow hover:bg-accent-strong transition-colors"
          >
            <UploadIcon :size="13" />
            <span>导出</span>
            <ChevronDownIcon :size="12" :class="['transition-transform', isExportOpen ? 'rotate-180' : '']" />
          </button>

          <!-- Export Dropdown Options -->
          <div
            v-if="isExportOpen"
            class="absolute right-0 mt-1 w-72 rounded-xl bg-bg-1 border border-border shadow-lg z-50 p-2 space-y-1 text-xs"
          >
            <div class="max-h-[350px] overflow-y-auto space-y-1">
              <template v-for="(option, index) in exportOptions" :key="index">
                <div v-if="option.separator" class="h-px bg-border my-1" />
                <button
                  v-else
                  @click="handleExport(option.type)"
                  :class="[
                    'w-full text-left flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-hover transition-colors',
                    option.recommended ? 'bg-accent-soft' : ''
                  ]"
                >
                  <div
                    :class="[
                      'p-2 rounded-lg transition-colors',
                      option.recommended ? 'bg-accent-soft text-accent' : 'bg-bg-2 text-fg-2'
                    ]"
                  >
                    <component :is="option.icon" :size="18" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div :class="['text-sm font-medium', option.recommended ? 'text-accent' : 'text-fg']">
                      {{ option.label }}
                      <span v-if="option.recommended" class="ml-1.5 text-[9px] bg-accent-soft text-accent px-1.5 py-0.5 rounded">
                        推荐
                      </span>
                    </div>
                    <div class="text-[11px] text-fg-muted mt-0.5 truncate">{{ option.desc }}</div>
                    <div v-if="exportEstimates.get(option.type)" class="text-[9px] text-fg-3 mt-1">
                      预计 {{ exportEstimates.get(option.type)?.formatted }}
                    </div>
                  </div>
                </button>
              </template>

              <div class="h-px bg-border my-1" />
              <button
                @click="isExportDialogOpen = true; isExportOpen = false"
                class="w-full text-left flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-hover transition-colors"
              >
                <div class="p-2 bg-accent-soft rounded-lg text-accent">
                  <SettingsIcon :size="18" />
                </div>
                <div class="flex-1">
                  <div class="text-sm font-medium text-accent">自定义导出...</div>
                  <div class="text-[11px] text-fg-muted mt-0.5">完整参数与 AI 放大</div>
                </div>
                <MoreHorizontalIcon :size="14" class="text-fg-muted" />
              </button>
            </div>
            <div class="bg-bg-2 px-3 py-2 text-center text-[10px] text-fg-muted border-t border-border rounded-b-lg">
              {{ projectState.project.settings.width }}×{{ projectState.project.settings.height }} • {{ projectState.project.settings.frameRate }}fps
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- ─── Auxiliary React popups & dialogs rendered via ReactAdapter ─── -->
    <ReactAdapter
      :component="ExportDialogRaw"
      :componentProps="{
        isOpen: isExportDialogOpen,
        onClose: () => isExportDialogOpen = false,
        onExport: handleCustomExport,
        duration: projectState.project.timeline?.duration ?? 0,
        projectWidth: projectState.project.settings?.width ?? 1920,
        projectHeight: projectState.project.settings?.height ?? 1080
      }"
    />

    <ReactAdapter
      :component="ScreenRecorderRaw"
      :componentProps="{
        isOpen: isRecorderOpen,
        onClose: () => isRecorderOpen = false,
        onRecordingComplete: handleRecordingComplete
      }"
    />

    <ReactAdapter :component="SettingsDialogRaw" />

    <div v-if="isHistoryOpen">
      <div
        class="fixed inset-0 bg-black/20 z-40"
        @click="isHistoryOpen = false"
      />
      <div class="fixed top-topbar right-0 bottom-0 w-80 bg-bg-1 border-l border-border z-50 shadow-lg flex flex-col">
        <div class="flex items-center justify-between p-3 border-b border-border bg-bg-1">
          <span class="text-sm font-medium text-fg">操作历史</span>
          <button
            @click="isHistoryOpen = false"
            class="p-1.5 rounded hover:bg-hover text-fg-3 hover:text-fg transition-colors"
          >
            <XIcon :size="14" />
          </button>
        </div>
        <div class="flex-1 min-h-0 bg-bg-1">
          <ReactAdapter :component="HistoryPanelRaw" />
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, markRaw } from "vue";
import {
  ChevronDown as ChevronDownIcon,
  FileVideo as FileVideoIcon,
  Film as FilmIcon,
  Music as MusicIcon,
  Sun as SunIcon,
  Moon as MoonIcon,
  SunMoon as SunMoonIcon,
  Loader2 as Loader2Icon,
  X as XIcon,
  Check as CheckIcon,
  FileCode as FileCodeIcon,
  Settings as SettingsIcon,
  Zap as ZapIcon,
  Circle as CircleIcon,
  History as HistoryIcon,
  HelpCircle as HelpCircleIcon,
  Diamond as DiamondIcon,
  Sparkles as SparklesIcon,
  Play as PlayIcon,
  Undo2 as UndoIcon,
  Redo2 as RedoIcon,
  MessageSquare as MessageSquareIcon,
  Star as StarIcon,
  Upload as UploadIcon,
  MoreHorizontal as MoreHorizontalIcon,
  Command as CommandIcon,
  Search as SearchIcon,
  UserRound as UserRoundIcon,
  LogOut as LogOutIcon,
} from "lucide-vue-next";

import ReactAdapter from "../ReactAdapter.vue";
import { useZustandStore } from "../../hooks/useZustandStore";
import { useProjectStore } from "../../stores/project-store";
import { useUIStore } from "../../stores/ui-store";
import { useThemeStore } from "../../stores/theme-store";
import { useSettingsStore } from "../../stores/settings-store";
import { useAuthStore } from "../../stores/auth-store";
import { navigateToRoute } from "../../hooks/use-router";
import { toast } from "../../stores/notification-store";
import {
  createStudioCloudExportTask,
  getStudioCloudExportTask,
  type StudioExportTask,
} from "../../services/studio-cloud-export";

// React components that we wrap inside our ReactAdapter
import ExportDialog from "./ExportDialog";
import { ScreenRecorder } from "./ScreenRecorder";
import { SettingsDialog } from "./settings/SettingsDialog";
import HistoryPanel from "./inspector/HistoryPanel";
import { ProjectSwitcher } from "./ProjectSwitcher";

import {
  getExportEngine,
  getDeviceProfile,
  estimateExportTime,
  type VideoExportSettings,
  type AudioExportSettings,
  type ExportResult,
  type DeviceProfile,
  type TimeEstimate,
} from "@mybilibili-studio/core";

const ExportDialogRaw = markRaw(ExportDialog);
const ScreenRecorderRaw = markRaw(ScreenRecorder);
const SettingsDialogRaw = markRaw(SettingsDialog);
const HistoryPanelRaw = markRaw(HistoryPanel);
const ProjectSwitcherRaw = markRaw(ProjectSwitcher);

const projectStore = useProjectStore;
const uiStore = useUIStore;
const themeStore = useThemeStore;
const settingsStore = useSettingsStore;
const authStore = useAuthStore;

// Bind Zustand stores reactively
const projectState = useZustandStore<any>(projectStore);
const uiState = useZustandStore<any>(uiStore);
const themeState = useZustandStore<any>(themeStore);
const authState = useZustandStore<any>(authStore);

// Router actions
const openSettings = settingsStore.getState().openSettings;

// State management
const isMenuOpen = ref(false);
const isExportOpen = ref(false);
const isExportDialogOpen = ref(false);
const isRecorderOpen = ref(false);
const isHistoryOpen = ref(false);
const isAccountOpen = ref(false);

const projectNameDraft = ref(projectState.value.project.name);

watch(
  () => projectState.value.project.name,
  (name) => {
    projectNameDraft.value = name;
  }
);

const autosaveLabel = computed(() => {
  const ts = projectState.value.project.modifiedAt ?? Date.now();
  const d = new Date(ts);
  return d.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
});

const commitProjectName = () => {
  const next = projectNameDraft.value.trim();
  if (next && next !== projectState.value.project.name) {
    projectStore.getState().renameProject(next);
  } else {
    projectNameDraft.value = projectState.value.project.name;
  }
};

const cancelRename = (e: Event) => {
  projectNameDraft.value = projectState.value.project.name;
  (e.target as HTMLInputElement).blur();
};

const blurInput = (e: Event) => {
  (e.target as HTMLInputElement).blur();
};

const navigateToWelcome = () => {
  navigateToRoute("welcome");
};

const accountLabel = computed(() => {
  const user = authState.value.user;
  return user?.nickname || user?.username || "登录";
});

const authAvatar = computed(() => authState.value.user?.avatar || "");

const themeLabel = computed(() => {
  if (themeState.value.mode === "light") return "浅色";
  if (themeState.value.mode === "dark") return "深色";
  return "跟随系统";
});

const openAuthDialog = () => {
  isAccountOpen.value = false;
  uiStore.getState().openModal("auth");
};

const handleLogout = () => {
  authStore.getState().logout();
  isAccountOpen.value = false;
  toast.success("已退出登录");
};

const handleUndo = () => {
  projectStore.getState().undo();
};

const handleRedo = () => {
  projectStore.getState().redo();
};

const handleStartTour = () => {
  localStorage.removeItem("openreel_onboarding_tour");
  import("./tour").then((m) => m.startTour());
};

const handleStartMoGraphTour = () => {
  localStorage.removeItem("openreel_mograph_tour");
  import("./tour").then((m) => m.startMoGraphTour());
};

// Export handling
const exportStateLocal = ref({
  isExporting: false,
  progress: 0,
  phase: "",
  error: null as string | null,
  complete: false,
});

const exportPhaseLabels: Record<string, string> = {
  complete: "完成",
  initializing: "初始化中",
  loading: "加载中",
  rendering: "渲染中",
  encoding: "编码中",
  muxing: "封装中",
  writing: "写入中",
  processing: "处理中",
};

const formatExportPhase = (phase: string) => exportPhaseLabels[phase] || "处理中";

const cloudExportStageLabels: Record<string, string> = {
  QUEUED: "等待渲染节点",
  RENDERER_RECEIVED: "渲染节点已接收",
  RENDERER_NOT_READY: "渲染器未接入",
  QUEUE_FAILED: "任务入队失败",
  WORKER_FAILED: "渲染节点异常",
  CANCELLED: "已取消",
  SUCCEEDED: "完成",
};

const sleep = (ms: number) => new Promise((resolve) => window.setTimeout(resolve, ms));

const formatCloudExportPhase = (task: StudioExportTask) => {
  if (task.errorMessage) return task.errorMessage;
  if (task.stage && cloudExportStageLabels[task.stage]) return cloudExportStageLabels[task.stage];
  return task.message || "云端导出处理中";
};

const syncCloudExportState = (task: StudioExportTask) => {
  const terminal = ["SUCCEEDED", "FAILED", "CANCELLED"].includes(task.status);
  exportStateLocal.value = {
    isExporting: !terminal,
    progress: task.progress ?? 0,
    phase: formatCloudExportPhase(task),
    error: task.status === "FAILED" ? formatCloudExportPhase(task) : null,
    complete: task.status === "SUCCEEDED",
  };
};

const pollCloudExportTask = async (taskId: string) => {
  for (let i = 0; i < 8; i += 1) {
    await sleep(1000);
    const task = await getStudioCloudExportTask(taskId);
    syncCloudExportState(task);
    if (["SUCCEEDED", "FAILED", "CANCELLED"].includes(task.status)) {
      return task;
    }
  }
  return null;
};

watch(
  () => exportStateLocal.value,
  (state) => {
    uiStore.getState().setExportState({
      isExporting: state.isExporting,
      progress: state.progress,
      phase: state.phase,
    });
  },
  { deep: true }
);

const deviceProfile = ref<DeviceProfile | null>(null);
const exportEstimates = ref<Map<string, TimeEstimate>>(new Map());

onMounted(() => {
  getDeviceProfile().then((profile) => {
    deviceProfile.value = profile;
  });
});

watch(
  [deviceProfile, () => projectState.value.project.timeline?.duration, () => projectState.value.project.settings.width, () => projectState.value.project.settings.height],
  () => {
    if (!deviceProfile.value || !projectState.value.project.timeline?.duration) return;

    const duration = projectState.value.project.timeline.duration;
    const estimates = new Map<string, TimeEstimate>();

    const configs: Array<{ key: string; width: number; height: number; frameRate: number; codec: "h264" | "h265" | "vp9" | "av1" }> = [
      { key: "mp4", width: projectState.value.project.settings.width, height: projectState.value.project.settings.height, frameRate: 30, codec: "h264" },
      { key: "4k", width: 3840, height: 2160, frameRate: 30, codec: "h264" },
      { key: "1080p-high", width: 1920, height: 1080, frameRate: 30, codec: "h264" },
      { key: "1080p-60", width: 1920, height: 1080, frameRate: 60, codec: "h264" },
      { key: "prores", width: projectState.value.project.settings.width, height: projectState.value.project.settings.height, frameRate: 30, codec: "h264" },
    ];

    for (const config of configs) {
      const estimate = estimateExportTime(deviceProfile.value, {
        width: config.width,
        height: config.height,
        frameRate: config.frameRate,
        duration,
        codec: config.codec,
      });
      estimates.value.set(config.key, estimate);
    }

    exportEstimates.value = estimates;
  }
);

const runExport = async (videoSettings: Partial<VideoExportSettings>, writableStream: FileSystemWritableFileStream) => {
  const engine = getExportEngine();
  await engine.initialize();

  const generator = engine.exportVideo(projectState.value.project, videoSettings, writableStream);
  let finalResult: ExportResult | undefined;

  while (true) {
    const { value, done } = await generator.next();
    if (done) {
      finalResult = value;
      break;
    }
    exportStateLocal.value.progress = value.progress * 100;
    exportStateLocal.value.phase = formatExportPhase(value.phase);
  }

  if (finalResult?.success) {
    exportStateLocal.value.complete = true;
    exportStateLocal.value.phase = "已保存";
  } else {
    throw new Error(finalResult?.error?.message || "导出失败");
  }
};

const showSavePicker = async (filename: string, ext: string): Promise<FileSystemWritableFileStream> => {
  const mimeMap: Record<string, string> = {
    mp4: "video/mp4",
    webm: "video/webm",
    mov: "video/quicktime",
    wav: "audio/wav",
  };
  const mime = mimeMap[ext] || "application/octet-stream";

  if ("showSaveFilePicker" in window) {
    const handle = await (window as any).showSaveFilePicker({
      suggestedName: filename,
      types: [{
        description: "媒体文件",
        accept: { [mime]: [`.${ext}`] },
      }],
    });
    return handle.createWritable();
  }

  // Fallback for browsers that don't support showSaveFilePicker
  let buffer = new Uint8Array(16 * 1024 * 1024);
  let length = 0;
  let cursor = 0;

  const writeBytes = (bytes: Uint8Array, position: number) => {
    const end = position + bytes.byteLength;
    if (end > buffer.length) {
      let newSize = buffer.length;
      while (newSize < end) newSize *= 2;
      const next = new Uint8Array(newSize);
      next.set(buffer.subarray(0, length));
      buffer = next;
    }
    buffer.set(bytes, position);
    if (end > length) length = end;
    cursor = end;
  };

  const triggerDownload = () => {
    const blob = new Blob([buffer.slice(0, length)], { type: mime });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return {
    seek(position: number) {
      cursor = position;
      return Promise.resolve();
    },
    write(data: any) {
      if (data instanceof ArrayBuffer) {
        writeBytes(new Uint8Array(data), cursor);
      } else if (ArrayBuffer.isView(data)) {
        writeBytes(new Uint8Array(data.buffer, data.byteOffset, data.byteLength), cursor);
      }
      return Promise.resolve();
    },
    close() {
      triggerDownload();
      return Promise.resolve();
    },
    abort() {
      return Promise.resolve();
    },
    truncate() {
      return Promise.resolve();
    },
  } as any;
};

const handleExport = async (type: ExportType) => {
  isExportOpen.value = false;

  try {
    if (type === "cloud") {
      if (authState.value.status !== "authenticated") {
        openAuthDialog();
        toast.warning("请先登录", "登录后才能创建云端导出任务。");
        return;
      }

      exportStateLocal.value = {
        isExporting: true,
        progress: 0,
        phase: "创建云端任务",
        error: null,
        complete: false,
      };

      const project = projectState.value.project;
      const task = await createStudioCloudExportTask(project, {
        format: "mp4",
        codec: "h264",
        width: project.settings.width,
        height: project.settings.height,
        frameRate: project.settings.frameRate,
        bitrate: 12000,
        quality: 85,
      });

      syncCloudExportState(task);
      toast.success("云端导出任务已创建", `任务编号：${task.taskId}`);
      const latestTask = await pollCloudExportTask(task.taskId);
      if (latestTask?.status === "SUCCEEDED") {
        toast.success("云端导出完成", latestTask.outputUrl || "导出文件已生成");
        setTimeout(() => {
          exportStateLocal.value = { isExporting: false, progress: 0, phase: "", error: null, complete: false };
        }, 2500);
      } else if (latestTask?.status === "FAILED") {
        toast.error("云端导出失败", latestTask.errorMessage || latestTask.message || "任务执行失败");
      }
      return;
    } else if (type === "wav") {
      const writable = await showSavePicker(`${projectState.value.project.name || "export"}.wav`, "wav");

      exportStateLocal.value = {
        isExporting: true,
        progress: 0,
        phase: "初始化中",
        error: null,
        complete: false,
      };

      const engine = getExportEngine();
      await engine.initialize();

      const audioSettings: Partial<AudioExportSettings> = {
        format: "wav",
        sampleRate: 48000,
        channels: 2,
        bitDepth: 24,
      };

      const generator = engine.exportAudio(projectState.value.project, audioSettings);
      let finalResult: ExportResult | undefined;

      while (true) {
        const { value, done } = await generator.next();
        if (done) {
          finalResult = value;
          break;
        }
        exportStateLocal.value.progress = value.progress * 100;
        exportStateLocal.value.phase = formatExportPhase(value.phase);
      }

      if (finalResult?.success && finalResult.blob) {
        if ("showSaveFilePicker" in window) {
          await finalResult.blob.stream().pipeTo(writable as any);
        } else {
          const url = URL.createObjectURL(finalResult.blob);
          const a = document.createElement("a");
          a.href = url;
          a.download = `${projectState.value.project.name || "export"}.wav`;
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
          URL.revokeObjectURL(url);
        }
        exportStateLocal.value.complete = true;
        exportStateLocal.value.phase = "已保存";
      } else {
        try { await writable.abort(); } catch {}
        throw new Error(finalResult?.error?.message || "导出失败");
      }
    } else {
      const base = {
        width: projectState.value.project.settings.width,
        height: projectState.value.project.settings.height,
        frameRate: projectState.value.project.settings.frameRate,
      };

      const presets: Record<string, { settings: Partial<VideoExportSettings>; ext: string }> = {
        mp4: { settings: { ...base, format: "mp4", codec: "h264", bitrate: 12000, quality: 85 }, ext: "mp4" },
        gif: { settings: { ...base, format: "webm", codec: "vp9", bitrate: 8000 }, ext: "webm" },
        "1080p-high": { settings: { ...base, width: 1920, height: 1080, frameRate: 30, format: "mp4", codec: "h264", bitrate: 20000, quality: 95 }, ext: "mp4" },
        "1080p-60": { settings: { ...base, width: 1920, height: 1080, frameRate: 60, format: "mp4", codec: "h264", bitrate: 25000, quality: 95 }, ext: "mp4" },
        prores: { settings: { ...base, format: "mov", codec: "prores", bitrate: 220000, quality: 100 }, ext: "mov" },
      };

      const preset = presets[type] ?? presets.mp4;
      const writable = await showSavePicker(`${projectState.value.project.name || "export"}.${preset.ext}`, preset.ext);

      exportStateLocal.value = {
        isExporting: true,
        progress: 0,
        phase: "初始化中",
        error: null,
        complete: false,
      };

      await runExport(preset.settings, writable);
    }

    setTimeout(() => {
      exportStateLocal.value = { isExporting: false, progress: 0, phase: "", error: null, complete: false };
    }, 2000);
  } catch (error: any) {
    if (error.name === "AbortError") return;
    exportStateLocal.value.isExporting = false;
    exportStateLocal.value.error = error.message || "导出失败";
  }
};

const handleCancelExport = () => {
  getExportEngine().cancel();
  exportStateLocal.value = {
    isExporting: false,
    progress: 0,
    phase: "",
    error: null,
    complete: false,
  };
};

const handleCustomExport = async (settings: VideoExportSettings) => {
  isExportDialogOpen.value = false;

  try {
    const ext = settings.format === "mov" ? "mov" : settings.format === "webm" ? "webm" : "mp4";
    const writable = await showSavePicker(`${projectState.value.project.name || "export"}.${ext}`, ext);

    exportStateLocal.value = {
      isExporting: true,
      progress: 0,
      phase: "初始化中",
      error: null,
      complete: false,
    };

    const needsUpscaling =
      settings.width > projectState.value.project.settings.width ||
      settings.height > projectState.value.project.settings.height;

    const exportSettings: Partial<VideoExportSettings> = {
      ...settings,
      upscaling:
        settings.upscaling?.enabled && needsUpscaling
          ? settings.upscaling
          : undefined,
    };

    await runExport(exportSettings, writable);

    setTimeout(() => {
      exportStateLocal.value = { isExporting: false, progress: 0, phase: "", error: null, complete: false };
    }, 2000);
  } catch (error: any) {
    if (error.name === "AbortError") return;
    exportStateLocal.value.isExporting = false;
    exportStateLocal.value.error = error.message || "导出失败";
  }
};

const handleRecordingComplete = async (screenBlob: Blob, webcamBlob?: Blob) => {
  if (!screenBlob || screenBlob.size === 0) {
    toast.error("录制失败", "没有捕获到视频数据，请重试。");
    return;
  }

  const timestamp = new Date().toISOString().slice(0, 19).replace(/[:-]/g, "");
  let importCount = 0;
  const errors: string[] = [];

  const screenFile = new File([screenBlob], `Screen_${timestamp}.webm`, {
    type: screenBlob.type || "video/webm",
  });
  const screenResult = await projectStore.getState().importMedia(screenFile);
  if (screenResult.success) {
    importCount++;
  } else {
    errors.push(screenResult.error?.message || "屏幕录制导入失败");
  }

  if (webcamBlob && webcamBlob.size > 0) {
    const webcamFile = new File([webcamBlob], `Webcam_${timestamp}.webm`, {
      type: webcamBlob.type || "video/webm",
    });
    const webcamResult = await projectStore.getState().importMedia(webcamFile);
    if (webcamResult.success) {
      importCount++;
    } else {
      errors.push(webcamResult.error?.message || "摄像头录制导入失败");
    }
  }

  if (importCount > 0) {
    toast.success(
      `已导入 ${importCount} 个录制素材`,
      webcamBlob && webcamBlob.size > 0
        ? "屏幕和摄像头录制已加入素材，可在时间线中合成。"
        : "屏幕录制已加入素材。"
    );
  } else if (errors.length > 0) {
    toast.error("导入失败", errors.join("。"));
  }
};

const toggleTheme = () => {
  themeStore.getState().toggleTheme();
};

const projectRes = computed(() => `${projectState.value.project.settings.width}×${projectState.value.project.settings.height}`);
const aspectRatio = computed(() => projectState.value.project.settings.width / projectState.value.project.settings.height);
const isVertical = computed(() => aspectRatio.value < 0.9);

type ExportType =
  | "cloud"
  | "mp4"
  | "gif"
  | "wav"
  | "1080p-high"
  | "1080p-60"
  | "prores";

const exportOptions = computed(() => {
  const options: Array<{
    label: string;
    icon: any;
    desc: string;
    type: ExportType;
    recommended?: boolean;
    separator?: boolean;
  }> = [
    {
      label: "MP4 标准",
      icon: ZapIcon,
      desc: `${projectRes.value} H.264 - 网页与社交平台`,
      type: "mp4",
      recommended: true,
    },
    {
      label: "云端导出任务",
      icon: UploadIcon,
      desc: "提交到媒体节点，后续由 FFmpeg 渲染",
      type: "cloud",
    },
    {
      label: "",
      icon: FilmIcon,
      desc: "",
      type: "mp4",
      separator: true,
    },
    {
      label: "1080p 高质量",
      icon: FileVideoIcon,
      desc: "1920×1080 30fps - 高码率",
      type: "1080p-high",
    },
    {
      label: "1080p 60fps",
      icon: FileVideoIcon,
      desc: "1920×1080 - 流畅播放",
      type: "1080p-60",
    },
    {
      label: "仅音频 (WAV)",
      icon: MusicIcon,
      desc: "未压缩音频",
      type: "wav",
    },
  ];
  return options;
});
</script>

<style scoped>
/* Scoped styles if needed */
</style>
