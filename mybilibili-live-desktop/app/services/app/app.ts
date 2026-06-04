import electron from 'electron';
import * as remote from '@electron/remote';
import { Subject } from 'rxjs';
import { mutation, StatefulService } from 'services/core/stateful-service';
import { Inject } from 'services/core/injector';
import { IpcServerService } from 'services/api/ipc-server';
import { PerformanceService } from 'services/performance';
import { RealmService } from 'services/realm';
import { WindowsService } from 'services/windows';
import { MybilibiliLiveService } from 'services/mybilibili';

interface IAppState {
  loading: boolean;
  argv: string[];
  errorAlert: boolean;
  onboarded: boolean;
}

export interface IRunInLoadingModeOptions {
  hideStyleBlockers?: boolean;
}

export class AppService extends StatefulService<IAppState> {
  @Inject() private realmService: RealmService;
  @Inject() private windowsService: WindowsService;
  @Inject() private ipcServerService: IpcServerService;
  @Inject() private performanceService: PerformanceService;
  @Inject() private mybilibiliLiveService: MybilibiliLiveService;

  static initialState: IAppState = {
    loading: true,
    argv: remote.process.argv,
    errorAlert: false,
    onboarded: true,
  };

  readonly appDataDirectory = remote.app.getPath('userData');
  loadingChanged = new Subject<boolean>();

  async load() {
    if (electron.ipcRenderer) {
      electron.ipcRenderer.on('shutdown', () => {
        this.windowsService.hideMainWindow();
        electron.ipcRenderer.send('acknowledgeShutdown');
        this.shutdownHandler();
      });
    }

    await this.realmService.connect();
    this.ipcServerService.listen();
    this.performanceService.startMonitoringPerformance();

    try {
      this.mybilibiliLiveService.applyStreamSettings();
    } catch {
      // Stream settings need a loaded room. The control panel can load it later.
    }

    this.SET_LOADING(false);
  }

  async shutdownHandler() {
    this.performanceService.stop();
    this.ipcServerService.stopListening();
    this.windowsService.shutdown();
    electron.ipcRenderer.send('shutdownComplete');
  }

  showErrorAlert() {
    this.SET_ERROR_ALERT(true);
  }

  @mutation()
  SET_LOADING(loading: boolean) {
    this.state.loading = loading;
    this.loadingChanged.next(loading);
  }

  @mutation()
  SET_ERROR_ALERT(errorAlert: boolean) {
    this.state.errorAlert = errorAlert;
  }
}
