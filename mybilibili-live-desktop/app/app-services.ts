/**
 * 全部服务在本文件注册
 *
 * mybilibili Live Desktop 派生版本只保留本地 OBS + mybilibili 自有推流。
 * 旧 Streamlabs 多平台 / 高级会员 / 高光剪辑 / 小组件 / 应用商店 已移除。
 */

// OFFLINE SERVICES
export { AppService } from 'services/app';
export { InternalApiService } from 'services/api/internal-api';
export { ExternalApiService } from 'services/api/external-api';
export { ExternalApiLimitsService } from 'services/api/external-api-limits';
export { SourcesService, Source } from 'services/sources';
export { Scene, SceneItem, SceneItemFolder, ScenesService } from 'services/scenes';
export { ObsImporterService } from 'services/obs-importer';
export { ClipboardService } from 'services/clipboard';
export { AudioService, AudioSource } from 'services/audio';
export { HostsService, UrlService } from 'services/hosts';
export { Hotkey, HotkeysService } from 'services/hotkeys';
export { KeyListenerService } from 'services/key-listener';
export { ShortcutsService } from 'services/shortcuts';
export { CustomizationService } from 'services/customization';
export { NotificationsService } from 'services/notifications';
export { OnboardingService } from 'services/onboarding';
export { NavigationService } from 'services/navigation';
export { PerformanceService } from 'services/performance';
export { SettingsService, OutputSettingsService, EncoderQueryService } from 'services/settings';
export { VideoService } from 'services/video';
export { WindowsService } from 'services/windows';
export { TransitionsService } from 'services/transitions';
export { FontLibraryService } from 'services/font-library';
export { SourceFiltersService } from 'services/source-filters';
export { CacheUploaderService } from 'services/cache-uploader';
export { TcpServerService } from 'services/api/tcp-server';
export { IpcServerService } from 'services/api/ipc-server';
export { JsonrpcService } from 'services/api/jsonrpc';
export { DismissablesService } from 'services/dismissables';
export { SceneCollectionsServerApiService } from 'services/scene-collections/server-api';
export { SceneCollectionsService } from 'services/scene-collections';
export { GlobalSelection, Selection, SelectionService } from 'services/selection';
export { OverlaysPersistenceService } from 'services/scene-collections/overlays';
export { SceneCollectionsStateService } from 'services/scene-collections/state';
export { FileManagerService } from 'services/file-manager';
export { ProtocolLinksService } from 'services/protocol-links';
export { ProjectorService } from 'services/projector';
export { I18nService } from 'services/i18n';
export { ObsUserPluginsService } from 'services/obs-user-plugins';
export { HardwareService, DefaultHardwareService } from 'services/hardware';
export { EditorCommandsService } from 'services/editor-commands';
export { EditorService } from 'services/editor';
export { SignalsService } from 'services/signals-manager';
export { StreamSettingsService } from 'services/settings/streaming';
export { TouchBarService } from 'services/touch-bar';
export { ApplicationMenuService } from 'services/application-menu';
export { MacPermissionsService } from 'services/mac-permissions';
export { VirtualWebcamService } from 'services/virtual-webcam';
export { MetricsService } from 'services/metrics';
export { DiagnosticsService } from 'services/diagnostics';
export { RecordingModeService } from 'services/recording-mode';
export { SideNavService } from 'services/side-nav';
export { VideoSettingsService } from 'services/settings-v2/video';
export { SettingsManagerService } from 'services/settings-manager';
export { MarkersService } from 'services/markers';
export { MybilibiliLiveService } from 'services/mybilibili';
export { RealmService } from 'services/realm';
export { OnboardingV2Service } from 'services/onboarding/onboarding-v2';

// ONLINE SERVICES
export { UserService } from './services/user';
export { UsageStatisticsService } from './services/usage-statistics';
export { VideoEncodingOptimizationService } from 'services/video-encoding-optimizations';
export { StreamingService } from 'services/streaming';