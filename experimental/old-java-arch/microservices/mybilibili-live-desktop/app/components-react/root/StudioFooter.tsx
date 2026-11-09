import React, { useCallback, useMemo } from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { EStreamQuality } from '../../services/performance';
import { EStreamingState, EReplayBufferState, ERecordingState } from '../../services/streaming';
import { Services } from '../service-provider';
import { $t } from '../../services/i18n';
import { useDebounce, useVuex } from '../hooks';
import PerformanceMetrics from '../shared/PerformanceMetrics';
import styles from './StudioFooter.m.less';

export default function StudioFooter() {
  const {
    StreamingService,
    WindowsService,
    UsageStatisticsService,
    PerformanceService,
    SettingsService,
  } = Services;

  const {
    streamingStatus,
    streamQuality,
    recordingStatus,
    replayBufferEnabled,
    replayBufferStatus,
    isReplayBufferActive,
  } = useVuex(
    () => ({
      streamingStatus: StreamingService.views.streamingStatus,
      streamQuality: PerformanceService.views.streamQuality,
      recordingStatus: StreamingService.views.recordingStatus,
      replayBufferEnabled: SettingsService.views.values.Output.RecRB,
      replayBufferStatus: StreamingService.views.replayBufferStatus,
      isReplayBufferActive: StreamingService.views.isReplayBufferActive,
    }),
    false,
  );

  const performanceIconClassName = useMemo(() => {
    if (!streamingStatus || streamingStatus === EStreamingState.Offline) return '';
    if (streamingStatus === EStreamingState.Reconnecting || streamQuality === EStreamQuality.POOR) {
      return 'warning';
    }
    if (streamQuality === EStreamQuality.FAIR) return 'info';
    return 'success';
  }, [streamingStatus, streamQuality]);

  const openMetricsWindow = useCallback(() => {
    WindowsService.actions.showWindow({
      componentName: 'AdvancedStatistics',
      title: $t('Performance Metrics'),
      size: { width: 700, height: 550 },
      resizable: true,
      maximizable: false,
      minWidth: 500,
      minHeight: 400,
    });
    UsageStatisticsService.actions.recordFeatureUsage('PerformanceStatistics');
  }, []);

  const toggleRecording = useCallback(() => {
    StreamingService.actions.toggleRecording();
  }, []);

  const toggleStreaming = useCallback(() => {
    StreamingService.actions.toggleStreaming();
  }, []);

  const toggleReplayBuffer = useCallback(() => {
    if (replayBufferStatus === EReplayBufferState.Offline) {
      StreamingService.actions.startReplayBuffer();
    } else {
      StreamingService.actions.stopReplayBuffer();
    }
  }, [replayBufferStatus]);

  const saveReplay = useCallback(() => {
    StreamingService.actions.saveReplay();
  }, []);

  const isRecording = recordingStatus === ERecordingState.Recording;
  const recordingBusy = [
    ERecordingState.Starting,
    ERecordingState.Stopping,
    ERecordingState.Writing,
  ].includes(recordingStatus);
  const isStreaming = streamingStatus !== EStreamingState.Offline;
  const streamingBusy = [EStreamingState.Starting, EStreamingState.Ending].includes(streamingStatus);
  const replayBufferOffline = replayBufferStatus === EReplayBufferState.Offline;
  const replayBufferBusy = [
    EReplayBufferState.Starting,
    EReplayBufferState.Stopping,
    EReplayBufferState.Saving,
  ].includes(replayBufferStatus);

  return (
    <div className={cx('footer', styles.footer)}>
      <div className={cx('flex flex--center flex--grow flex--justify-start', styles.footerLeft)}>
        <Tooltip placement="left" title={$t('Open Performance Window')}>
          <i
            className={cx(
              'icon-leaderboard-4',
              'metrics-icon',
              styles.metricsIcon,
              performanceIconClassName,
            )}
            onClick={openMetricsWindow}
          />
        </Tooltip>
        <PerformanceMetrics mode="limited" className="performance-metrics" />
      </div>

      <div className={styles.navRight}>
        <div className={styles.navItem}>
          <Tooltip placement="left" title={isRecording ? $t('Stop Recording') : $t('Record')}>
            <button
              className={cx(styles.recordButton, 'record-button', { active: isRecording })}
              onClick={useDebounce(200, toggleRecording)}
            >
              <span>{recordingBusy ? <i className="fa fa-spinner fa-pulse" /> : <>REC</>}</span>
            </button>
          </Tooltip>
        </div>

        {replayBufferEnabled && replayBufferOffline && (
          <div className={styles.navItem}>
            <Tooltip placement="left" title={$t('Start Replay Buffer')}>
              <button className="circle-button" onClick={toggleReplayBuffer}>
                <i className="icon-replay-buffer" />
              </button>
            </Tooltip>
          </div>
        )}

        {isReplayBufferActive && (
          <div className={cx(styles.navItem, styles.replayButtonGroup)}>
            <Tooltip placement="left" title={$t('Stop')}>
              <button className={cx('circle-button', styles.leftReplay)} onClick={toggleReplayBuffer}>
                {replayBufferBusy ? <i className="fa fa-spinner fa-pulse" /> : <i className="fa fa-stop" />}
              </button>
            </Tooltip>
            <Tooltip placement="right" title={$t('Save Replay')}>
              <button className={cx('circle-button', styles.rightReplay)} onClick={saveReplay}>
                <i className="icon-save" />
              </button>
            </Tooltip>
          </div>
        )}

        <div className={styles.navItem}>
          <button
            className={cx('button', isStreaming ? 'button--warning' : 'button--action')}
            onClick={useDebounce(200, toggleStreaming)}
          >
            {streamingBusy && <i className="fa fa-spinner fa-pulse" />}
            {!streamingBusy && (isStreaming ? $t('End Stream') : $t('Go Live'))}
          </button>
        </div>
      </div>
    </div>
  );
}
