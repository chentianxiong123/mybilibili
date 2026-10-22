<template>
  <div class="article-editor">
    <!-- 顶部工具栏 -->
    <div v-if="editor" class="eva3-toolbar">
      <button class="eva3-toolbar-item" title="撤销" :disabled="!editor.can().undo()" @click="editor.chain().focus().undo().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M9 14 4 9l5-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M4 9h10a6 6 0 0 1 0 12h-3" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
      <button class="eva3-toolbar-item" title="重做" :disabled="!editor.can().redo()" @click="editor.chain().focus().redo().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="m15 14 5-5-5-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M20 9H10a6 6 0 0 0 0 12h3" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
      <span class="eva3-toolbar-splitter"></span>
      <button class="eva3-toolbar-item" title="清除格式" @click="editor.chain().focus().unsetAllMarks().clearNodes().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M4 6h16M9 6l-2 12M14 6l-2 12M5 18h8" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
      </button>
      <span class="eva3-toolbar-splitter"></span>

      <!-- 标题 -->
      <div class="tb-dropdown" @mouseleave="openMenu = null">
        <button class="eva3-toolbar-item" :class="{ active: editor.isActive('heading') }" title="标题" @click="openMenu = openMenu === 'head' ? null : 'head'">
          <span class="tb-glyph">H</span>
          <svg class="tb-caret" viewBox="0 0 12 12" width="10" height="10"><path d="M2 4l4 4 4-4" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg>
        </button>
        <div v-if="openMenu === 'head'" class="tb-menu">
          <button v-for="l in [1,2,3,4,5,6]" :key="l" class="tb-menu-item" :class="{ active: editor.isActive('heading', { level: l }) }" @click="setHeading(l)">标题 {{ l }}</button>
          <button class="tb-menu-item" :class="{ active: editor.isActive('paragraph') }" @click="setParagraph()">正文</button>
        </div>
      </div>

      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('bold') }" title="加粗" @click="editor.chain().focus().toggleBold().run()"><span class="tb-glyph" style="font-weight:700">B</span></button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('italic') }" title="斜体" @click="editor.chain().focus().toggleItalic().run()"><span class="tb-glyph" style="font-style:italic">I</span></button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('underline') }" title="下划线" @click="editor.chain().focus().toggleUnderline().run()"><span class="tb-glyph" style="text-decoration:underline">U</span></button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('strike') }" title="删除线" @click="editor.chain().focus().toggleStrike().run()"><span class="tb-glyph" style="text-decoration:line-through">S</span></button>

      <label class="eva3-toolbar-item tb-color" title="文字颜色">
        <span class="tb-glyph">A</span>
        <span class="tb-color-bar" :style="{ background: textColor }"></span>
        <input type="color" :value="textColor" @input="onTextColor($event)" />
      </label>
      <label class="eva3-toolbar-item tb-color" title="背景颜色">
        <span class="tb-glyph tb-highlight" :style="{ background: highlightColor }">A</span>
        <input type="color" :value="highlightColor" @input="onHighlight($event)" />
      </label>

      <span class="eva3-toolbar-splitter"></span>
      <div class="tb-dropdown" @mouseleave="openMenu = null">
        <button class="eva3-toolbar-item" title="对齐" @click="openMenu = openMenu === 'align' ? null : 'align'">
          <svg viewBox="0 0 24 24" width="18" height="18"><path d="M4 6h16M4 10h10M4 14h16M4 18h10" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>
          <svg class="tb-caret" viewBox="0 0 12 12" width="10" height="10"><path d="M2 4l4 4 4-4" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg>
        </button>
        <div v-if="openMenu === 'align'" class="tb-menu">
          <button class="tb-menu-item" :class="{ active: editor.isActive({ textAlign: 'left' }) }" @click="setAlign('left')">左对齐</button>
          <button class="tb-menu-item" :class="{ active: editor.isActive({ textAlign: 'center' }) }" @click="setAlign('center')">居中</button>
          <button class="tb-menu-item" :class="{ active: editor.isActive({ textAlign: 'right' }) }" @click="setAlign('right')">右对齐</button>
        </div>
      </div>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('orderedList') }" title="有序列表" @click="editor.chain().focus().toggleOrderedList().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M10 7h10M10 12h10M10 17h10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><text x="4" y="9" font-size="6" fill="currentColor">1</text><text x="4" y="14" font-size="6" fill="currentColor">2</text><text x="4" y="19" font-size="6" fill="currentColor">3</text></svg>
      </button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('bulletList') }" title="无序列表" @click="editor.chain().focus().toggleBulletList().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M10 7h10M10 12h10M10 17h10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><circle cx="4.5" cy="7" r="1.6" fill="currentColor"/><circle cx="4.5" cy="12" r="1.6" fill="currentColor"/><circle cx="4.5" cy="17" r="1.6" fill="currentColor"/></svg>
      </button>

      <span class="eva3-toolbar-splitter"></span>
      <button class="eva3-toolbar-item" title="插入图片" @click="pickImage">
        <svg viewBox="0 0 24 24" width="18" height="18"><rect x="3.5" y="4.5" width="17" height="15" rx="2" fill="none" stroke="currentColor" stroke-width="1.6"/><circle cx="8.5" cy="9.5" r="1.6" fill="currentColor"/><path d="m4 17 5-5 4 4 3-3 4 4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>
      </button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('link') }" title="插入链接" @click="setLinkByPrompt">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M9.5 14.5 14.5 9.5" stroke="currentColor" stroke-width="1.7" fill="none" stroke-linecap="round"/><path d="M8 12 6 14a3 3 0 0 0 4 4l2-2M16 12l2-2a3 3 0 0 0-4-4l-2 2" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
      <button class="eva3-toolbar-item" title="分割线" @click="editor.chain().focus().setHorizontalRule().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M4 12h16" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
      </button>
      <button class="eva3-toolbar-item" :class="{ active: editor.isActive('blockquote') }" title="引用" @click="editor.chain().focus().toggleBlockquote().run()">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M6 8v8M10 9h9M10 15h9" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/><rect x="4" y="7" width="2" height="10" rx="1" fill="currentColor"/></svg>
      </button>

      <div class="tb-dropdown" @mouseleave="openMenu = null">
        <button class="eva3-toolbar-item" title="表情" @click="openMenu = openMenu === 'emoji' ? null : 'emoji'">
          <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="8.5" fill="none" stroke="currentColor" stroke-width="1.6"/><circle cx="9" cy="10" r="1.2" fill="currentColor"/><circle cx="15" cy="10" r="1.2" fill="currentColor"/><path d="M8.5 14.5a4 4 0 0 0 7 0" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
        </button>
        <div v-if="openMenu === 'emoji'" class="tb-menu tb-menu--emoji">
          <button v-for="e in emojis" :key="e" class="tb-emoji" @click="insertEmoji(e)">{{ e }}</button>
        </div>
      </div>
      <div class="tb-dropdown" @mouseleave="openMenu = null">
        <button class="eva3-toolbar-item" title="插入" @click="openMenu = openMenu === 'plus' ? null : 'plus'">
          <svg viewBox="0 0 24 24" width="18" height="18"><path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
        </button>
        <div v-if="openMenu === 'plus'" class="tb-menu">
          <button class="tb-menu-item" @click="insertFromPlus('hr')">分割线</button>
          <button class="tb-menu-item" @click="insertFromPlus('quote')">引用</button>
          <button class="tb-menu-item" @click="insertFromPlus('code')">代码块</button>
          <button class="tb-menu-item" @click="insertFromPlus('ol')">有序列表</button>
          <button class="tb-menu-item" @click="insertFromPlus('ul')">无序列表</button>
        </div>
      </div>

      <span class="eva3-toolbar-splitter"></span>
      <button class="eva3-toolbar-item" title="预览" @click="emit('preview')">
        <svg viewBox="0 0 24 24" width="18" height="18"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" fill="none" stroke="currentColor" stroke-width="1.6"/><circle cx="12" cy="12" r="2.6" fill="none" stroke="currentColor" stroke-width="1.6"/></svg>
      </button>
    </div>

    <!-- 编辑区 -->
    <div class="editor-scroll">
      <div class="editor-inner">
        <textarea
          class="article-title"
          :value="title"
          rows="1"
          maxlength="50"
          placeholder="请输入标题（建议30字以内）"
          @input="onTitleInput"
        ></textarea>
        <EditorContent :editor="editor" class="article-content" />
      </div>
    </div>

    <div class="editor-counter">{{ charCount }}/100000</div>

    <input ref="fileRef" type="file" accept="image/*" style="display:none" @change="onFilePicked" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle } from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
import Highlight from '@tiptap/extension-highlight'
import Placeholder from '@tiptap/extension-placeholder'
import CharacterCount from '@tiptap/extension-character-count'
import { ElMessage } from 'element-plus'
import { articleApi } from '@/api/article'

const props = defineProps<{
  title: string
  content: string
}>()

const emit = defineEmits<{
  (e: 'update:title', v: string): void
  (e: 'update:content', v: string): void
  (e: 'preview'): void
}>()

const openMenu = ref<string | null>(null)
const fileRef = ref<HTMLInputElement | null>(null)
const textColor = ref('#18191c')
const highlightColor = ref('#fff359')
const emojis = ['😀','😄','😂','🤣','😊','😍','🤔','😎','😭','😡','👍','👎','🎉','🔥','💯','❤️','🌹','🍺','🐶','🌟']

const editor = useEditor({
  content: props.content || '',
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3, 4, 5, 6] },
      link: { openOnClick: false, autolink: true, HTMLAttributes: { rel: 'noopener noreferrer nofollow', target: '_blank' } }
    } as any),
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    TextStyle,
    Color,
    Highlight.configure({ multicolor: true }),
    Image.configure({ inline: false }),
    Placeholder.configure({ placeholder: '请输入正文' }),
    CharacterCount.configure({ limit: 100000 })
  ],
  editorProps: {
    attributes: { class: 'eva3-editor' }
  },
  onUpdate: ({ editor }) => {
    emit('update:content', editor.getHTML())
  }
})

const charCount = computed(() => {
  const e: any = editor.value
  return e?.storage?.characterCount?.characters?.() ?? 0
})

watch(
  () => props.content,
  v => {
    const e: any = editor.value
    if (e && v !== e.getHTML()) {
      e.commands.setContent(v || '', false)
    }
  }
)

function onTitleInput(e: Event) {
  emit('update:title', (e.target as HTMLTextAreaElement).value)
}
function setHeading(level: number) {
  editor.value?.chain().focus().toggleHeading({ level: level as any }).run()
  openMenu.value = null
}
function setParagraph() {
  editor.value?.chain().focus().setParagraph().run()
  openMenu.value = null
}
function setAlign(a: string) {
  editor.value?.chain().focus().setTextAlign(a).run()
  openMenu.value = null
}
function onTextColor(e: Event) {
  const c = (e.target as HTMLInputElement).value
  textColor.value = c
  editor.value?.chain().focus().setColor(c).run()
}
function onHighlight(e: Event) {
  const c = (e.target as HTMLInputElement).value
  highlightColor.value = c
  editor.value?.chain().focus().setHighlight({ color: c }).run()
}
function insertEmoji(e: string) {
  editor.value?.chain().focus().insertContent(e).run()
  openMenu.value = null
}
function insertFromPlus(kind: string) {
  const ch = editor.value?.chain().focus()
  if (!ch) return
  if (kind === 'hr') ch.setHorizontalRule().run()
  else if (kind === 'quote') ch.toggleBlockquote().run()
  else if (kind === 'code') ch.toggleCodeBlock().run()
  else if (kind === 'ol') ch.toggleOrderedList().run()
  else if (kind === 'ul') ch.toggleBulletList().run()
  openMenu.value = null
}
function setLinkByPrompt() {
  const prev = editor.value?.getAttributes('link').href || ''
  const url = window.prompt('请输入链接地址', prev)
  if (url === null) return
  if (url === '') {
    editor.value?.chain().focus().unsetLink().run()
    return
  }
  editor.value?.chain().focus().setLink({ href: url }).run()
}
function pickImage() {
  fileRef.value?.click()
}
async function onFilePicked(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    const res: any = await articleApi.uploadImage(file)
    const url = res?.data?.url || res?.url
    if (url) {
      editor.value?.chain().focus().setImage({ src: url }).run()
    } else {
      ElMessage.error('图片上传失败')
    }
  } catch {
    ElMessage.error('图片上传失败')
  }
}

defineExpose({
  getHTML: () => editor.value?.getHTML() || '',
  getText: () => editor.value?.getText() || '',
  isEmpty: () => editor.value?.isEmpty ?? true
})

onBeforeUnmount(() => editor.value?.destroy())
</script>

<style scoped>
.article-editor {
  background: #fff;
  border-radius: 2px;
  display: flex;
  flex-direction: column;
}

/* ===== 工具栏 ===== */
.eva3-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
  padding: 8px 16px;
  border-bottom: 1px solid #f1f2f3;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 5;
}
.eva3-toolbar-item {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 1px;
  min-width: 28px;
  height: 28px;
  padding: 0 4px;
  border: none;
  background: transparent;
  border-radius: 4px;
  color: #61666d;
  cursor: pointer;
  font-size: 15px;
}
.eva3-toolbar-item:hover { background: #f1f2f3; color: #18191c; }
.eva3-toolbar-item.active { color: #00aeec; background: rgba(0,174,236,.1); }
.eva3-toolbar-item:disabled { opacity: .35; cursor: not-allowed; }
.eva3-toolbar-splitter { width: 1px; height: 18px; background: #e3e5e7; margin: 0 6px; }
.tb-glyph { font-size: 15px; line-height: 1; }
.tb-caret { opacity: .6; }
.tb-color { overflow: hidden; }
.tb-color input[type="color"] { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
.tb-color-bar { position: absolute; left: 5px; right: 5px; bottom: 3px; height: 3px; border-radius: 2px; }
.tb-highlight { border-radius: 2px; padding: 0 2px; }

.tb-dropdown { position: relative; }
.tb-menu {
  position: absolute;
  top: 32px;
  left: 0;
  min-width: 116px;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0,0,0,.12);
  padding: 4px;
  z-index: 20;
}
.tb-menu--emoji { display: grid; grid-template-columns: repeat(5, 1fr); width: 200px; }
.tb-emoji { border: none; background: transparent; font-size: 18px; padding: 4px; cursor: pointer; border-radius: 4px; }
.tb-emoji:hover { background: #f1f2f3; }
.tb-menu-item {
  display: block;
  width: 100%;
  text-align: left;
  padding: 6px 10px;
  border: none;
  background: transparent;
  font-size: 13px;
  color: #18191c;
  border-radius: 4px;
  cursor: pointer;
}
.tb-menu-item:hover { background: #f6f7f8; }
.tb-menu-item.active { color: #00aeec; font-weight: 600; }

/* ===== 编辑区 ===== */
.editor-scroll {
  flex: 1;
  overflow-y: auto;
  display: flex;
  justify-content: center;
}
.editor-inner {
  width: 100%;
  max-width: 672px;
  padding: 20px 40px 80px;
  box-sizing: border-box;
}
.article-title {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.5;
  color: #18191c;
  font-family: inherit;
  padding: 0;
  margin-bottom: 12px;
  overflow: hidden;
}
.article-title::placeholder { color: #c9ccd0; }
.article-content :deep(.eva3-editor) {
  outline: none;
  min-height: 420px;
  font-size: 17px;
  line-height: 1.8;
  color: #2f3238;
  word-break: break-word;
}
.article-content :deep(.eva3-editor p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: #c9ccd0;
  float: left;
  height: 0;
  pointer-events: none;
}
.article-content :deep(.eva3-editor h1) { font-size: 28px; font-weight: 700; margin: 16px 0 8px; }
.article-content :deep(.eva3-editor h2) { font-size: 24px; font-weight: 700; margin: 14px 0 8px; }
.article-content :deep(.eva3-editor h3) { font-size: 20px; font-weight: 700; margin: 12px 0 6px; }
.article-content :deep(.eva3-editor h4) { font-size: 18px; font-weight: 700; }
.article-content :deep(.eva3-editor p) { margin: 8px 0; }
.article-content :deep(.eva3-editor ul),
.article-content :deep(.eva3-editor ol) { padding-left: 26px; margin: 8px 0; }
.article-content :deep(.eva3-editor blockquote) {
  border-left: 3px solid #e3e5e7; padding: 4px 0 4px 14px; margin: 10px 0; color: #61666d; background: #fafafa;
}
.article-content :deep(.eva3-editor img) { max-width: 100%; border-radius: 4px; margin: 8px 0; }
.article-content :deep(.eva3-editor a) { color: #00aeec; text-decoration: none; }
.article-content :deep(.eva3-editor hr) { border: none; border-top: 1px solid #e3e5e7; margin: 16px 0; }
.article-content :deep(.eva3-editor pre) { background: #f6f7f8; padding: 12px; border-radius: 6px; overflow:auto; }
.article-content :deep(.eva3-editor code) { background: #f6f7f8; padding: 2px 4px; border-radius: 3px; font-size: 14px; }

.editor-counter {
  text-align: right;
  padding: 6px 24px;
  font-size: 12px;
  color: #c9ccd0;
  border-top: 1px solid #f1f2f3;
}
</style>
