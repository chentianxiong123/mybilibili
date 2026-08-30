import { StatefulService, ViewHandler } from 'services/core/stateful-service';

export enum EAvailableFeatures {
  platform = 'mybilibili-platform',
  creatorSites = 'mybilibili-creator-sites',
  highlighter = 'mybilibili-highlighter',
  themeAudit = 'mybilibili-theme-audit',
  sharedStorage = 'mybilibili-shared-storage',
  verticalRecording = 'mybilibili-vertical-recording',
}

interface IIncrementalRolloutServiceState {
  availableFeatures: string[];
}

class IncrementalRolloutView extends ViewHandler<IIncrementalRolloutServiceState> {
  get availableFeatures() {
    return this.state.availableFeatures || [];
  }

  featureIsEnabled(feature: EAvailableFeatures): boolean {
    return this.availableFeatures.includes(feature);
  }
}

export class IncrementalRolloutService extends StatefulService<IIncrementalRolloutServiceState> {
  static initialState: IIncrementalRolloutServiceState = {
    availableFeatures: [],
  };

  featuresReady = Promise.resolve();

  get views() {
    return new IncrementalRolloutView(this.state);
  }

  fetchAvailableFeatures() {}

  setCommandLineFeatures() {}

  resetAvailableFeatures() {}
}
