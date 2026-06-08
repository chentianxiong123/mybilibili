<template>
  <div class="flex flex-col h-full bg-bg text-text-primary overflow-hidden">
    <!-- Hidden file input -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      accept="video/*,audio/*,image/*"
      class="hidden"
      @change="handleFileInputChange"
    />

    <!-- Loading overlay -->
    <div
      v-if="isImporting"
      class="absolute inset-0 bg-background-secondary/90 backdrop-blur-sm flex flex-col items-center justify-center z-50"
    >
      <div class="w-10 h-10 border-2 border-primary border-t-transparent rounded-full animate-spin mb-3" />
      <p class="text-sm text-text-secondary">{{ importProgress || "Importing..." }}</p>
    </div>

    <!-- Tab navigation -->
    <div class="flex items-center gap-1 px-2 pt-2 pb-1 border-b border-border/70 shrink-0">
      <button
        v-for="tab in assetsTabs"
        :key="tab.value"
        @click="setActiveTab(tab.value)"
        :class="[
          'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-[11px] font-medium transition-all',
          activeTab === tab.value
            ? 'bg-primary/10 text-primary'
            : 'text-text-muted hover:text-text-secondary hover:bg-background-tertiary'
        ]"
        :title="tab.description"
      >
        <component :is="tab.icon" :size="13" />
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <!-- Tab content -->
    <div class="flex-1 min-h-0 flex flex-col border-t border-border/70">
      <!-- Media tab -->
      <template v-if="activeTab === 'media'">
        <div class="px-3 pt-2.5 pb-2 flex items-center gap-2 shrink-0">
          <div class="relative flex-1">
            <SearchIcon :size="14" class="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted z-10" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索媒体"
              class="w-full pl-8 pr-3 py-1.5 text-xs bg-background-tertiary border border-border rounded-md text-text-primary h-8 focus:outline-none focus:border-primary/50"
            />
          </div>
          <div class="flex items-center bg-background-tertiary border border-border rounded-md p-0.5">
            <button
              v-for="mode in viewModes"
              :key="mode.key"
              @click="mediaViewMode = mode.key"
              :title="mode.title"
              :class="[
                'p-1 rounded transition-colors',
                mediaViewMode === mode.key
                  ? 'bg-background-elevated text-text-primary'
                  : 'text-text-muted hover:text-text-secondary'
              ]"
            >
              <component :is="mode.icon" :size="13" />
            </button>
          </div>
        </div>

        <!-- Missing assets warning -->
        <div v-if="missingAssetsCount > 0" class="px-3 pb-2 space-y-2 shrink-0">
          <button
            @click="showOnlyMissing = !showOnlyMissing"
            :class="[
              'w-full px-3 py-1.5 rounded-lg border text-[11px] font-medium transition-all flex items-center justify-between',
              showOnlyMissing
                ? 'bg-yellow-500/10 border-yellow-500 text-yellow-500'
                : 'bg-background-tertiary border-border text-text-secondary hover:border-yellow-500/50'
            ]"
          >
            <div class="flex items-center gap-1.5">
              <AlertTriangleIcon :size="13" />
              <span>仅显示缺失素材</span>
            </div>
            <div class="px-1.5 py-0.5 rounded-full bg-yellow-500 text-black text-[10px] font-bold">
              {{ missingAssetsCount }}
            </div>
          </button>
          <button
            @click="handleRelinkFromFolder"
            class="w-full px-3 py-1.5 rounded-lg border border-yellow-500/40 bg-yellow-500/5 text-yellow-500 text-[11px] font-medium transition-all hover:bg-yellow-500/15 flex items-center gap-1.5"
          >
            <RefreshCwIcon :size="13" />
            <span>从文件夹重新关联...</span>
          </button>
        </div>

        <div v-if="cloudMissingCount > 0" class="px-3 pb-2 shrink-0">
          <div class="px-3 py-1.5 rounded-lg border border-primary/30 bg-primary/5 text-primary text-[11px] font-medium flex items-center justify-between">
            <span>云端导出素材准备中</span>
            <span>{{ cloudReadyCount }}/{{ cloudTotalCount }}</span>
          </div>
        </div>

        <!-- Media grid / list -->
        <div
          class="flex-1 min-h-0 overflow-y-auto px-3 pb-3 relative"
          :class="{ 'bg-primary/5': isDragOver }"
          @dragover.prevent="handleDragOver"
          @dragleave="handleDragLeave"
          @drop.prevent="handleDrop"
        >
          <template v-if="filteredItems.length === 0">
            <div class="flex-1 flex flex-col items-center justify-center p-6 text-center">
              <div class="w-14 h-14 rounded-xl bg-background-tertiary border border-border flex items-center justify-center mb-3 shadow-inner">
                <UploadIcon :size="22" class="text-text-muted" />
              </div>
              <p class="text-xs text-text-secondary mb-1.5 font-medium">尚未导入媒体</p>
              <p class="text-[10px] text-text-muted mb-4">拖拽文件到此处，或点击导入</p>
              <button
                @click="triggerFileInput"
                class="px-3 py-1.5 bg-background-elevated hover:bg-background-tertiary border border-border text-text-primary text-[11px] font-medium rounded-md transition-all hover:border-primary/50"
              >
                导入媒体
              </button>
            </div>
          </template>
          <template v-else>
            <div
              :class="[
                mediaViewMode === 'list'
                  ? 'flex flex-col gap-1'
                  : mediaViewMode === 'small'
                    ? 'grid grid-cols-3 gap-1.5'
                    : 'grid grid-cols-2 gap-2'
              ]"
            >
              <template v-for="item in filteredItems" :key="item.id">
                <!-- Grid / large / small view -->
                <div
                  v-if="mediaViewMode !== 'list'"
                  draggable
                  @click="handleSelectItem(item.id)"
                  @dblclick.prevent="handleAddToTimeline(item)"
                  @mouseenter="hoveredItemId = item.id"
                  @mouseleave="hoveredItemId = null"
                  @dragstart="handleItemDragStart($event, item)"
                  :class="[
                    'flex flex-col rounded-lg border-2 cursor-pointer transition-all overflow-hidden shadow-sm group',
                    borderClass(item)
                  ]"
                >
                  <div class="aspect-video bg-background-tertiary relative overflow-hidden">
                    <img
                      v-if="item.thumbnailUrl"
                      :src="item.thumbnailUrl"
                      :alt="item.name"
                      class="w-full h-full object-cover"
                    />
                    <div
                      v-else
                      class="absolute inset-0 flex items-center justify-center bg-background-tertiary"
                    >
                      <component :is="getItemIcon(item.type)" :size="mediaViewMode === 'small' ? 16 : 24" :class="iconColorClass(item.type)" />
                    </div>

                    <!-- Audio waveform placeholder -->
                    <div v-if="item.type === 'audio'" class="absolute top-1/2 left-0 right-0 h-3 flex items-center gap-px px-2 -translate-y-1/2">
                      <div v-for="i in 10" :key="i" class="flex-1 bg-primary/30 rounded-full" :style="{ height: `${Math.random() * 100}%` }" />
                    </div>

                    <!-- Badges -->
                    <div v-if="item.kieaiError" class="absolute top-1 left-1 px-1 py-0.5 bg-red-500 rounded text-[7px] text-white font-bold flex items-center gap-0.5">
                      <AlertTriangleIcon :size="7" /> 失败
                    </div>
                    <div v-else-if="item.isPending" class="absolute top-1 left-1 px-1 py-0.5 bg-purple-500 rounded text-[7px] text-white font-bold flex items-center gap-0.5">
                      <div class="h-1.5 w-1.5 animate-spin rounded-full border border-white border-t-transparent" /> AI
                    </div>
                    <div v-else-if="item.isPlaceholder" class="absolute top-1 left-1 px-1 py-0.5 bg-yellow-500 rounded text-[7px] text-black font-bold flex items-center gap-0.5">
                      <AlertTriangleIcon :size="8" /> 缺失
                    </div>
                    <div v-else-if="item.cloudUploadError" class="absolute top-1 left-1 px-1 py-0.5 bg-red-500 rounded text-[7px] text-white font-bold">
                      上云失败
                    </div>
                    <div v-else-if="item.cloudObjectKey" class="absolute top-1 left-1 px-1 py-0.5 bg-primary rounded text-[7px] text-black font-bold">
                      已上云
                    </div>
                    <div v-else class="absolute top-1 left-1 px-1 py-0.5 bg-background-elevated/90 border border-border rounded text-[7px] text-text-muted font-bold">
                      未上云
                    </div>

                    <!-- Duration badge -->
                    <div v-if="item.metadata?.duration" class="absolute bottom-1 right-1 px-1 py-0.5 bg-black/70 rounded text-[8px] text-white font-mono">
                      {{ formatDuration(item.metadata.duration) }}
                    </div>

                    <!-- Overlays -->
                    <div v-if="item.kieaiError && hoveredItemId !== item.id" class="absolute inset-0 flex items-center justify-center bg-red-500/10">
                      <AlertTriangleIcon :size="mediaViewMode === 'small' ? 18 : 28" class="text-red-400/60" />
                    </div>
                    <div v-else-if="item.isPending && hoveredItemId !== item.id" class="absolute inset-0 flex items-center justify-center bg-purple-500/10">
                      <div class="h-7 w-7 animate-spin rounded-full border-4 border-purple-400 border-t-transparent" />
                    </div>
                    <div v-else-if="item.isPlaceholder && hoveredItemId !== item.id" class="absolute inset-0 flex items-center justify-center bg-yellow-500/10">
                      <AlertTriangleIcon :size="mediaViewMode === 'small' ? 18 : 28" class="text-yellow-500/50" />
                    </div>

                    <!-- Hover overlay -->
                    <div
                      v-if="hoveredItemId === item.id"
                      class="absolute inset-0 bg-black/40 backdrop-blur-[1px] flex items-center justify-center gap-1.5 animate-in fade-in duration-200"
                    >
                      <button
                        v-if="item.type === 'image' && !item.isPending && !item.kieaiError"
                        @click.stop="handleOpenKieAI(item)"
                        title="用 KieAI 创建"
                        class="p-1.5 bg-purple-500/20 rounded-full hover:bg-purple-500/40 backdrop-blur-sm transition-colors"
                      >
                        <SparklesIcon :size="12" class="text-purple-300" />
                      </button>
                      <button
                        @click.stop="handleAddToTimeline(item)"
                        title="添加到时间线"
                        class="p-1.5 bg-primary/20 rounded-full hover:bg-primary/40 backdrop-blur-sm transition-colors"
                      >
                        <PlusIcon :size="12" class="text-primary" />
                      </button>
                      <button
                        v-if="item.kieaiError"
                        @click.stop="handleRetryKieAI(item)"
                        title="重试生成"
                        class="p-1.5 bg-red-500/20 rounded-full hover:bg-red-500/40 backdrop-blur-sm transition-colors"
                      >
                        <RefreshCwIcon :size="12" class="text-red-400" />
                      </button>
                      <button
                        v-else-if="item.isPending"
                        class="p-1.5"
                        title="正在生成..."
                      >
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-purple-400 border-t-transparent" />
                      </button>
                      <button
                        v-else-if="item.isPlaceholder"
                        @click.stop="handleReplaceAsset(item.id)"
                        title="替换素材"
                        class="p-1.5 bg-yellow-500/20 rounded-full hover:bg-yellow-500/40 backdrop-blur-sm transition-colors"
                      >
                        <RefreshCwIcon :size="12" class="text-yellow-500" />
                      </button>
                      <button
                        @click.stop="handleDeleteItem(item.id)"
                        title="删除"
                        class="p-1.5 bg-red-500/20 rounded-full hover:bg-red-500/40 backdrop-blur-sm transition-colors"
                      >
                        <Trash2Icon :size="12" class="text-red-400" />
                      </button>
                    </div>

                    <!-- Selection indicator -->
                    <div
                      v-if="isSelected(item.id)"
                      class="absolute top-1 right-1 w-1.5 h-1.5 bg-primary rounded-full shadow-[0_0_6px_#22c55e]"
                    />
                  </div>

                  <!-- Metadata -->
                  <div class="mt-1 px-0.5">
                    <div
                      :class="['text-[9px] truncate font-medium', isSelected(item.id) ? 'text-primary' : 'text-text-primary']"
                      :title="item.name"
                    >
                      {{ item.name }}
                    </div>
                    <div v-if="mediaViewMode === 'large'" class="flex items-center gap-1 text-[8px] text-text-muted mt-0.5">
                      <span v-if="formatResolution(item)">{{ formatResolution(item) }}</span>
                      <span v-if="formatResolution(item) && formatFileSize(item.metadata?.fileSize)">•</span>
                      <span v-if="formatFileSize(item.metadata?.fileSize)">{{ formatFileSize(item.metadata?.fileSize) }}</span>
                    </div>
                  </div>
                </div>

                <!-- List view -->
                <div
                  v-else
                  draggable
                  @click="handleSelectItem(item.id)"
                  @dblclick.prevent="handleAddToTimeline(item)"
                  @mouseenter="hoveredItemId = item.id"
                  @mouseleave="hoveredItemId = null"
                  @dragstart="handleItemDragStart($event, item)"
                  :class="[
                    'flex items-center gap-2 px-2 py-1.5 rounded-lg border-2 cursor-pointer transition-all group',
                    borderClass(item)
                  ]"
                >
                  <div class="w-10 h-7 rounded bg-background-tertiary relative overflow-hidden flex-shrink-0">
                    <img v-if="item.thumbnailUrl" :src="item.thumbnailUrl" :alt="item.name" class="w-full h-full object-cover" />
                    <div v-else class="w-full h-full flex items-center justify-center">
                      <component :is="getItemIcon(item.type)" :size="12" :class="iconColorClass(item.type)" />
                    </div>
                    <div v-if="item.kieaiError" class="absolute inset-0 flex items-center justify-center bg-red-500/10">
                      <AlertTriangleIcon :size="10" class="text-red-400" />
                    </div>
                    <div v-else-if="item.isPending" class="absolute inset-0 flex items-center justify-center bg-purple-500/10">
                      <div class="h-3 w-3 animate-spin rounded-full border-2 border-purple-400 border-t-transparent" />
                    </div>
                    <div v-else-if="item.isPlaceholder" class="absolute inset-0 flex items-center justify-center bg-yellow-500/10">
                      <AlertTriangleIcon :size="10" class="text-yellow-500/70" />
                    </div>
                  </div>

                  <div class="flex-1 min-w-0">
                    <div :class="['text-[10px] truncate font-medium', isSelected(item.id) ? 'text-primary' : 'text-text-primary']" :title="item.name">
                      {{ item.name }}
                    </div>
                    <div class="flex items-center gap-1 text-[8px] text-text-muted">
                      <span v-if="item.metadata?.duration">{{ formatDuration(item.metadata.duration) }}</span>
                      <span v-if="item.metadata?.duration && formatResolution(item)">•</span>
                      <span v-if="formatResolution(item)">{{ formatResolution(item) }}</span>
                      <span v-if="(item.metadata?.duration || formatResolution(item)) && formatFileSize(item.metadata?.fileSize)">•</span>
                      <span v-if="formatFileSize(item.metadata?.fileSize)">{{ formatFileSize(item.metadata?.fileSize) }}</span>
                      <span v-if="cloudStatusLabel(item)">•</span>
                      <span v-if="cloudStatusLabel(item)">{{ cloudStatusLabel(item) }}</span>
                    </div>
                  </div>

                  <!-- Hover actions for list -->
                  <div v-if="hoveredItemId === item.id" class="flex items-center gap-0.5 flex-shrink-0">
                    <button v-if="item.type === 'image' && !item.isPending && !item.kieaiError" @click.stop="handleOpenKieAI(item)" title="用 KieAI 创建" class="p-1 bg-purple-500/20 rounded hover:bg-purple-500/40 transition-colors">
                      <SparklesIcon :size="10" class="text-purple-300" />
                    </button>
                    <button @click.stop="handleAddToTimeline(item)" title="添加到时间线" class="p-1 bg-primary/20 rounded hover:bg-primary/40 transition-colors">
                      <PlusIcon :size="10" class="text-primary" />
                    </button>
                    <button v-if="item.kieaiError" @click.stop="handleRetryKieAI(item)" title="重试生成" class="p-1 bg-red-500/20 rounded hover:bg-red-500/40 transition-colors">
                      <RefreshCwIcon :size="10" class="text-red-400" />
                    </button>
                    <button v-else-if="item.isPending" class="p-1" title="正在生成...">
                      <div class="h-2.5 w-2.5 animate-spin rounded-full border-2 border-purple-400 border-t-transparent" />
                    </button>
                    <button v-else-if="item.isPlaceholder" @click.stop="handleReplaceAsset(item.id)" title="替换素材" class="p-1 bg-yellow-500/20 rounded hover:bg-yellow-500/40 transition-colors">
                      <RefreshCwIcon :size="10" class="text-yellow-500" />
                    </button>
                    <button @click.stop="handleDeleteItem(item.id)" title="删除" class="p-1 bg-red-500/20 rounded hover:bg-red-500/40 transition-colors">
                      <Trash2Icon :size="10" class="text-red-400" />
                    </button>
                  </div>

                  <div v-if="isSelected(item.id)" class="w-1.5 h-1.5 bg-primary rounded-full shadow-[0_0_6px_#22c55e] flex-shrink-0" />
                </div>
              </template>

              <!-- 添加媒体按钮 -->
              <template v-if="mediaViewMode === 'list'">
                <button
                  @click="triggerFileInput"
                  class="flex items-center gap-2 px-2 py-1.5 rounded-lg border-2 border-dashed border-border hover:border-text-secondary cursor-pointer transition-all group"
                >
                  <div class="w-9 h-6 rounded bg-background-tertiary flex items-center justify-center flex-shrink-0">
                    <UploadIcon :size="12" class="text-text-muted group-hover:text-text-secondary transition-colors" />
                  </div>
                  <span class="text-[10px] text-text-muted group-hover:text-text-secondary transition-colors font-medium">添加媒体</span>
                </button>
              </template>
              <template v-else>
                <div class="flex flex-col">
                  <button
                    @click="triggerFileInput"
                    class="aspect-video bg-background-tertiary rounded-lg border-2 border-dashed border-border hover:border-text-secondary relative flex items-center justify-center cursor-pointer transition-all overflow-hidden shadow-sm group"
                  >
                    <div class="flex flex-col items-center gap-1">
                      <UploadIcon :size="mediaViewMode === 'small' ? 14 : 18" class="text-text-muted group-hover:text-text-secondary transition-colors" />
                      <span class="text-[9px] text-text-muted group-hover:text-text-secondary transition-colors">添加媒体</span>
                    </div>
                  </button>
                </div>
              </template>
            </div>
          </template>

          <!-- Drag overlay -->
          <div
            v-if="isDragOver"
            class="absolute inset-3 border-2 border-dashed border-primary rounded-xl flex items-center justify-center bg-primary/5 pointer-events-none z-50 backdrop-blur-sm"
          >
            <div class="text-primary text-xs font-bold bg-background-secondary px-3 py-1.5 rounded-full shadow-lg">
              Drop files to import
            </div>
          </div>
        </div>
      </template>

      <!-- Graphics tab -->
      <template v-else-if="activeTab === 'graphics'">
        <div class="flex-1 min-h-0 overflow-y-auto">
          <div class="px-3 py-3">
            <div class="mb-5">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-[11px] font-medium text-text-secondary flex items-center gap-1.5">
                  <PaletteIcon :size="11" />
                  Backgrounds
                </h4>
              </div>
              <div class="flex gap-1 mb-2 flex-wrap">
                <button
                  v-for="cat in backgroundCategories"
                  :key="cat"
                  @click="backgroundCategory = cat"
                  :class="[
                    'px-2 py-0.5 text-[9px] rounded-md transition-all capitalize',
                    backgroundCategory === cat
                      ? 'bg-primary text-white'
                      : 'bg-background-tertiary text-text-muted hover:text-text-secondary'
                  ]"
                >
                  {{ cat }}
                </button>
              </div>
              <div class="grid grid-cols-4 gap-1.5">
                <button
                  v-for="preset in filteredBackgrounds"
                  :key="preset.id"
                  @click="handleImportBackground(preset)"
                  :disabled="generatingBackground !== null"
                  :class="[
                    'aspect-square rounded-lg border border-border hover:border-primary/50 transition-all overflow-hidden relative group',
                    { 'opacity-50': generatingBackground !== null }
                  ]"
                  :title="preset.name"
                  :style="{ background: preset.thumbnail }"
                >
                  <div v-if="generatingBackground === preset.id" class="absolute inset-0 bg-black/50 flex items-center justify-center">
                    <div class="h-3.5 w-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  </div>
                  <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                    <PlusIcon :size="14" class="text-white" />
                  </div>
                  <span class="absolute bottom-0 inset-x-0 text-[7px] text-white bg-black/60 py-0.5 px-1 truncate opacity-0 group-hover:opacity-100 transition-opacity">
                    {{ preset.name }}
                  </span>
                </button>
              </div>
            </div>

            <div>
              <h4 class="text-[11px] font-medium text-text-secondary mb-2">Shapes</h4>
              <div class="grid grid-cols-4 gap-1.5">
                <button
                  v-for="shape in shapePresets"
                  :key="shape.type"
                  @click="handleAddShape(shape.type)"
                  class="aspect-square bg-background-tertiary rounded-lg border border-border hover:border-primary/50 hover:bg-primary/5 transition-all flex flex-col items-center justify-center gap-0.5 group"
                  :title="shape.label"
                >
                  <component :is="shape.icon" :size="18" class="text-text-secondary group-hover:text-primary transition-colors" />
                  <span class="text-[8px] text-text-muted group-hover:text-text-secondary">{{ shape.label }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Other tabs: text, effects, transitions, ai, recipes, templates -->
      <template v-else>
        <div class="flex-1 min-h-0 overflow-y-auto">
          <ReactAdapter :component="getTabComponent(activeTab)" />
        </div>
      </template>
    </div>

    <!-- KieAI Image Dialog (React) -->
    <ReactAdapter
      v-if="kieaiDialog"
      :component="KieAIImageDialogRaw"
      :componentProps="{
        file: kieaiDialog.file,
        previewUrl: kieaiDialog.previewUrl,
        onClose: () => kieaiDialog = null
      }"
    />

    <!-- Aspect Ratio Match Dialog (React) -->
    <ReactAdapter
      v-if="showAspectRatioDialog"
      :component="AspectRatioMatchDialogRaw"
      :componentProps="{
        isOpen: showAspectRatioDialog,
        videoWidth: aspectRatioDialogData?.videoWidth ?? 0,
        videoHeight: aspectRatioDialogData?.videoHeight ?? 0,
        onConfirm: handleConfirmAspectRatioMatch,
        onCancel: handleCancelAspectRatioMatch
      }"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, markRaw } from "vue";
import {
  Search as SearchIcon,
  Upload as UploadIcon,
  Film,
  Music,
  Image as ImageIcon,
  Plus as PlusIcon,
  Trash2 as Trash2Icon,
  Square,
  Circle,
  Triangle,
  Star,
  ArrowRight,
  Hexagon,
  AlertTriangle as AlertTriangleIcon,
  RefreshCw as RefreshCwIcon,
  Palette as PaletteIcon,
  LayoutGrid,
  Grid2x2,
  List,
  Sparkles as SparklesIcon,
  Video,
  Type,
  Shapes,
  Wand2,
  LayoutTemplate,
  Zap,
  Shuffle,
} from "lucide-vue-next";

import ReactAdapter from "../ReactAdapter.vue";
import { useZustandStore } from "../../hooks/useZustandStore";
import { useProjectStore } from "../../stores/project-store";
import { useUIStore } from "../../stores/ui-store";
import { useKieAIStore } from "../../stores/kieai-store";
import { useTtsAudioStore } from "../../stores/tts-store";
import { toast } from "../../stores/notification-store";
import { saveFileHandle, saveDirectoryHandle, loadMediaBlob } from "../../services/media-storage";
import {
  BACKGROUND_PRESETS,
  generateBackgroundBlob,
  type BackgroundPreset,
} from "../../services/background-generator";
import type { MediaItem, ShapeType } from "@mybilibili-studio/core";

// React sub-components
import { AIGenTab } from "./AIGenTab";
import { RecipesTab } from "./panels/RecipesTab";
import { TemplatesTab } from "./panels/TemplatesTab";
import { EffectsPanel, TransitionsPanel } from "./panels/EffectsTransitionsPanel";
import { KieAIImageDialog } from "./kieai/KieAIImageDialog";
import { AspectRatioMatchDialog } from "./dialogs/AspectRatioMatchDialog";

const AIGenTabRaw = markRaw(AIGenTab);
const RecipesTabRaw = markRaw(RecipesTab);
const TemplatesTabRaw = markRaw(TemplatesTab);
const EffectsPanelRaw = markRaw(EffectsPanel);
const TransitionsPanelRaw = markRaw(TransitionsPanel);
const KieAIImageDialogRaw = markRaw(KieAIImageDialog);
const AspectRatioMatchDialogRaw = markRaw(AspectRatioMatchDialog);

// --- Store bindings ---
const projectStore = useProjectStore;
const uiStore = useUIStore;
const kieaiStore = useKieAIStore;

const projectState = useZustandStore<any>(projectStore);
const uiState = useZustandStore<any>(uiStore);
const kieaiState = useZustandStore<any>(kieaiStore);
const ttsState = useTtsAudioStore;

// --- Local state ---
const fileInputRef = ref<HTMLInputElement | null>(null);
const searchQuery = ref("");
const activeTab = ref<AssetsTab>("media");
const isDragOver = ref(false);
const isImporting = ref(false);
const importProgress = ref("");
const showOnlyMissing = ref(false);
const showAspectRatioDialog = ref(false);
const aspectRatioDialogData = ref<{ videoWidth: number; videoHeight: number; itemToAdd: MediaItem } | null>(null);
const mediaViewMode = ref<MediaViewMode>("large");
const generatingBackground = ref<string | null>(null);
const backgroundCategory = ref<"all" | "solid" | "gradient" | "pattern" | "mesh">("all");
const kieaiDialog = ref<{ file: File; previewUrl: string | null } | null>(null);
const hoveredItemId = ref<string | null>(null);

// --- Computed ---
const mediaItems = computed(() => projectState.value.project.mediaLibrary.items);
const missingAssetsCount = computed(() => mediaItems.value.filter((i: MediaItem) => i.isPlaceholder).length);
const cloudRelevantItems = computed(() => mediaItems.value.filter((i: MediaItem) => !i.isPlaceholder));
const cloudTotalCount = computed(() => cloudRelevantItems.value.length);
const cloudReadyCount = computed(() => cloudRelevantItems.value.filter((i: MediaItem) => Boolean(i.cloudObjectKey)).length);
const cloudMissingCount = computed(() => Math.max(0, cloudTotalCount.value - cloudReadyCount.value));
const filteredItems = computed(() => {
  const q = searchQuery.value.toLowerCase();
  return mediaItems.value.filter((item: MediaItem) => {
    const matchesSearch = item.name.toLowerCase().includes(q);
    const matchesFilter = showOnlyMissing.value ? item.isPlaceholder : true;
    return matchesSearch && matchesFilter;
  });
});

const ttsHasUnsaved = computed(() => {
  const s = ttsState();
  return s.generatedAudio !== null && !s.isAudioSaved;
});

watch(activeTab, (tab) => {
  if (tab === "ai" && ttsHasUnsaved.value) {
    toast.warning("未保存音频已丢弃", "下次请先保存到素材库或下载。");
  }
});

// --- Tab definitions ---
type AssetsTab = "media" | "text" | "graphics" | "effects" | "transitions" | "ai" | "recipes" | "templates";

const assetsTabs = [
  { value: "media" as AssetsTab, label: "媒体", description: "导入视频、音频和图片素材。", icon: Video },
  { value: "text" as AssetsTab, label: "文字", description: "添加标题预设和字幕元素。", icon: Type },
  { value: "graphics" as AssetsTab, label: "图形", description: "创建形状、箭头和 SVG 叠加层。", icon: Shapes },
  { value: "effects" as AssetsTab, label: "效果", description: "将效果拖到片段上即可应用。", icon: Zap },
  { value: "transitions" as AssetsTab, label: "转场", description: "将转场拖到片段边缘即可应用。", icon: Shuffle },
  { value: "ai" as AssetsTab, label: "AI 生成", description: "生成片段、字幕和辅助剪辑内容。", icon: SparklesIcon },
  { value: "recipes" as AssetsTab, label: "方案", description: "应用片段级外观、叠加层和文字组合。", icon: Wand2 },
  { value: "templates" as AssetsTab, label: "项目模板", description: "加载完整项目起始布局和预设。", icon: LayoutTemplate },
];

const viewModes = [
  { key: "large" as MediaViewMode, icon: LayoutGrid, title: "大图标" },
  { key: "small" as MediaViewMode, icon: Grid2x2, title: "小图标" },
  { key: "list" as MediaViewMode, icon: List, title: "列表视图" },
];

type MediaViewMode = "large" | "small" | "list";

const backgroundCategories = ["all", "solid", "gradient", "mesh", "pattern"] as const;

const shapePresets = [
  { type: "rectangle" as ShapeType, icon: Square, label: "矩形" },
  { type: "circle" as ShapeType, icon: Circle, label: "圆形" },
  { type: "triangle" as ShapeType, icon: Triangle, label: "三角形" },
  { type: "star" as ShapeType, icon: Star, label: "星形" },
  { type: "arrow" as ShapeType, icon: ArrowRight, label: "箭头" },
  { type: "polygon" as ShapeType, icon: Hexagon, label: "多边形" },
];

// --- Helpers ---
function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
}

function formatResolution(item: MediaItem): string | null {
  if (item.metadata?.width && item.metadata?.height) return `${item.metadata.width}×${item.metadata.height}`;
  return null;
}

function formatFileSize(bytes?: number): string | null {
  if (!bytes) return null;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function cloudStatusLabel(item: MediaItem): string | null {
  if (item.isPlaceholder) return null;
  if (item.cloudUploadError) return "上云失败";
  if (item.cloudObjectKey) return "已上云";
  return "未上云";
}

function getItemIcon(type: string) {
  switch (type) {
    case "video": return Film;
    case "audio": return Music;
    case "image": return ImageIcon;
    default: return Film;
  }
}

function iconColorClass(type: string) {
  if (type === "audio" || type === "image") return "text-primary/50";
  return "text-status-info/50";
}

function borderClass(item: MediaItem) {
  if (item.kieaiError) return "border-red-500 ring-1 ring-red-500/50 shadow-[0_0_10px_rgba(239,68,68,0.3)]";
  if (item.isPending) return "border-purple-500 ring-1 ring-purple-500/50 shadow-[0_0_10px_rgba(168,85,247,0.3)]";
  if (item.isPlaceholder) return "border-yellow-500 ring-1 ring-yellow-500/50 shadow-[0_0_10px_rgba(234,179,8,0.3)]";
  if (uiState.value.isSelected(item.id)) return "border-primary ring-1 ring-primary/50 shadow-[0_0_10px_rgba(34,197,94,0.2)]";
  return "border-border hover:border-text-secondary";
}

const filteredBackgrounds = computed(() => {
  return BACKGROUND_PRESETS.filter((p) => backgroundCategory.value === "all" || p.category === backgroundCategory.value);
});

// --- Tab component resolver ---
function getTabComponent(tab: AssetsTab) {
  switch (tab) {
    case "text": return markRaw(AIGenTabRaw);
    case "effects": return markRaw(EffectsPanelRaw);
    case "transitions": return markRaw(TransitionsPanelRaw);
    case "ai": return markRaw(AIGenTabRaw);
    case "recipes": return markRaw(RecipesTabRaw);
    case "templates": return markRaw(TemplatesTabRaw);
    default: return markRaw(AIGenTabRaw);
  }
}

// --- Actions ---
function setActiveTab(tab: AssetsTab) {
  if (activeTab.value === "ai" && tab !== "ai" && ttsHasUnsaved.value) {
    toast.warning("未保存音频已丢弃", "下次请先保存到素材库或下载。");
  }
  activeTab.value = tab;
}

function triggerFileInput() {
  fileInputRef.value?.click();
}

function handleFileInputChange(e: Event) {
  const target = e.target as HTMLInputElement;
  handleFileImport(target.files);
  target.value = "";
}

async function handleFileImport(files: FileList | null) {
  if (!files || files.length === 0) return;
  isImporting.value = true;
  const fileArray = Array.from(files);
  try {
    for (let i = 0; i < fileArray.length; i++) {
      const file = fileArray[i];
      importProgress.value = `正在导入 ${file.name} (${i + 1}/${fileArray.length})...`;
      await projectStore.getState().importMedia(file);
    }
  } catch (error) {
    console.error("Import failed:", error);
  } finally {
    isImporting.value = false;
    importProgress.value = "";
  }
}

async function handleDrop(e: DragEvent) {
  isDragOver.value = false;
  const droppedFiles = e.dataTransfer?.files;
  if (!droppedFiles) return;

  const handlePromises =
    "getAsFileSystemHandle" in DataTransferItem.prototype
      ? Array.from(e.dataTransfer?.items || [])
          .filter((item) => item.kind === "file")
          .map(async (item) => {
            try {
              const handle = await (item as DataTransferItem & { getAsFileSystemHandle(): Promise<FileSystemHandle> }).getAsFileSystemHandle();
              if (handle.kind === "file") {
                const fileHandle = handle as FileSystemFileHandle;
                const file = await fileHandle.getFile();
                await saveFileHandle(file.name, file.size, fileHandle);
              }
            } catch { /* best-effort */ }
          })
      : [];

  await Promise.all(handlePromises);
  await handleFileImport(droppedFiles);
}

function handleDragOver(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = true;
}

function handleDragLeave() {
  isDragOver.value = false;
}

function handleSelectItem(itemId: string) {
  uiStore.getState().select({ type: "clip", id: itemId });
}

async function handleDeleteItem(itemId: string) {
  await projectStore.getState().deleteMedia(itemId);
}

async function handleReplaceAsset(itemId: string) {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "video/*,audio/*,image/*";
  input.onchange = async (e) => {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (file) {
      isImporting.value = true;
      importProgress.value = "正在替换素材...";
      try {
        await projectStore.getState().replaceMediaAsset(itemId, file);
      } catch (err) {
        console.error("Asset replacement failed:", err);
      } finally {
        isImporting.value = false;
        importProgress.value = "";
      }
    }
  };
  input.click();
}

async function handleRelinkFromFolder() {
  if (!("showDirectoryPicker" in window)) {
    toast.error("不支持选择文件夹", "请使用每个缺失素材上的刷新按钮逐个重新关联。");
    return;
  }
  let dirHandle: FileSystemDirectoryHandle;
  try {
    dirHandle = await (window as unknown as { showDirectoryPicker: () => Promise<FileSystemDirectoryHandle> }).showDirectoryPicker();
  } catch {
    return;
  }

  const { project: currentProject } = projectStore.getState();
  const placeholders = currentProject.mediaLibrary.items.filter((item: MediaItem) => item.isPlaceholder);
  if (placeholders.length === 0) return;

  try { await saveDirectoryHandle(currentProject.id, dirHandle); } catch { /* best-effort */ }

  const fileMap = new Map<string, { file: File; handle: FileSystemFileHandle }>();
  for await (const [, fh] of (dirHandle as unknown as { entries(): AsyncIterableIterator<[string, FileSystemHandle]> }).entries()) {
    if (fh.kind === "file") {
      const fileHandle = fh as FileSystemFileHandle;
      const file = await fileHandle.getFile();
      fileMap.set(`${file.name.toLowerCase()}:${file.size}`, { file, handle: fileHandle });
    }
  }

  isImporting.value = true;
  let linked = 0;
  for (const item of placeholders) {
    const key = item.sourceFile ? `${item.sourceFile.name.toLowerCase()}:${item.sourceFile.size}` : null;
    const entry = key ? fileMap.get(key) : null;
    if (entry) {
      importProgress.value = `正在重新关联 ${item.name}...`;
      try {
        try { await saveFileHandle(entry.file.name, entry.file.size, entry.handle); } catch { /* best-effort */ }
        await projectStore.getState().replaceMediaAsset(item.id, entry.file, dirHandle.name);
        linked++;
      } catch (err) {
        console.error(`[AssetsPanel] Failed to relink ${item.name}:`, err);
      }
    }
  }
  isImporting.value = false;
  importProgress.value = "";

  if (linked > 0) {
    toast.success(`已重新关联 ${linked}/${placeholders.length} 个素材`);
  } else {
    toast.error("未找到匹配文件", "所选文件夹中没有文件名匹配缺失素材。");
  }
}

function handleItemDragStart(e: DragEvent, item: MediaItem) {
  e.dataTransfer?.setData("application/json", JSON.stringify({ mediaId: item.id }));
  e.dataTransfer!.effectAllowed = "copy";
  uiStore.getState().startDrag("media", { mediaId: item.id, mediaType: item.type });
}

async function addMediaToTimeline(item: MediaItem) {
  const { addClipToNewTrack } = projectStore.getState();
  await addClipToNewTrack(item.id);
}

async function handleConfirmAspectRatioMatch() {
  if (!aspectRatioDialogData.value) return;
  await projectStore.getState().updateSettings({
    width: aspectRatioDialogData.value.videoWidth,
    height: aspectRatioDialogData.value.videoHeight,
  });
  const itemToAdd = aspectRatioDialogData.value.itemToAdd;
  showAspectRatioDialog.value = false;
  aspectRatioDialogData.value = null;
  await addMediaToTimeline(itemToAdd);
}

async function handleCancelAspectRatioMatch() {
  if (!aspectRatioDialogData.value) return;
  const itemToAdd = aspectRatioDialogData.value.itemToAdd;
  showAspectRatioDialog.value = false;
  aspectRatioDialogData.value = null;
  await addMediaToTimeline(itemToAdd);
}

async function handleAddToTimeline(item: MediaItem) {
  const { project: currentProject } = projectStore.getState();
  const tracks = currentProject.timeline.tracks;
  const hasClips = tracks.some((t) => t.clips.length > 0);

  if (!hasClips && item.type === "video" && item.metadata?.width && item.metadata?.height) {
    const videoWidth = item.metadata.width;
    const videoHeight = item.metadata.height;
    const projectWidth = currentProject.settings.width;
    const projectHeight = currentProject.settings.height;

    if (videoWidth !== projectWidth || videoHeight !== projectHeight) {
      aspectRatioDialogData.value = { videoWidth, videoHeight, itemToAdd: item };
      showAspectRatioDialog.value = true;
      return;
    }
  }

  await addMediaToTimeline(item);
}

async function handleImportBackground(preset: BackgroundPreset) {
  generatingBackground.value = preset.id;
  try {
    const { width, height } = projectState.value.project.settings;
    const blob = await generateBackgroundBlob(preset, width, height);
    const file = new File([blob], `${preset.name}_${width}x${height}.png`, { type: "image/png" });
    const result = await projectStore.getState().importMedia(file);
    if (result.success && result.actionId) {
      const { addClipToNewTrack } = projectStore.getState();
      await addClipToNewTrack(result.actionId);
    }
  } catch (error) {
    console.error("Failed to generate background:", error);
  } finally {
    generatingBackground.value = null;
  }
}

async function handleAddShape(shapeType: ShapeType) {
  const state = projectStore.getState();
  const { createShapeClip, addTrack } = state;
  const tracksBefore = state.project.timeline.tracks;
  await addTrack("graphics", 0);
  const tracksAfter = projectStore.getState().project.timeline.tracks;
  const newGraphicsTrack = tracksAfter.find(
    (t: any) => t.type === "graphics" && !tracksBefore.some((bt: any) => bt.id === t.id)
  );
  if (newGraphicsTrack) {
    createShapeClip(newGraphicsTrack.id, 0, shapeType);
  }
}

async function handleOpenKieAI(item: MediaItem) {
  try {
    const blob = await loadMediaBlob(item.id);
    if (!blob) {
      toast.error("素材不存在", "无法加载该素材的图片数据。");
      return;
    }
    const mimeType = blob.type || (item.name.match(/\.png$/i) ? "image/png" : "image/jpeg");
    const file = new File([blob], item.name, { type: mimeType });
    kieaiDialog.value = { file, previewUrl: item.thumbnailUrl };
  } catch (err) {
    console.error("[KieAI] Failed to load media blob:", err);
    toast.error("打开 KieAI 失败", err instanceof Error ? err.message : "未知错误");
  }
}

function handleRetryKieAI(item: MediaItem) {
  if (!item.kieaiTaskId) return;
  projectStore.getState().setKieAIItemState(item.id, true, false);
  kieaiStore.getState().retryTask(item.kieaiTaskId);
}
</script>

<style scoped>
/* Scoped styles */
</style>
