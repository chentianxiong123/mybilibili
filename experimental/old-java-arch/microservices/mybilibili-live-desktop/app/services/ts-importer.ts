import { StatefulService } from 'services/core/stateful-service';

export class TsImporterService extends StatefulService<{ progress: number; total: number }> {
  static initialState = { progress: 0, total: 0 };

  async import() {
    throw new Error('旧场景包导入器已从 mybilibili 直播工作台中移除。');
  }
}
