import React, { useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import { Services } from 'components-react/service-provider';
import { useVuex } from 'components-react/hooks';
import { EStreamingState } from 'services/streaming';
import styles from './MybilibiliLive.m.less';

const maskKey = (key?: string) => {
  if (!key) return '未加载';
  if (key.length <= 8) return '********';
  return `${key.slice(0, 4)}****${key.slice(-4)}`;
};

const statusLabel = (status?: string) => {
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
    default:
      return '未开播';
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

export default function MybilibiliLivePage(p: { className?: string }) {
  const { MybilibiliLiveService, StreamingService } = Services;
  const {
    baseUrl,
    token,
    user,
    room,
    rtmpUrl,
    lastError,
    obsStreamingStatus,
  } = useVuex(() => ({
    baseUrl: MybilibiliLiveService.state.baseUrl,
    token: MybilibiliLiveService.state.token,
    user: MybilibiliLiveService.state.user,
    room: MybilibiliLiveService.state.room,
    rtmpUrl: MybilibiliLiveService.state.rtmpUrl,
    lastError: MybilibiliLiveService.state.lastError,
    obsStreamingStatus: StreamingService.state.streamingStatus,
  }));

  const [draftBaseUrl, setDraftBaseUrl] = useState(baseUrl);
  const [draftUsername, setDraftUsername] = useState('');
  const [draftPassword, setDraftPassword] = useState('');
  const [draftRoomName, setDraftRoomName] = useState(room?.roomName || '');
  const [draftCategory, setDraftCategory] = useState(room?.category || '');
  const [draftCoverUrl, setDraftCoverUrl] = useState(room?.coverUrl || '');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');

  const roomName = room?.roomName || '我的直播间';
  const roomCategory = room?.category || '未设置';
  const roomStatus = room?.status || '未加载';
  const isObsStreaming = useMemo(
    () => obsStreamingStatus !== EStreamingState.Offline,
    [obsStreamingStatus],
  );
  const streamKey = room?.streamKey || '';
  const isLoggedIn = Boolean(token);
  const accountName = user?.nickname || user?.username || (isLoggedIn ? '已登录账号' : '未登录');

  useEffect(() => {
    setDraftRoomName(room?.roomName || '');
    setDraftCategory(room?.category || '');
    setDraftCoverUrl(room?.coverUrl || '');
  }, [room?.id, room?.roomName, room?.category, room?.coverUrl]);

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
    setMessage('Gateway 地址已保存');
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

  function logoutFromSite() {
    MybilibiliLiveService.actions.logout();
    setMessage('已退出 mybilibili 站点登录');
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
  }

  async function stopStreaming() {
    await StreamingService.actions.return.stopStreaming();
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
      <section className={styles.pageHeader}>
        <div className={styles.headerMain}>
          <div className={styles.titleRow}>
            <h1>直播控制台</h1>
            <span className={cx(styles.liveChip, isObsStreaming && styles.liveChipOn)}>
              OBS：{statusLabel(obsStreamingStatus)}
            </span>
          </div>
          <div className={styles.roomLine}>
            <span className={styles.roomPill}>{roomName}</span>
            <span>账号：{accountName}</span>
            <span>{roomCategory}</span>
            <span>平台状态：{statusLabel(roomStatus)}</span>
          </div>
        </div>
        <div className={styles.headerActions}>
          {!isObsStreaming ? (
            <button
              className={styles.dangerButton}
              onClick={() => run('OBS 开播请求已发送', startStreaming)}
              disabled={busy || !room || !streamKey}
            >
              开始直播
            </button>
          ) : (
            <button
              className={styles.warningButton}
              onClick={() => run('已请求结束直播', stopStreaming)}
              disabled={busy}
            >
              结束直播
            </button>
          )}
        </div>
      </section>

      {(message || lastError) && (
        <div className={cx(styles.notice, lastError && styles.noticeError)}>
          {lastError || message}
        </div>
      )}

      <section className={styles.heroGrid}>
        <div className={styles.coverPanel}>
          <div className={styles.coverFrame}>
            {room?.coverUrl ? (
              <img src={room.coverUrl} alt="直播间封面" />
            ) : (
              <div className={styles.coverPlaceholder}>
                <span>16:9</span>
                <strong>直播间封面</strong>
              </div>
            )}
            <span className={cx(styles.coverBadge, isObsStreaming && styles.coverBadgeLive)}>
              {isObsStreaming ? '直播中' : '未开播'}
            </span>
          </div>
        </div>
        <div className={styles.metricsPanel}>
          <div className={styles.metricItem}>
            <span>房间</span>
            <strong>{room ? '已加载' : '未加载'}</strong>
          </div>
          <div className={styles.metricItem}>
            <span>平台</span>
            <strong>{statusLabel(roomStatus)}</strong>
          </div>
          <div className={styles.metricItem}>
            <span>OBS</span>
            <strong>{statusLabel(obsStreamingStatus)}</strong>
          </div>
          <div className={styles.metricItem}>
            <span>流密钥</span>
            <strong>{streamKey ? '已就绪' : '待加载'}</strong>
          </div>
        </div>
      </section>

      <section className={styles.workspace}>
        <main className={styles.mainColumn}>
          <section className={styles.panel}>
            <div className={styles.panelHeader}>
              <div>
                <h2>站点登录</h2>
                <span>连接 mybilibili 网关服务</span>
              </div>
              <button
                className={styles.primaryButton}
                onClick={() =>
                  run('直播间已加载', () =>
                    MybilibiliLiveService.actions.return.loadOrCreateRoom(),
                  )
                }
                disabled={busy}
              >
                加载房间
              </button>
            </div>
            <div className={styles.formGrid}>
              <Field label="网关接口">
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
              <div className={cx(styles.accountBox, styles.fullField)}>
                <span>当前账号</span>
                <strong>{accountName}</strong>
                <small>{isLoggedIn ? '已保存访问凭证，可加载直播间' : '请先登录站点账号'}</small>
              </div>
            </div>
            <div className={styles.actionRow}>
              <button
                className={styles.primaryButton}
                onClick={() => run('登录成功，直播间已加载', loginToSite)}
                disabled={busy || !draftUsername.trim() || !draftPassword}
              >
                登录并加载房间
              </button>
              <button className={styles.secondaryButton} onClick={saveConnection} disabled={busy}>
                保存网关
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run('直播间已刷新', () =>
                    MybilibiliLiveService.actions.return.refreshRoom(),
                  )
                }
                disabled={busy || !isLoggedIn}
              >
                刷新房间
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run('登录状态已刷新', () =>
                    MybilibiliLiveService.actions.return.refreshLogin(),
                  )
                }
                disabled={busy || !isLoggedIn}
              >
                刷新登录
              </button>
              <button className={styles.secondaryButton} onClick={logoutFromSite} disabled={!isLoggedIn}>
                退出登录
              </button>
            </div>
          </section>

          <section className={styles.panel}>
            <div className={styles.panelHeader}>
              <div>
                <h2>OBS 推流</h2>
                <span>写入自定义 RTMP 配置</span>
              </div>
              <button
                className={styles.primaryButton}
                onClick={() =>
                  run('OBS 推流配置已写入', () =>
                    MybilibiliLiveService.actions.return.applyStreamSettings(),
                  )
                }
                disabled={busy || !room}
              >
                写入 OBS
              </button>
            </div>
            <div className={styles.streamRows}>
              <div className={styles.streamRow}>
                <span>推流地址</span>
                <code>{rtmpUrl || '未加载'}</code>
                <button onClick={() => copyText(rtmpUrl, '推流地址')} disabled={!rtmpUrl}>
                  复制
                </button>
              </div>
              <div className={styles.streamRow}>
                <span>流密钥</span>
                <code>{maskKey(streamKey)}</code>
                <button onClick={() => copyText(streamKey, '流密钥')} disabled={!streamKey}>
                  复制
                </button>
              </div>
            </div>
          </section>

          <section className={styles.panel}>
            <div className={styles.panelHeader}>
              <div>
                <h2>直播间信息</h2>
                <span>名称、分类与封面</span>
              </div>
              <button
                className={styles.primaryButton}
                onClick={() => run('直播间信息已保存', saveRoom)}
                disabled={busy || !room}
              >
                保存
              </button>
            </div>
            <div className={styles.formGrid}>
              <Field label="房间名称">
                <input
                  value={draftRoomName}
                  onChange={event => setDraftRoomName(event.target.value)}
                  placeholder="我的直播间"
                  disabled={!room}
                />
              </Field>
              <Field label="分类">
                <input
                  value={draftCategory}
                  onChange={event => setDraftCategory(event.target.value)}
                  placeholder="游戏 / 学习 / 生活"
                  disabled={!room}
                />
              </Field>
              <Field label="封面 URL" className={styles.fullField}>
                <input
                  value={draftCoverUrl}
                  onChange={event => setDraftCoverUrl(event.target.value)}
                  placeholder="https://..."
                  disabled={!room}
                />
              </Field>
            </div>
          </section>
        </main>

        <aside className={styles.sideColumn}>
          <section className={styles.sidePanel}>
            <div className={styles.sideHeader}>
              <h2>推流状态</h2>
              <span className={cx(styles.dot, isObsStreaming && styles.dotLive)} />
            </div>
            <dl className={styles.statusList}>
              <div>
                <dt>OBS 状态</dt>
                <dd>{statusLabel(obsStreamingStatus)}</dd>
              </div>
              <div>
                <dt>平台状态</dt>
                <dd>{statusLabel(roomStatus)}</dd>
              </div>
              <div>
                <dt>推流地址</dt>
                <dd>{rtmpUrl ? '已配置' : '未配置'}</dd>
              </div>
            </dl>
            <p className={styles.diagnosticHint}>
              平台状态应由 RTMP 服务回调更新；下面按钮只用于本地调试状态修正。
            </p>
            <div className={styles.actionRow}>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run('房间已标记为直播中', () =>
                    MybilibiliLiveService.actions.return.updateRoomStatus('live'),
                  )
                }
                disabled={busy || !room}
              >
                诊断标记直播
              </button>
              <button
                className={styles.secondaryButton}
                onClick={() =>
                  run('房间已标记为未开播', () =>
                    MybilibiliLiveService.actions.return.updateRoomStatus('offline'),
                  )
                }
                disabled={busy || !room}
              >
                诊断标记下播
              </button>
            </div>
          </section>
        </aside>
      </section>
    </div>
  );
}
