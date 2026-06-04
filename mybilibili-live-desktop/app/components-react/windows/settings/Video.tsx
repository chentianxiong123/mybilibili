import React, { useEffect, useState } from 'react';
import { $t } from 'services/i18n';
import { Services } from '../../service-provider';
import { ObsGenericSettingsForm } from './ObsSettings';
import styles from './Common.m.less';

const BASE_RES_OPTIONS = [
  { label: '1920x1080', value: '1920x1080' },
  { label: '1280x720', value: '1280x720' },
];

const OUTPUT_RES_OPTIONS = [
  { label: '1920x1080', value: '1920x1080' },
  { label: '1536x864', value: '1536x864' },
  { label: '1440x810', value: '1440x810' },
  { label: '1280x720', value: '1280x720' },
  { label: '1152x648', value: '1152x648' },
  { label: '1096x616', value: '1096x616' },
  { label: '960x540', value: '960x540' },
  { label: '852x480', value: '852x480' },
  { label: '768x432', value: '768x432' },
  { label: '698x392', value: '698x392' },
  { label: '640x360', value: '640x360' },
];

const FPS_OPTIONS = [
  { label: '10', value: '10-1' },
  { label: '20', value: '20-1' },
  { label: '24', value: '24000-1001' },
  { label: '25', value: '25-1' },
  { label: '29.97', value: '30000-1001' },
  { label: '30', value: '30-1' },
  { label: '48', value: '48-1' },
  { label: '59.94', value: '60000-1001' },
  { label: '60', value: '60-1' },
];

export function VideoSettings() {
  const { VideoSettingsService, StreamingService } = Services;
  const [values, setValues] = useState<Record<string, any>>({});
  const cantEdit = StreamingService.views.isStreaming || StreamingService.views.isRecording;

  useEffect(() => {
    setValues({ ...(VideoSettingsService as any).values?.horizontal });
  }, []);

  function update(key: string, value: any) {
    setValues(prev => ({ ...prev, [key]: value }));
    (VideoSettingsService as any).actions.setSetting?.(key, value);
  }

  return (
    <div className={styles.container}>
      <ObsGenericSettingsForm title={$t('画布与输出')}>
        <ObsGenericSettingsForm.Selector
          label={$t('基础（画布）分辨率')}
          value={values.baseRes || '1920x1080'}
          options={BASE_RES_OPTIONS}
          onChange={(v: string) => update('baseRes', v)}
          disabled={cantEdit}
        />
        <ObsGenericSettingsForm.Selector
          label={$t('输出（缩放后）分辨率')}
          value={values.outputRes || '1920x1080'}
          options={OUTPUT_RES_OPTIONS}
          onChange={(v: string) => update('outputRes', v)}
          disabled={cantEdit}
        />
        <ObsGenericSettingsForm.Selector
          label={$t('帧率')}
          value={values.fpsType || '30-1'}
          options={FPS_OPTIONS}
          onChange={(v: string) => update('fpsType', v)}
          disabled={cantEdit}
        />
      </ObsGenericSettingsForm>
    </div>
  );
}