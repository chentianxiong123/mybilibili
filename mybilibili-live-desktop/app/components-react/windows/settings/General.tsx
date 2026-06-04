import React, { useState } from 'react';
import { ObsGenericSettingsForm, ObsSettingsSection } from './ObsSettings';
import { $t, I18nService } from '../../../services/i18n';
import { confirmAsync } from '../../modals';
import { CheckboxInput, ListInput } from '../../shared/inputs';
import { Services } from '../../service-provider';
import fs from 'fs';
import path from 'path';
import { useVuex } from 'components-react/hooks';

export function GeneralSettings() {
  return (
    <div>
      <LanguageSettings />
      <ExtraSettings />
      <ObsGenericSettingsForm page="General" />
    </div>
  );
}

function LanguageSettings() {
  const i18nService = I18nService.instance as I18nService;
  const localeOptions = i18nService.state.localeList;
  const currentLocale = i18nService.state.locale;

  async function save(lang: string) {
    if (!(await confirmAsync('切换语言需要重启应用，是否继续？'))) {
      return;
    }
    i18nService.actions.setLocale(lang);
  }

  return (
    <ObsSettingsSection>
      <ListInput options={localeOptions} label="语言" onChange={save} value={currentLocale} />
    </ObsSettingsSection>
  );
}

function ExtraSettings() {
  const {
    AppService,
    OnboardingService,
    WindowsService,
    RecordingModeService,
  } = Services;
  const disableHAFilePath = path.join(AppService.appDataDirectory, 'HADisable');
  const [disableHA, setDisableHA] = useState(() => fs.existsSync(disableHAFilePath));

  // TODO: unused fields
  const { recordingMode } = useVuex(() => ({
    recordingMode: RecordingModeService.views.isRecordingModeEnabled,
  }));

  function configureDefaults() {
    OnboardingService.actions.start({ isHardware: true });
    WindowsService.actions.closeChildWindow();
  }

  function importFromObs() {
    // TODO: there's no check that OBS is installed like in Onboarding
    OnboardingService.actions.setImport('obs');
    OnboardingService.actions.start({ isImport: true });
    WindowsService.actions.closeChildWindow();
  }

  function disableHardwareAcceleration(val: boolean) {
    try {
      if (val) {
        // Touch the file
        fs.closeSync(fs.openSync(disableHAFilePath, 'w'));
        setDisableHA(true);
      } else {
        fs.unlinkSync(disableHAFilePath);
        setDisableHA(false);
      }
    } catch (e: unknown) {
      console.error('Error setting hardware acceleration', e);
    }
  }

  return (
    <>
      <ObsSettingsSection>
        <CheckboxInput
          label={$t('Disable hardware acceleration (requires restart)')}
          value={disableHA}
          onChange={disableHardwareAcceleration}
          name="disable_ha"
        />
        <CheckboxInput
          label={$t('Disable live streaming features (Recording Only mode)')}
          value={recordingMode}
          onChange={RecordingModeService.actions.setRecordingMode}
        />

      </ObsSettingsSection>

      <ObsSettingsSection>
        <div className="actions">
          <div className="input-container">
            <button className="button button--default" onClick={configureDefaults}>
              配置默认设备
            </button>
          </div>
          <div className="input-container">
            <button className="button button--default" onClick={importFromObs}>
              导入 OBS 配置
            </button>
          </div>
        </div>
      </ObsSettingsSection>
    </>
  );
}
