import React from 'react';
import { Services } from '../../service-provider';
import { $t } from '../../../services/i18n';
import { CheckboxInput, ListInput, SliderInput } from '../../shared/inputs';
import { ObsSettingsSection } from './ObsSettings';
import styles from './Appearance.m.less';
import { CustomizationState } from 'services/customization';
import { useRealmObject } from 'components-react/hooks/realm';
import { bindFormState } from 'components-react/shared/inputs';

export function AppearanceSettings() {
  const {
    CustomizationService,
  } = Services;

  // Hooks up reactivity for Customization state
  useRealmObject(CustomizationService.state);

  const bind = bindFormState(
    () => CustomizationService.state.toObject() as CustomizationState,
    (newSettings: CustomizationState) => CustomizationService.setSettings(newSettings as any),
  );

  return (
    <div className={styles.container}>
      <ObsSettingsSection>
        <ListInput {...bind.theme} label={'主题'} options={CustomizationService.themeOptions} />
      </ObsSettingsSection>

      <ObsSettingsSection title={'直播侧栏'}>
        <CheckboxInput
          {...bind.leftDock}
          label={'在左侧显示直播侧栏'}
        />
        <SliderInput
          {...bind.chatZoomFactor}
          label={$t('Text Size')}
          tipFormatter={(val: number) => `${val * 100}%`}
          min={0.25}
          max={2}
          step={0.25}
        />

      </ObsSettingsSection>

      <ObsSettingsSection>
        <CheckboxInput
          {...bind.enableAnnouncements}
          label={'显示客户端更新提示'}
          className={styles.extraMargin}
        />
      </ObsSettingsSection>

      <ObsSettingsSection className={styles.extraMargin}>
        <ListInput
          {...bind.folderSelection}
          label={'场景元素选择方式'}
          options={[
            { value: true, label: '单击选择分组，双击选择元素' },
            {
              value: false,
              label: '双击选择分组，单击选择元素',
            },
          ]}
        />
      </ObsSettingsSection>
    </div>
  );
}
