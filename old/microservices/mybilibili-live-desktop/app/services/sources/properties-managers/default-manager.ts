import { PropertiesManager } from './properties-manager';
import { TObsValue } from 'components/obs/inputs/ObsInput';

export interface IDefaultManagerSettings {}

export class DefaultManager extends PropertiesManager {
  settings: IDefaultManagerSettings;

  handleSettingsChange(_settings: Dictionary<TObsValue>) {}
}
