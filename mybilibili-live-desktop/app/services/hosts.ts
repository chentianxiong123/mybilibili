import { Service } from './core/service';

export class HostsService extends Service {
  get mybilibili() {
    return 'localhost';
  }

  get overlays() {
    return 'localhost';
  }

  get media() {
    return 'localhost';
  }

  get io() {
    return 'http://localhost';
  }

  get cdn() {
    return 'localhost';
  }

  get platform() {
    return 'localhost';
  }

  get analitycs() {
    return 'localhost';
  }
}

export class UrlService extends Service {
  get protocol() {
    return 'http://';
  }

  getMybilibiliApi(endpoint: string) {
    return `http://localhost/api/${endpoint}`;
  }

  get supportLink() {
    return 'http://localhost';
  }
}
