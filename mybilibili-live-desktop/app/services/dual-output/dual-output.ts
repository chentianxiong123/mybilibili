import { Subject } from 'rxjs';
import { mutation, PersistentStatefulService, ViewHandler } from 'services/core';

type TDisplayType = 'horizontal' | 'vertical';

interface IDualOutputServiceState {
  dualOutputMode: boolean;
  videoSettings: Record<TDisplayType, any> & {
    activeDisplays: Record<TDisplayType, boolean>;
  };
  isLoading: boolean;
}

class DualOutputViews extends ViewHandler<IDualOutputServiceState> {
  get isLoading() {
    return this.state.isLoading;
  }

  get dualOutputMode() {
    return false;
  }

  get isDualOutputCollection() {
    return false;
  }

  get hasSceneNodeMaps() {
    return false;
  }

  get sceneNodeMaps() {
    return {};
  }

  get activeSceneNodeMap() {
    return {};
  }

  get activeDisplays() {
    return this.state.videoSettings.activeDisplays;
  }

  get showBothDisplays() {
    return true;
  }

  get hideBothDisplays() {
    return false;
  }

  get videoSettings() {
    return this.state.videoSettings;
  }

  hasNodeMap(_sceneId?: string) {
    return false;
  }

  getDualOutputNodeId(id: string) {
    return id;
  }

  getVerticalNodeId(id: string) {
    return id;
  }

  getHorizontalNodeId(id: string) {
    return id;
  }

  getNodeDisplay(_nodeId: string, _sceneId?: string): TDisplayType {
    return 'horizontal';
  }

  getEnabledTargets() {
    return { horizontal: [], vertical: [] };
  }
}

export class DualOutputService extends PersistentStatefulService<IDualOutputServiceState> {
  static defaultState: IDualOutputServiceState = {
    dualOutputMode: false,
    videoSettings: {
      horizontal: null,
      vertical: null,
      activeDisplays: { horizontal: true, vertical: false },
    },
    isLoading: false,
  };

  sceneNodeHandled = new Subject<number>();
  collectionHandled = new Subject<Record<string, Dictionary<string>> | null>();
  dualOutputModeChanged = new Subject<boolean>();
  displayToggled = new Subject<void>();

  get views() {
    return new DualOutputViews(this.state);
  }

  setDualOutputModeIfPossible(status = false) {
    this.setDualOutputMode(status);
  }

  setDualOutputMode(status = false) {
    this.SET_DUAL_OUTPUT_MODE(status);
    this.dualOutputModeChanged.next(status);
  }

  toggleDualOutputMode(status = false) {
    this.setDualOutputMode(status);
  }

  setIsLoading(status: boolean) {
    this.SET_IS_LOADING(status);
  }

  toggleDisplay(visible: boolean, display: TDisplayType) {
    this.SET_DISPLAY(visible, display);
    this.displayToggled.next();
  }

  createOrAssignOutputNode(..._args: unknown[]) {
    return null;
  }

  createPartnerNode(..._args: unknown[]) {
    return null;
  }

  validateDualOutputCollection() {
    this.collectionHandled.next(null);
  }

  convertSingleOutputToDualOutputCollection() {
    this.collectionHandled.next(null);
  }

  setVideoSetting(settings: Dictionary<any>, display: TDisplayType = 'horizontal') {
    this.UPDATE_VIDEO_SETTING(settings, display);
  }

  updateVideoSettings(settings: Dictionary<any>, display: TDisplayType = 'horizontal') {
    this.UPDATE_VIDEO_SETTING(settings, display);
  }

  @mutation()
  private SET_DUAL_OUTPUT_MODE(status: boolean) {
    this.state.dualOutputMode = false;
  }

  @mutation()
  private SET_IS_LOADING(status: boolean) {
    this.state.isLoading = status;
  }

  @mutation()
  private SET_DISPLAY(visible: boolean, display: TDisplayType) {
    this.state.videoSettings.activeDisplays[display] = visible;
  }

  @mutation()
  private UPDATE_VIDEO_SETTING(settings: Dictionary<any>, display: TDisplayType) {
    this.state.videoSettings[display] = {
      ...(this.state.videoSettings[display] || {}),
      ...settings,
    };
  }
}
