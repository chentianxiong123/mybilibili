import React, { useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import * as remote from '@electron/remote';
import { Services } from 'components-react/service-provider';
import { useVuex } from 'components-react/hooks';
import Display from 'components-react/shared/Display';
import {
  ERecordingState,
  EReplayBufferState,
  EStreamingState,
} from 'services/streaming';
import { TSourceType } from 'services/sources';
import styles from './MybilibiliLive.m.less';

type TQuickSource = {
  label: string;
  type: TSourceType;
  settings: Record<string, unknown>;
};

const QUICK_SOURCES: TQuickSource[] = [
  {
    label: '显示器采集',
    type: 'monitor_capture',
    settings: {},
  },
  {
    label: '窗口采集',
    type: 'window_capture',
    settings: {},
  },
  {
    label: '游戏采集',
    type: 'game_capture',
    settings: {},
  },
  {
    label: '视频采集设备',
    type: 'dshow_input',
    settings: {},
  },
  {
    label: '音频输入',
    type: 'wasapi_input_capture',
    settings: {},
  },
  {
    label: '桌面音频',
    type: 'wasapi_output_capture',
    settings: {},
  },
  {
    label: '图片',
    type: 'image_source',
    settings: {},
  },
  {
    label: '媒体',
    type: 'ffmpeg_source',
    settings: {},
  },
  {
    label: '色块',
    type: 'color_source',
    settings: { color: 0xff202124, width: 1280, height: 720 },
  },
  {
    label: '文字',
    type: 'text_gdiplus',
    settings: {
      text: 'mybilibili Live',
      font: { face: 'Arial', size: 48, style: '' },
      color: 0xffffffff,
    },
  },
  {
    label: '浏览器',
    type: 'browser_source',
    settings: { url: 'about:blank', width: 1280, height: 720, shutdown: true },
  },
];

const sourceTypeNames: Partial<Record<TSourceType, string>> = {
  color_source: '色块',
  text_gdiplus: '文字',
  browser_source: '浏览器',
  image_source: '图片',
  ffmpeg_source: '媒体',
  monitor_capture: '显示器采集',
  window_capture: '窗口采集',
  game_capture: '游戏采集',
  dshow_input: '视频采集设备',
  wasapi_input_capture: '音频输入',
  wasapi_output_capture: '桌面音频',
  scene: '场景',
};

const maskKey = (key?: string) => {
  if (!key) return '未加载';
  if (key.length <= 8) return '********';
  return `${key.slice(0, 4)}****${key.slice(-4)}`;
};

const streamingLabel = (status?: string) => {
  switch (status) {
    case EStreamingState.Live:
      return '直播中';
    case EStreamingState.Starting:
      return '启动中';
    case EStreamingState.Ending:
      return '结束中';
    case EStreamingState.Reconnecting:
      return '重连中';
    case EStreamingState.Offline:
    default:
      return '未开播';
  }
};

const recordingLabel = (status?: string) => {
  switch (status) {
    case ERecordingState.Recording:
      return '录制中';
    case ERecordingState.Starting:
    case ERecordingState.Start:
      return '启动中';
    case ERecordingState.Stopping:
    case ERecordingState.Writing:
      return '停止中';
    case ERecordingState.Wrote:
      return '已写入';
    case ERecordingState.Offline:
    default:
      return '未录制';
  }
};

const replayLabel = (status?: string) => {
  switch (status) {
    case EReplayBufferState.Running:
      return '运行中';
    case EReplayBufferState.Saving:
      return '保存中';
    case EReplayBufferState.Stopping:
      return '停止中';
    case EReplayBufferState.Wrote:
      return '已写入';
    case EReplayBufferState.Offline:
    default:
      return '未启用';
  }
};

const roomStatusLabel = (status?: string) => {
  switch (status) {
    case 'live':
      return '直播中';
    case 'starting':
      return '启动中';
    case 'ending':
      return '结束中';
    case 'reconnecting':
      return '重连中';
    case 'offline':
      return '未开播';
    default:
      return status || '未加载';
  }
};

function Field(p: { label: string; children: React.ReactNode; className?: string }) {
  return (
    <label className={cx(styles.field, p.className)}>
      <span>{p.label}</span>
      {p.children}
    </label>
  );
}

function Panel(p: {
  id?: string;
  title: string;
  meta?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section id={p.id} className={cx(styles.panel, p.className)}>
      <div className={styles.panelHeader}>
        <div>
          <h2>{p.title}</h2>
          {p.meta && <span>{p.meta}</span>}
        </div>
        {p.action}
      </div>
      {p.children}
    </section>
  );
}

function WindowChrome() {
  const [isMaximized, setIsMaximized] = useState(false);

  useEffect(() => {
    const win = remote.getCurrentWindow();
    const syncMaximized = () => setIsMaximized(win.isMaximized());
    syncMaximized();
    win.on('maximize', syncMaximized);
    win.on('unmaximize', syncMaximized);
    win.on('restore', syncMaximized);

    return () => {
      win.removeListener('maximize', syncMaximized);
      win.removeListener('unmaximize', syncMaximized);
      win.removeListener('restore', syncMaximized);
    };
  }, []);

  function minimize() {
    remote.getCurrentWindow().minimize();
  }

  function toggleMaximize() {
    const win = remote.getCurrentWindow();
    if (win.isMaximized()) {
      win.unmaximize();
    } else {
      win.maximize();
    }
    setIsMaximized(win.isMaximized());
  }

  function close() {
    remote.getCurrentWindow().close();
  }

  return (
    <div className={styles.windowChrome} onDoubleClick={toggleMaximize}>
      <div className={styles.windowTitle}>
        <strong>mybilibili Live Desktop</strong>
        <span>OBS 推流器</span>
      </div>
      <div className={styles.windowControls} onDoubleClick={event => event.stopPropagation()}>
        <button aria-label="最小化" title="最小化" onClick={minimize}>
          <span />
        </button>
        <button
          aria-label={isMaximized ? '还原' : '最大化'}
          title={isMaximized ? '还原' : '最大化'}
          onClick={toggleMaximize}
        >
          <i className={isMaximized ? styles.restoreIcon : styles.maximizeIcon} />
        </button>
        <button aria-label="关闭" title="关闭" className={styles.closeButton} onClick={close}>
          <b />
        </button>
      </div>
    </div>
  );
}

export default function MybilibiliLivePage(p: { className?: string }) {
  const {
    AudioService,
    MybilibiliLiveService,
    ScenesService,
    SettingsService,
    SourcesService,
    StreamingService,
  } = Services;

  const state = useVuex(() => {
    const scenesState = ScenesService.state;
    const sourcesState = SourcesService.state;
    const audioState = AudioService.state;
    const activeScene = scenesState.scenes[scenesState.activeSceneId];

    return {
      baseUrl: MybilibiliLiveService.state.baseUrl,
      token: MybilibiliLiveService.state.token,
      user: MybilibiliLiveService.state.user,
      room: MybilibiliLiveService.state.room,
      rtmpUrl: MybilibiliLiveService.state.rtmpUrl,
      lastError: MybilibiliLiveService.state.lastError,
      streamingStatus: StreamingService.state.streamingStatus,
      recordingStatus: StreamingService.state.recordingStatus,
      replayBufferStatus: StreamingService.state.replayBufferStatus,
      scenes: scenesState.displayOrder
        .map(sceneId => scenesState.scenes[sceneId])
        .filter(Boolean),
      activeSceneId: scenesState.activeSceneId,
      activeScene,
      activeNodes: activeScene?.nodes || [],
      sourcesById: sourcesState.sources,
      sources: Object.values(sourcesState.sources),
      audioSources: Object.values(audioState.audioSources),
    };
  });

  const [draftBaseUrl, setDraftBaseUrl] = useState(state.baseUrl);
  const [draftUsername, setDraftUsername] = useState('');
  const [draftPassword, setDraftPassword] = useState('');
  const [draftRoomName, setDraftRoomName] = useState(state.room?.roomName || '');
  const [draftCategory, setDraftCategory] = useState(state.room?.category || '');
  const [draftCoverUrl, setDraftCoverUrl] = useState(state.room?.coverUrl || '');
  const [draftSceneName, setDraftSceneName] = useState('');
  const [draftSourceName, setDraftSourceName] = useState('');
  const [selectedQuickSource, setSelectedQuickSource] = useState<TSourceType>('monitor_capture');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');

  const isStreaming = state.streamingStatus !== EStreamingState.Offline;
  const isRecording = state.recordingStatus !== ERecordingState.Offline;
  const isReplayBufferRunning = state.replayBufferStatus !== EReplayBufferState.Offline;
  const isLoggedIn = Boolean(state.token);
  const accountName =
    state.user?.nickname || state.user?.username || (isLoggedIn ? '已登录账号' : '未登录');
  const roomName = state.room?.roomName || '我的直播间';
  const streamKey = state.room?.streamKey || '';
  const activeSceneName = state.activeScene?.name || '未选择场景';
  const quickSource = useMemo(
    () => QUICK_SOURCES.find(item => item.type === selectedQuickSource) || QUICK_SOURCES[0],
    [selectedQuickSource],
  );

  useEffect(() => {
    setDraftBaseUrl(state.baseUrl);
  }, [state.baseUrl]);

  useEffect(() => {
    setDraftRoomName(state.room?.roomName || '');
    setDraftCategory(state.room?.category || '');
    setDraftCoverUrl(state.room?.coverUrl || '');
  }, [state.room?.id, state.room?.roomName, state.room?.category, state.room?.coverUrl]);

  async function run(label: string, action: () => Promise<unknown> | unknown) {
    setBusy(true);
    setMessage('');
    try {
      await action();
      setMessage(label);
    } catch (error: unknown) {
      setMessage(error instanceof Error ? error.message : '操作失败');
    } finally {
      setBusy(false);
    }
  }

  function saveConnection() {
    MybilibiliLiveService.actions.setConnection(draftBaseUrl);
    setMessage('网关地址已保存');
  }

  async function loginToSite() {
    await MybilibiliLiveService.actions.return.login(
      draftBaseUrl,
      draftUsername,
      draftPassword,
    );
    setDraftPassword('');
    await MybilibiliLiveService.actions.return.loadOrCreateRoom();
  }

  async function saveRoom() {
    await MybilibiliLiveService.actions.return.updateRoom({
      roomName: draftRoomName,
      category: draftCategory,
      coverUrl: draftCoverUrl,
    });
  }

  async function startStreaming() {
    await MybilibiliLiveService.actions.return.applyStreamSettings();
    await StreamingService.actions.return.startStreaming();
    if (state.room) {
      await MybilibiliLiveService.actions.return.updateRoomStatus('live');
    }
  }

  async function stopStreaming() {
    await StreamingService.actions.return.stopStreaming();
    if (state.room) {
      await MybilibiliLiveService.actions.return.updateRoomStatus('offline');
    }
  }

  async function toggleRecording() {
    if (isRecording) {
      await StreamingService.actions.return.stopRecording();
    } else {
      await StreamingService.actions.return.startRecording();
    }
  }

  async function toggleReplayBuffer() {
    if (isReplayBufferRunning) {
      await StreamingService.actions.return.stopReplayBuffer();
    } else {
      await StreamingService.actions.return.startReplayBuffer();
    }
  }

  async function createScene() {
    const name = draftSceneName.trim() || `场景 ${state.scenes.length + 1}`;
    await ScenesService.actions.return.createScene(name, { makeActive: true } as any);
    setDraftSceneName('');
  }

  function renameScene(sceneId: string, currentName: string) {
    const name = window.prompt('重命名场景', currentName)?.trim();
    if (!name || name === currentName) return;
    const scene = (ScenesService.views as any).getScene(sceneId);
    if (scene?.setName) scene.setName(name);
  }

  function removeScene(sceneId: string) {
    const scene = (ScenesService.views as any).getScene(sceneId);
    if (!scene) return;
    if (state.scenes.length < 2) {
      setMessage('至少需要保留一个场景');
      return;
    }
    scene.remove();
    setMessage('场景已删除');
  }

  function moveScene(sceneId: string, direction: -1 | 1) {
    const ids = state.scenes.map(scene => scene.id);
    const index = ids.indexOf(sceneId);
    const nextIndex = index + direction;
    if (index < 0 || nextIndex < 0 || nextIndex >= ids.length) return;
    const next = ids.slice();
    [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
    ScenesService.actions.setSceneOrder(next);
  }

  async function addQuickSource() {
    let sceneId = state.activeSceneId;

    if (!sceneId) {
      const scene = (await ScenesService.actions.return.createScene('场景 1', {
        makeActive: true,
      } as any)) as any;
      sceneId = scene?.id;
    }

    if (!sceneId) throw new Error('无法创建场景');

    const sourceName = draftSourceName.trim() || quickSource.label;
    const sceneItemId = await ScenesService.actions.return.createAndAddSource(
      sceneId,
      sourceName,
      quickSource.type,
      quickSource.settings,
    );

    const sceneItem = (ScenesService.views as any).getSceneItem(sceneItemId);
    if (sceneItem?.fitToScreen) sceneItem.fitToScreen();
    setDraftSourceName('');
  }

  async function toggleNodeVisibility(nodeId: string) {
    await (ScenesService.views as any).toggleNodeVisibility(nodeId);
  }

  function removeNode(nodeId: string) {
    const node = (ScenesService.views as any).getSceneNode(nodeId);
    if (node) node.remove();
  }

  function renameNode(nodeId: string) {
    const node = (ScenesService.views as any).getSceneNode(nodeId);
    if (!node) return;
    const source = node.sourceId ? (SourcesService.views as any).getSource(node.sourceId) : null;
    const currentName = source?.name || node.name || '未命名';
    const name = window.prompt('重命名来源', currentName)?.trim();
    if (!name || name === currentName) return;
    if (source?.setName) {
      source.setName(name);
    } else if (node.setName) {
      node.setName(name);
    }
  }

  function moveNode(nodeId: string, direction: -1 | 1) {
    const scene = (ScenesService.views as any).activeScene;
    if (!scene) return;
    const ids = state.activeNodes.map(node => node.id);
    const index = ids.indexOf(nodeId);
    const nextIndex = index + direction;
    if (index < 0 || nextIndex < 0 || nextIndex >= ids.length) return;
    if (direction < 0) {
      scene.placeBefore(nodeId, ids[nextIndex]);
    } else {
      scene.placeAfter(nodeId, ids[nextIndex]);
    }
  }

  async function copyText(text: string, label: string) {
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
      setMessage(`${label}已复制`);
    } catch (error: unknown) {
      setMessage(error instanceof Error ? error.message : '复制失败');
    }
  }

  return (
    <div className={cx(styles.page, p.className)}>
      <WindowChrome />
      <div className={styles.menuBar}>
        <button onClick={() => SettingsService.actions.showSettings('General' as any)}>文件</button>
        <button onClick={() => SettingsService.actions.showSettings('Hotkeys' as any)}>编辑</button>
        <button onClick={() => SettingsService.actions.showSettings('Video' as any)}>视图</button>
        <button>配置文件</button>
        <button>场景集合</button>
        <button onClick={() => SettingsService.actions.showSettings('Output' as any)}>工具</button>
        <button onClick={() => setMessage('mybilibili Live Desktop OBS 推流器')}>帮助</button>
      </div>

      <main className={styles.obsShell}>
        <section className={styles.previewArea}>
          <div className={styles.previewToolbar}>
            <span>{activeSceneName}</span>
            <span>
              直播：{streamingLabel(state.streamingStatus)} / 录制：
              {recordingLabel(state.recordingStatus)}
            </span>
          </div>
          <div className={styles.previewCanvas}>
            <Display drawUI paddingSize={8} />
            {!state.activeSceneId && (
              <div className={styles.emptyPreview}>
                <strong>未加载场景</strong>
                <span>先创建场景，再添加来源</span>
              </div>
            )}
          </div>
          {(message || state.lastError) && (
            <div className={cx(styles.notice, state.lastError && styles.noticeError)}>
              {state.lastError || message}
            </div>
          )}
        </section>

        <section className={styles.dockBar}>
          <Panel
            id="scenes"
            title="场景"
            meta={`${state.scenes.length} 个`}
            action={
              <button
                className={styles.primaryButton}
                onClick={() => run('场景已创建', createScene)}
                disabled={busy}
              >
                新建
              </button>
            }
          >
            <div className={styles.inlineCreate}>
              <input
                value={draftSceneName}
                onChange={event => setDraftSceneName(event.target.value)}
                placeholder="新场景名称"
              />
            </div>
            <div className={styles.sceneList}>
              {state.scenes.map(scene => (
                <div
                  key={scene.id}
                  className={cx(
                    styles.sceneItem,
                    scene.id === state.activeSceneId && styles.sceneItemActive,
                  )}
                >
                  <button
                    className={styles.sceneSelectButton}
                    onClick={() => ScenesService.actions.makeSceneActive(scene.id)}
                  >
                    <span>{scene.name}</span>
                    <small>{scene.nodes.length} 个来源</small>
                  </button>
                  <div className={styles.rowActions}>
                    <button onClick={() => moveScene(scene.id, -1)} title="上移">
                      ↑
                    </button>
                    <button onClick={() => moveScene(scene.id, 1)} title="下移">
                      ↓
                    </button>
                    <button onClick={() => renameScene(scene.id, scene.name)}>重命名</button>
                    <button onClick={() => removeScene(scene.id)}>删除</button>
                  </div>
                </div>
              ))}
              {!state.scenes.length && <div className={styles.emptyState}>暂无场景</div>}
            </div>
          </Panel>

          <Panel
            id="sources"
            title="来源"
            meta={activeSceneName}
            action={
              <button
                className={styles.primaryButton}
                onClick={() => run('来源已添加', addQuickSource)}
                disabled={busy}
              >
                添加
              </button>
            }
          >
            <div className={styles.sourceCreate}>
              <select
                value={selectedQuickSource}
                onChange={event => setSelectedQuickSource(event.target.value as TSourceType)}
              >
                {QUICK_SOURCES.map(option => (
                  <option key={option.type} value={option.type}>
                    {option.label}
                  </option>
                ))}
              </select>
              <input
                value={draftSourceName}
                onChange={event => setDraftSourceName(event.target.value)}
                placeholder="来源名称"
              />
            </div>
            <div className={styles.sourceList}>
              {state.activeNodes.map(node => {
                const source = node.sceneNodeType === 'item'
                  ? state.sourcesById[(node as any).sourceId]
                  : null;
                const sourceName = source?.name || (node as any).name || '未命名';
                const sourceType = source?.type ? sourceTypeNames[source.type] || source.type : '文件夹';

                return (
                  <div key={node.id} className={styles.sourceRow}>
                    <button
                      className={cx(styles.visibilityButton, !(node as any).visible && styles.off)}
                      onClick={() => run('可见性已更新', () => toggleNodeVisibility(node.id))}
                    >
                      {(node as any).visible === false ? '隐藏' : '可见'}
                    </button>
                    <div className={styles.sourceMeta}>
                      <strong>{sourceName}</strong>
                      <span>{sourceType}</span>
                    </div>
                    {source && (
                      <button
                        className={styles.iconButton}
                        onClick={() => SourcesService.actions.showSourceProperties(source.sourceId)}
                      >
                        设置
                      </button>
                    )}
                    <button className={styles.iconButton} onClick={() => moveNode(node.id, -1)}>
                      ↑
                    </button>
                    <button className={styles.iconButton} onClick={() => moveNode(node.id, 1)}>
                      ↓
                    </button>
                    <button className={styles.iconButton} onClick={() => renameNode(node.id)}>
                      改名
                    </button>
                    <button
                      className={styles.iconButton}
                      onClick={() => run('来源已删除', () => removeNode(node.id))}
                    >
                      删除
                    </button>
                  </div>
                );
              })}
              {!state.activeNodes.length && <div className={styles.emptyState}>当前场景暂无来源</div>}
            </div>
          </Panel>

          <Panel
            id="audio"
            title="混音器"
            meta={`${state.audioSources.length} 路音频`}
            action={
              <button
                className={styles.secondaryButton}
                onClick={() => AudioService.actions.showAdvancedSettings()}
              >
                高级音频
              </button>
            }
          >
            <div className={styles.mixerList}>
              {state.audioSources.map(audio => {
                const source = state.sourcesById[audio.sourceId];
                const level = Math.round((audio.fader?.deflection ?? 0) * 100);

                return (
                  <div key={audio.sourceId} className={styles.mixerRow}>
                    <div className={styles.mixerName}>
                      <strong>{source?.name || audio.sourceId}</strong>
                      <span>{audio.muted ? '已静音' : `${level}%`}</span>
                    </div>
                    <input
                      type="range"
                      min={0}
                      max={1}
                      step={0.01}
                      value={audio.fader?.deflection ?? 0}
                      onChange={event =>
                        AudioService.actions.setFader(audio.sourceId, {
                          deflection: Number(event.target.value),
                        })
                      }
                    />
                    <button
                      className={cx(styles.iconButton, audio.muted && styles.muted)}
                      onClick={() => SourcesService.actions.setMuted(audio.sourceId, !audio.muted)}
                    >
                      {audio.muted ? '取消静音' : '静音'}
                    </button>
                  </div>
                );
              })}
              {!state.audioSources.length && <div className={styles.emptyState}>暂无音频来源</div>}
            </div>
          </Panel>

          <Panel title="场景转场" meta="默认">
            <div className={styles.transitionDock}>
              <Field label="转场">
                <select defaultValue="cut">
                  <option value="cut">切换</option>
                  <option value="fade">淡入淡出</option>
                </select>
              </Field>
              <Field label="持续时间">
                <input defaultValue="300 ms" />
              </Field>
            </div>
          </Panel>

          <Panel title="控制" meta={roomStatusLabel(state.room?.status)} className={styles.controlsDock}>
            <div className={styles.controlButtons}>
              {!isStreaming ? (
                <button
                  className={styles.dangerButton}
                  onClick={() => run('开播请求已发送', startStreaming)}
                  disabled={busy || !state.room || !streamKey}
                >
                  开始直播
                </button>
              ) : (
                <button
                  className={styles.warningButton}
                  onClick={() => run('已请求结束直播', stopStreaming)}
                  disabled={busy}
                >
                  停止直播
                </button>
              )}
              <button
                className={styles.secondaryButton}
                onClick={() => run(isRecording ? '录制已停止' : '录制已启动', toggleRecording)}
                disabled={busy}
              >
                {isRecording ? '停止录制' : '开始录制'}
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run(
                    isReplayBufferRunning ? '回放缓存已停止' : '回放缓存已启动',
                    toggleReplayBuffer,
                  )
                }
                disabled={busy}
              >
                {isReplayBufferRunning ? '停止回放缓存' : '启动回放缓存'}
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run('推流配置已写入', () =>
                    MybilibiliLiveService.actions.return.applyStreamSettings(),
                  )
                }
                disabled={busy || !state.room}
              >
                写入推流配置
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() => SettingsService.actions.showSettings('Output' as any)}
              >
                设置
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() => SettingsService.actions.showSettings('Video' as any)}
              >
                视频设置
              </button>
            </div>

            <details className={styles.streamDetails}>
              <summary>mybilibili 推流配置</summary>
              <div className={styles.formGrid}>
                <Field label="网关地址" className={styles.fullField}>
                  <input
                    value={draftBaseUrl}
                    onChange={event => setDraftBaseUrl(event.target.value)}
                    placeholder="http://localhost:8080/api"
                  />
                </Field>
                <Field label="账号">
                  <input
                    value={draftUsername}
                    onChange={event => setDraftUsername(event.target.value)}
                    placeholder="用户名或邮箱"
                  />
                </Field>
                <Field label="密码">
                  <input
                    value={draftPassword}
                    onChange={event => setDraftPassword(event.target.value)}
                    type="password"
                    placeholder="站点登录密码"
                  />
                </Field>
              </div>
              <div className={styles.actionRow}>
                <button
                  className={styles.primaryButton}
                  onClick={() => run('登录成功，直播间已加载', loginToSite)}
                  disabled={busy || !draftUsername.trim() || !draftPassword}
                >
                  登录并加载
                </button>
                <button className={styles.secondaryButton} onClick={saveConnection} disabled={busy}>
                  保存网关
                </button>
                <button
                  className={styles.secondaryButton}
                  onClick={() =>
                    run('直播间已刷新', () => MybilibiliLiveService.actions.return.refreshRoom())
                  }
                  disabled={busy || !isLoggedIn}
                >
                  刷新房间
                </button>
              </div>
              <div className={styles.streamRows}>
                <div className={styles.streamRow}>
                  <span>账号</span>
                  <code>{accountName}</code>
                  <button
                    onClick={() => {
                      MybilibiliLiveService.actions.logout();
                      setMessage('已退出 mybilibili 站点登录');
                    }}
                    disabled={!isLoggedIn}
                  >
                    退出
                  </button>
                </div>
                <div className={styles.streamRow}>
                  <span>房间</span>
                  <code>{roomName}</code>
                  <button
                    onClick={() => run('直播间信息已保存', saveRoom)}
                    disabled={busy || !state.room}
                  >
                    保存
                  </button>
                </div>
                <div className={styles.streamRow}>
                  <span>地址</span>
                  <code>{state.rtmpUrl || '未加载'}</code>
                  <button
                    onClick={() => copyText(state.rtmpUrl, '推流地址')}
                    disabled={!state.rtmpUrl}
                  >
                    复制
                  </button>
                </div>
                <div className={styles.streamRow}>
                  <span>密钥</span>
                  <code>{maskKey(streamKey)}</code>
                  <button onClick={() => copyText(streamKey, '流密钥')} disabled={!streamKey}>
                    复制
                  </button>
                </div>
              </div>
              <div className={styles.formGrid}>
                <Field label="房间名称">
                  <input
                    value={draftRoomName}
                    onChange={event => setDraftRoomName(event.target.value)}
                    disabled={!state.room}
                  />
                </Field>
                <Field label="分类">
                  <input
                    value={draftCategory}
                    onChange={event => setDraftCategory(event.target.value)}
                    disabled={!state.room}
                  />
                </Field>
                <Field label="封面地址" className={styles.fullField}>
                  <input
                    value={draftCoverUrl}
                    onChange={event => setDraftCoverUrl(event.target.value)}
                    disabled={!state.room}
                  />
                </Field>
              </div>
            </details>
          </Panel>
        </section>
      </main>
    </div>
  );
}
