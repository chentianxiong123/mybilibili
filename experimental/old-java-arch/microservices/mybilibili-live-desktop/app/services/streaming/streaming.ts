import { Subject } from 'rxjs';
import { mutation, StatefulService } from 'services/core/stateful-service';
import { Inject } from 'services/core/injector';
import {
  ERecordingState,
  EReplayBufferState,
  EStreamingState,
  IStreamingServiceApi,
  IStreamingServiceState,
} from './streaming-api';
import { StreamingServiceViews, StreamInfoView } from './streaming-view';
import { LiveOutputRuntimeService } from './live-output-runtime';

type TDisplayType = 'horizontal' | 'vertical';

function now() {
  return new Date().toISOString();
}

function emptyStatus() {
  return {
    streaming: EStreamingState.Offline,
    streamingTime: now(),
    recording: ERecordingState.Offline,
    recordingTime: now(),
    replayBuffer: EReplayBufferState.Offline,
    replayBufferTime: now(),
  };
}

export class StreamingService
  extends StatefulService<IStreamingServiceState>
  implements IStreamingServiceApi {
  @Inject() private liveOutputRuntimeService: LiveOutputRuntimeService;

  static initialState: IStreamingServiceState = {
    status: {
      horizontal: emptyStatus(),
      vertical: emptyStatus(),
    },
    streamingStatus: EStreamingState.Offline,
    streamingStatusTime: now(),
    recordingStatus: ERecordingState.Offline,
    recordingStatusTime: now(),
    replayBufferStatus: EReplayBufferState.Offline,
    replayBufferStatusTime: now(),
    selectiveRecording: false,
    dualOutputMode: false,
    enhancedBroadcasting: false,
    info: {
      lifecycle: 'empty',
      error: null,
      warning: '',
      settings: null,
      checklist: {},
    },
  };

  streamingStatusChange = new Subject<EStreamingState>();
  recordingStatusChange = new Subject<ERecordingState>();
  replayBufferStatusChange = new Subject<EReplayBufferState>();
  replayBufferFileWrite = new Subject<string>();
  streamInfoChanged = new Subject<StreamInfoView<unknown>>();
  signalInfoChanged = new Subject<unknown>();
  latestRecordingPath = new Subject<string>();
  streamErrorCreated = new Subject<string>();
  streamShiftEvent = new Subject<unknown>();
  streamingStateChange = new Subject<void>();

  get views() {
    return new StreamingServiceViews(this.state);
  }

  getModel() {
    return this.state;
  }

  get isStreaming() {
    return this.state.streamingStatus !== EStreamingState.Offline;
  }

  get isRecording() {
    return this.state.recordingStatus !== ERecordingState.Offline;
  }

  get isIdle() {
    return !this.isStreaming && !this.isRecording;
  }

  get formattedDurationInCurrentStreamingState() {
    return '00:00:00';
  }

  get formattedDurationInCurrentRecordingState() {
    return '00:00:00';
  }

  get streamingPerformanceStats() {
    return null;
  }

  get recordingPerformanceStats() {
    return null;
  }

  async startStreaming() {
    if (this.isStreaming) {
      throw new Error('直播输出已经在运行');
    }

    this.setStreamingStatus(EStreamingState.Starting);
    try {
      await this.liveOutputRuntimeService.start();
      this.setStreamingStatus(EStreamingState.Live);
    } catch (error: unknown) {
      this.setStreamingStatus(EStreamingState.Offline);
      throw error;
    }
  }

  async stopStreaming() {
    if (!this.isStreaming) {
      throw new Error('直播输出未运行');
    }

    this.setStreamingStatus(EStreamingState.Ending);
    try {
      await this.liveOutputRuntimeService.stop();
    } finally {
      this.setStreamingStatus(EStreamingState.Offline);
    }
  }

  async toggleStreaming() {
    this.isStreaming ? await this.stopStreaming() : await this.startStreaming();
  }

  startRecording() {
    this.setRecordingStatus(ERecordingState.Recording);
  }

  stopRecording() {
    this.setRecordingStatus(ERecordingState.Offline);
  }

  toggleRecording() {
    this.isRecording ? this.stopRecording() : this.startRecording();
  }

  startReplayBuffer() {
    this.setReplayBufferStatus(EReplayBufferState.Running);
  }

  stopReplayBuffer() {
    this.setReplayBufferStatus(EReplayBufferState.Offline);
  }

  saveReplay() {}

  splitFile(_display: TDisplayType = 'horizontal') {}

  setSelectiveRecording(enabled: boolean) {
    this.SET_SELECTIVE_RECORDING(enabled);
  }

  setDualOutputMode(enabled: boolean) {
    this.SET_DUAL_OUTPUT_MODE(enabled);
  }

  getStreamingInstance() {
    return this.liveOutputRuntimeService.instance;
  }

  getRecordingInstance() {
    return null;
  }

  showGoLiveWindow() {}

  private setStreamingStatus(status: EStreamingState) {
    this.SET_STREAMING_STATUS(status, now());
    this.streamingStatusChange.next(status);
    this.streamingStateChange.next();
  }

  private setRecordingStatus(status: ERecordingState) {
    this.SET_RECORDING_STATUS(status, now());
    this.recordingStatusChange.next(status);
  }

  private setReplayBufferStatus(status: EReplayBufferState) {
    this.SET_REPLAY_BUFFER_STATUS(status, now());
    this.replayBufferStatusChange.next(status);
  }

  @mutation()
  private SET_STREAMING_STATUS(status: EStreamingState, time: string) {
    this.state.streamingStatus = status;
    this.state.streamingStatusTime = time;
    Object.values(this.state.status).forEach(display => {
      display.streaming = status;
      display.streamingTime = time;
    });
  }

  @mutation()
  private SET_RECORDING_STATUS(status: ERecordingState, time: string) {
    this.state.recordingStatus = status;
    this.state.recordingStatusTime = time;
    Object.values(this.state.status).forEach(display => {
      display.recording = status;
      display.recordingTime = time;
    });
  }

  @mutation()
  private SET_REPLAY_BUFFER_STATUS(status: EReplayBufferState, time: string) {
    this.state.replayBufferStatus = status;
    this.state.replayBufferStatusTime = time;
    Object.values(this.state.status).forEach(display => {
      display.replayBuffer = status;
      display.replayBufferTime = time;
    });
  }

  @mutation()
  private SET_SELECTIVE_RECORDING(enabled: boolean) {
    this.state.selectiveRecording = enabled;
  }

  @mutation()
  private SET_DUAL_OUTPUT_MODE(enabled: boolean) {
    this.state.dualOutputMode = enabled;
  }
}
