import React from 'react';
import { SwitchInput, ListInput } from 'components-react/shared/inputs';
import { Services } from 'components-react/service-provider';
import { useVuex } from 'components-react/hooks';
import { $t } from 'services/i18n';
import Utils from 'services/utils';
import styles from './AdvancedAudio.m.less';

const trackOptions = [1, 2, 3, 4, 5, 6].map(value => ({ label: String(value), value }));

export default function GlobalSettings() {
  const { SettingsService, DefaultHardwareService } = Services;

  const { audioTracks, streamTrack, vodTrack, vodTrackEnabled, enableMuteNotifications } = useVuex(
    () => ({
      audioTracks: SettingsService.views.audioTracks,
      streamTrack: SettingsService.views.streamTrack,
      vodTrack: SettingsService.views.vodTrack,
      vodTrackEnabled: SettingsService.views.vodTrackEnabled,
      enableMuteNotifications: DefaultHardwareService.state.enableMuteNotifications,
    }),
  );

  return (
    <div style={{ display: 'grid', gap: 12 }}>
      <ListInput
        label={$t('Streaming Track')}
        value={streamTrack + 1}
        options={trackOptions}
        onChange={value => SettingsService.actions.setSettingValue('Output', 'TrackIndex', value)}
        layout="vertical"
      />
      <SwitchInput
        label={$t('Enable Twitch VOD Track')}
        value={vodTrackEnabled}
        onChange={value => SettingsService.actions.setSettingValue('Output', 'VodTrackEnabled', value)}
      />
      {vodTrackEnabled && (
        <ListInput
          label={$t('Twitch VOD Track')}
          value={vodTrack + 1}
          options={trackOptions.filter(opt => opt.value !== streamTrack + 1)}
          onChange={value => SettingsService.actions.setSettingValue('Output', 'VodTrackIndex', value)}
          layout="vertical"
        />
      )}
      <div className={styles.globalAudioTracks}>
        {audioTracks.map((track, i) => (
          <SwitchInput
            key={i}
            label={String(i + 1)}
            value={!!track}
            onChange={value => {
              const newArray = [...audioTracks];
              newArray[i] = Number(value);
              SettingsService.actions.setSettingValue(
                'Output',
                'RecTracks',
                Utils.binnaryArrayToNumber([...newArray].reverse()),
              );
            }}
          />
        ))}
      </div>
      <SwitchInput
        label={$t('Enable Muted Notifications')}
        value={enableMuteNotifications}
        onChange={() => DefaultHardwareService.actions.toggleMuteNotifications()}
      />
    </div>
  );
}
