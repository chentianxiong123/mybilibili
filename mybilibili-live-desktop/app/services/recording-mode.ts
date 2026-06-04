import { PersistentStatefulService, ViewHandler, mutation } from 'services/core';

export interface IRecordingEntry {
  timestamp: string;
  filename: string;
  display?: 'horizontal' | 'vertical';
}

export interface IUploadInfo {
  uploading: boolean;
  uploadedBytes?: number;
  totalBytes?: number;
  error?: string;
}

interface IRecordingModeState {
  enabled: boolean;
  recordingHistory: Dictionary<IRecordingEntry>;
  uploadInfo: IUploadInfo;
}

class RecordingModeViews extends ViewHandler<IRecordingModeState> {
  get isRecordingModeEnabled() {
    return this.state.enabled;
  }

  get sortedRecordings() {
    return Object.values(this.state.recordingHistory);
  }

  formattedTimestamp(timestamp: string) {
    return timestamp;
  }
}

export class RecordingModeService extends PersistentStatefulService<IRecordingModeState> {
  static defaultState: IRecordingModeState = {
    enabled: false,
    recordingHistory: {},
    uploadInfo: { uploading: false },
  };

  get views() {
    return new RecordingModeViews(this.state);
  }

  setRecordingMode(enabled: boolean) {
    this.SET_RECORDING_MODE(enabled);
  }

  setUpRecordingFirstTimeSetup() {}

  cancelUpload() {
    this.SET_UPLOAD_INFO({ uploading: false });
  }

  @mutation()
  private SET_RECORDING_MODE(enabled: boolean) {
    this.state.enabled = enabled;
  }

  @mutation()
  private SET_UPLOAD_INFO(uploadInfo: Partial<IUploadInfo>) {
    this.state.uploadInfo = { uploading: false, ...uploadInfo };
  }
}
