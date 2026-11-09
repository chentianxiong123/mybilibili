import { Service } from 'services/core';

export class UsageStatisticsService extends Service {
  getSysInfo() {
    return {};
  }

  recordEvent(_event: string, _metadata: object = {}) {}

  recordAnalyticsEvent(_event: string, _value: any = {}) {}

  recordClick(_component: string, _target: string) {}

  recordShown(_component: string, _target: string) {}

  recordFeatureUsage(_feature: string) {}

  flush() {}
}
