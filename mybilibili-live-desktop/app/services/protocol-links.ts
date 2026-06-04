import electron from 'electron';
import { Service } from 'services/core/service';

export interface IAppProtocolLink {
  url: string;
  base: string;
  path: string;
  query: URLSearchParams;
  appId: string;
}

export class ProtocolLinksService extends Service {
  start(argv: string[]) {
    electron.ipcRenderer.on('protocolLink', (_event: Electron.Event, link: string) => {
      this.handleLink(link);
    });
    argv.filter(arg => arg.startsWith('mybilibili://')).forEach(arg => this.handleLink(arg));
  }

  private handleLink(link: string) {
    console.info('[mybilibili] protocol link ignored in local OBS shell', link);
  }
}
