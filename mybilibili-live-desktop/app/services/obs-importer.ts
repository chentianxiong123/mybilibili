import { StatefulService } from 'services/core/stateful-service';

export class ObsImporterService extends StatefulService<{ progress: number; total: number }> {
  static initialState = { progress: 0, total: 0 };

  async import() {
    throw new Error('旧 OBS 导入器已从 mybilibili Live Desktop 中移除。');
  }
}
