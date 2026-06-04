import React, { useMemo } from 'react';
import cx from 'classnames';
import * as remote from '@electron/remote';
import { Tooltip } from 'antd';
import styles from './RecordingHistory.m.less';
import Scrollable from 'components-react/shared/Scrollable';
import { Services } from '../service-provider';
import { useController } from '../hooks/zustand';
import { useVuex } from '../hooks';
import { IRecordingEntry } from 'services/recording-mode';

const RecordingHistoryCtx = React.createContext<RecordingHistoryController | null>(null);

class RecordingHistoryController {
  private RecordingModeService = Services.RecordingModeService;

  get recordings() {
    return this.RecordingModeService.views.sortedRecordings;
  }

  formattedTimestamp = (timestamp: string) => {
    return this.RecordingModeService.views.formattedTimestamp(timestamp);
  };

  removeEntry = (timestamp: string) => {
    this.RecordingModeService.actions.removeRecordingEntry(timestamp);
  };

  showFile = (filename: string) => {
    remote.shell.showItemInFolder(filename);
  };
}

export default function RecordingHistoryPage(p: { className?: string }) {
  const controller = useMemo(() => new RecordingHistoryController(), []);
  return (
    <RecordingHistoryCtx.Provider value={controller}>
      <RecordingHistory className={p.className} />
    </RecordingHistoryCtx.Provider>
  );
}

export function RecordingHistory(p: { className?: string }) {
  const controller = useController(RecordingHistoryCtx);
  const { formattedTimestamp, showFile, removeEntry } = controller;
  const { recordings } = useVuex(() => ({
    recordings: controller.recordings,
  }));

  function openMarkersSettings() {
    Services.SettingsService.actions.showSettings('Hotkeys');
  }

  function RecordingActions(p: { recording: IRecordingEntry }) {
    return (
      <span className={styles.actionGroup}>
        <span className={styles.action} onClick={() => removeEntry(p.recording.timestamp)}>
          <i className="icon-trash" />
          &nbsp;
          <span>删除</span>
        </span>
      </span>
    );
  }

  return (
    <div className={cx(styles.container, p.className)}>
      <h1>录制历史</h1>
      <div style={{ marginBottom: 24, display: 'flex', flexDirection: 'column' }}>
        <span>这里显示本机 OBS 录制文件。点击文件路径可以在资源管理器中定位文件。</span>
        <span>
          <span className={styles.tipHighlight}>提示：</span>
          可以在热键设置里配置录制标记，用于给录制文件打时间点。
          <a onClick={openMarkersSettings} className={styles.tipLink}>
            去设置
          </a>
        </span>
      </div>
      <div className={styles.recordingsContainer} id="recordingHistory">
        <Scrollable style={{ height: '100%' }}>
          {recordings.length === 0 && (
            <div className={styles.empty}>暂无本地录制文件</div>
          )}
          {recordings.map(recording => (
            <div className={styles.recording} key={recording.timestamp}>
              {recording?.display && recording.display === 'vertical' && (
                <i className="icon-phone-case" style={{ paddingRight: '10px' }} />
              )}
              {recording?.display && recording.display === 'horizontal' && (
                <i className="icon-desktop" style={{ paddingRight: '10px' }} />
              )}
              <span style={{ marginRight: '8px' }}>{formattedTimestamp(recording.timestamp)}</span>
              <Tooltip title="在文件夹中显示">
                <span
                  data-test="filename"
                  onClick={() => showFile(recording.filename)}
                  className={styles.filename}
                >
                  {recording.filename}
                </span>
              </Tooltip>
              <RecordingActions recording={recording} />
            </div>
          ))}
        </Scrollable>
      </div>
    </div>
  );
}
