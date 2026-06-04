import { ViewHandler } from 'services/core/stateful-service';
import {
  ERecordingState,
  EReplayBufferState,
  EStreamingState,
  IGoLiveSettings,
  IStreamingServiceState,
} from './streaming-api';

export class StreamInfoView<T> {
  constructor(public readonly info: T) {}
}

export class StreamingServiceViews extends ViewHandler<IStreamingServiceState> {
  get streamingStatus() {
    return this.state.streamingStatus;
  }

  get recordingStatus() {
    return this.state.recordingStatus;
  }

  get replayBufferStatus() {
    return this.state.replayBufferStatus;
  }

  get isStreaming() {
    return this.state.streamingStatus !== EStreamingState.Offline;
  }

  get isRecording() {
    return this.state.recordingStatus !== ERecordingState.Offline;
  }

  get isMidStreamMode() {
    return this.isStreaming || this.isRecording;
  }

  get isDualOutputMode() {
    return this.state.dualOutputMode;
  }

  get isDualOutputRecording() {
    return false;
  }

  get isStreamShiftMode() {
    return false;
  }

  get enabledPlatforms(): string[] {
    return [];
  }

  get linkedPlatforms(): string[] {
    return [];
  }

  get allPlatforms(): string[] {
    return [];
  }

  get customDestinations() {
    return [];
  }

  get activeDisplayPlatforms() {
    return { horizontal: [], vertical: [] };
  }

  get settings(): IGoLiveSettings {
    return {
      platforms: {},
      customDestinations: [],
      advancedMode: false,
      recording: 'horizontal',
      streamShift: false,
      enhancedBroadcasting: false,
    };
  }

  getCanStreamDualOutput() {
    return false;
  }

  getOutputDisplayType(display: string) {
    return display === 'vertical' ? 'vertical' : 'horizontal';
  }
}
