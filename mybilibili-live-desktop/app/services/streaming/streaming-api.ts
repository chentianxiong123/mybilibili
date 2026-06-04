import { Observable } from 'rxjs';

export enum EStreamingState {
  Offline = 'offline',
  Starting = 'starting',
  Live = 'live',
  Ending = 'ending',
  Reconnecting = 'reconnecting',
}

export enum ERecordingState {
  Offline = 'offline',
  Starting = 'starting',
  Recording = 'recording',
  Stopping = 'stopping',
  Start = 'start',
  Writing = 'writing',
  Wrote = 'wrote',
}

export enum EReplayBufferState {
  Running = 'running',
  Stopping = 'stopping',
  Offline = 'offline',
  Saving = 'saving',
  Wrote = 'wrote',
}

export type TDisplayOutput = 'horizontal' | 'vertical' | 'both';
export type TGoLiveChecklistItemState = 'not-started' | 'pending' | 'done' | 'failed';

export interface IStreamInfo {
  lifecycle: 'empty' | 'prepopulate' | 'waitForNewSettings' | 'runChecklist' | 'live';
  error: null;
  warning: '';
  settings: IGoLiveSettings | null;
  checklist: Record<string, TGoLiveChecklistItemState>;
}

export interface ICustomStreamDestination {
  name: string;
  url: string;
  streamKey?: string;
  enabled: boolean;
}

export interface IStreamSettings {
  platforms: Record<string, never>;
  customDestinations: ICustomStreamDestination[];
  advancedMode: boolean;
  recording: TDisplayOutput;
  streamShift?: false;
  enhancedBroadcasting?: false;
}

export interface IGoLiveSettings extends IStreamSettings {
  optimizedProfile?: never;
  tweetText?: string;
  prepopulateOptions?: Record<string, never>;
  streamShiftSettings?: never;
}

export interface IOutputStatus {
  streaming: EStreamingState;
  streamingTime: string;
  recording: ERecordingState;
  recordingTime: string;
  replayBuffer: EReplayBufferState;
  replayBufferTime: string;
}

export interface IStreamingServiceState {
  status: Record<'horizontal' | 'vertical', IOutputStatus>;
  streamingStatus: EStreamingState;
  streamingStatusTime: string;
  recordingStatus: ERecordingState;
  recordingStatusTime: string;
  replayBufferStatus: EReplayBufferState;
  replayBufferStatusTime: string;
  selectiveRecording: boolean;
  dualOutputMode: boolean;
  enhancedBroadcasting: boolean;
  info: IStreamInfo;
}

export interface IStreamingServiceApi {
  getModel(): IStreamingServiceState;
  streamingStatusChange: Observable<EStreamingState>;
  recordingStatusChange: Observable<ERecordingState>;
  replayBufferStatusChange: Observable<EReplayBufferState>;
  streamingStateChange: Observable<void>;
  startStreaming(): void;
  stopStreaming(): void;
  toggleStreaming(): Promise<void>;
  startRecording(): void;
  stopRecording(): void;
  toggleRecording(): void;
  startReplayBuffer(): void;
  stopReplayBuffer(): void;
}
