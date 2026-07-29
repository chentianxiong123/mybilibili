import { Inject, mutation, PersistentStatefulService } from 'services/core';
import { ISettingsSubCategory, SettingsService } from 'services/settings';

export interface IStreamSettingsState {
  protectedModeEnabled: boolean;
  protectedModeMigrationRequired: boolean;
  title: string;
  description: string;
  warnNoVideoSources: boolean;
}

export interface ICustomStreamDestination {
  name: string;
  url: string;
  streamKey?: string;
  enabled: boolean;
}

export interface IStreamSettings extends IStreamSettingsState {
  streamType: string;
  server: string;
  key: string;
}

export class StreamSettingsService extends PersistentStatefulService<IStreamSettingsState> {
  @Inject() private settingsService: SettingsService;

  static defaultState: IStreamSettingsState = {
    protectedModeEnabled: false,
    protectedModeMigrationRequired: false,
    title: '',
    description: '',
    warnNoVideoSources: true,
  };

  setSettings(patch: Partial<IStreamSettings>) {
    const current = this.settingsService.state.Stream?.formData || [];
    const formData: ISettingsSubCategory[] = current.map(subCategory => ({
      ...subCategory,
      parameters: subCategory.parameters.map(parameter => {
        if (parameter.name === 'streamType' && patch.streamType !== undefined) {
          return { ...parameter, value: patch.streamType };
        }
        if (parameter.name === 'server' && patch.server !== undefined) {
          return { ...parameter, value: patch.server };
        }
        if (parameter.name === 'key' && patch.key !== undefined) {
          return { ...parameter, value: patch.key };
        }
        return parameter;
      }),
    }));

    if (formData.length) this.settingsService.setSettings('Stream', formData);

    const localPatch: Partial<IStreamSettingsState> = {};
    ([
      'protectedModeEnabled',
      'protectedModeMigrationRequired',
      'title',
      'description',
      'warnNoVideoSources',
    ] as (keyof IStreamSettingsState)[]).forEach(key => {
      if (key in patch) localPatch[key] = patch[key] as never;
    });
    if (Object.keys(localPatch).length) this.SET_LOCAL_STORAGE_SETTINGS(localPatch);
  }

  @mutation()
  private SET_LOCAL_STORAGE_SETTINGS(patch: Partial<IStreamSettingsState>) {
    Object.assign(this.state, patch);
  }
}
