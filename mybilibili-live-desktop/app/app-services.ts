/**
 * mybilibili 直播工作台 keeps local OBS control plus mybilibili live-room integration.
 * Only local OBS control and mybilibili live-room integration are registered.
 */

export { InternalApiService } from 'services/api/internal-api';
export { IpcServerService } from 'services/api/ipc-server';
export { JsonrpcService } from 'services/api/jsonrpc';
export { AppService } from 'services/app';
export { CustomizationService } from 'services/customization';
export { DiagnosticsService } from 'services/diagnostics';
export { DualOutputService } from 'services/dual-output';
export { ExternalApiService } from 'services/api/external-api';
export { HardwareService, DefaultHardwareService } from 'services/hardware';
export { HostsService } from 'services/hosts';
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
export { EncoderQueryService, OutputSettingsService } from 'services/settings/output';
export { StreamSettingsService } from 'services/settings/streaming';
export { SettingsManagerService } from 'services/settings-manager';
export { SideNavService } from 'services/side-nav';
export { LiveOutputRuntimeService, StreamingService } from 'services/streaming';
export { VideoSettingsService } from 'services/settings-v2/video';
export { SignalsService } from 'services/signals-manager';
export { UsageStatisticsService } from 'services/usage-statistics';
export { UserService } from 'services/user';
export { VideoService } from 'services/video';
export { WebsocketService } from 'services/websocket';
export { WindowsService } from 'services/windows';

export { EditorCommandsService } from 'services/editor-commands';
export { EditorService } from 'services/editor';
export { FileManagerService } from 'services/file-manager';
export { FontLibraryService } from 'services/font-library';
export { MacPermissionsService } from 'services/mac-permissions';
export { NavigationService } from 'services/navigation';
export { RecordingModeService } from 'services/recording-mode';
export { SourcesService } from 'services/sources';
export { VirtualWebcamService } from 'services/virtual-webcam';
