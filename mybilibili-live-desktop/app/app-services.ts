/**
 * mybilibili Live Desktop keeps local OBS control plus mybilibili live-room integration.
 * Streamlabs platform, widget, store, multistream, and membership services are not registered.
 */

export { InternalApiService } from 'services/api/internal-api';
export { IpcServerService } from 'services/api/ipc-server';
export { JsonrpcService } from 'services/api/jsonrpc';
export { AppService } from 'services/app';
export { CustomizationService } from 'services/customization';
export { HardwareService, DefaultHardwareService } from 'services/hardware';
export { I18nService } from 'services/i18n';
export { MetricsService } from 'services/metrics';
export { MybilibiliLiveService } from 'services/mybilibili';
export { NotificationsService } from 'services/notifications';
export { ObsUserPluginsService } from 'services/obs-user-plugins';
export { PerformanceService } from 'services/performance';
export { ProjectorService } from 'services/projector';
export { RealmService } from 'services/realm';
export { Scene, SceneItem, SceneItemFolder, ScenesService } from 'services/scenes';
export { ScenesTransitionsService } from 'services/scenes-transitions';
export { GlobalSelection, Selection, SelectionService } from 'services/selection';
export { SettingsService } from 'services/settings';
export { StreamSettingsService } from 'services/settings/streaming';
export { SettingsManagerService } from 'services/settings-manager';
export { VideoSettingsService } from 'services/settings-v2/video';
export { SignalsService } from 'services/signals-manager';
export { VideoService } from 'services/video';
export { WindowsService } from 'services/windows';

export { EditorCommandsService } from 'services/editor-commands';
export { EditorService } from 'services/editor';
export { FileManagerService } from 'services/file-manager';
export { FontLibraryService } from 'services/font-library';
export { MacPermissionsService } from 'services/mac-permissions';
