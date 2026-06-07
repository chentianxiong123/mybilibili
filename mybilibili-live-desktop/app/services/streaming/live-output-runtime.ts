import { Inject, Service } from 'services/core';
import { SettingsService } from 'services/settings';
import { OutputSettingsService } from 'services/settings/output';
import { VideoSettingsService } from 'services/settings-v2/video';
import {
  AdvancedStreamingFactory,
  ServiceFactory,
  SimpleStreamingFactory,
} from '../../../obs-api';

type TStreamingMode = 'Simple' | 'Advanced';

interface IObsStreamingInstance {
  service?: unknown;
  video?: unknown;
  start(): void | Promise<void>;
  stop(): void | Promise<void>;
}

interface IStreamingFactory {
  create(): IObsStreamingInstance;
  destroy(instance: IObsStreamingInstance): void;
}

interface IObsServiceFactory {
  create(type: string, name: string, settings: Record<string, unknown>): unknown;
  destroy?(service: unknown): void;
}

interface IStreamRuntimeConfig {
  streamType: string;
  server: string;
  key: string;
}

export class LiveOutputRuntimeService extends Service {
  @Inject() private settingsService: SettingsService;
  @Inject() private outputSettingsService: OutputSettingsService;
  @Inject() private videoSettingsService: VideoSettingsService;

  private streamingInstance: IObsStreamingInstance | null = null;
  private streamingMode: TStreamingMode | null = null;
  private streamService: unknown = null;

  get instance() {
    return this.streamingInstance;
  }

  get isRunning() {
    return this.streamingInstance !== null;
  }

  async start() {
    if (this.streamingInstance) {
      throw new Error('OBS streaming output is already running');
    }

    const mode = this.outputSettingsService.getSettings().mode;
    const factory = this.getFactory(mode);
    const instance = factory.create();

    try {
      this.configureStreamingInstance(instance, mode);
      await Promise.resolve(instance.start());
      this.streamingInstance = instance;
      this.streamingMode = mode;
    } catch (error: unknown) {
      this.destroyInstance(factory, instance);
      this.destroyStreamService();
      throw error;
    }
  }

  async stop() {
    if (!this.streamingInstance || !this.streamingMode) {
      throw new Error('OBS streaming output is not running');
    }

    const instance = this.streamingInstance;
    const factory = this.getFactory(this.streamingMode);
    try {
      await Promise.resolve(instance.stop());
    } finally {
      this.destroyInstance(factory, instance);
      this.destroyStreamService();
      this.streamingInstance = null;
      this.streamingMode = null;
    }
  }

  private configureStreamingInstance(instance: IObsStreamingInstance, mode: TStreamingMode) {
    const context = this.videoSettingsService.contexts.horizontal;
    if (!context) {
      throw new Error('OBS horizontal video context is not ready');
    }

    const stream = this.getStreamRuntimeConfig();
    const outputSettings = this.outputSettingsService.getStreamingSettings('horizontal');
    const serviceFactory = ServiceFactory as unknown as IObsServiceFactory;

    this.streamService = serviceFactory.create(stream.streamType, 'mybilibili Live', {
      server: stream.server,
      key: stream.key,
    });

    instance.video = context;
    instance.service = this.streamService;
    Object.assign(instance, outputSettings);

    if (mode !== this.outputSettingsService.getSettings().mode) {
      throw new Error('OBS output mode changed while preparing streaming output');
    }
  }

  private getStreamRuntimeConfig(): IStreamRuntimeConfig {
    const streamValues = this.settingsService.views.values.Stream;
    const streamType = String(streamValues.streamType || '');
    const server = String(streamValues.server || '');
    const key = String(streamValues.key || '');

    if (!streamType) throw new Error('OBS stream type is not configured');
    if (!server) throw new Error('OBS stream server is not configured');
    if (!key) throw new Error('OBS stream key is not configured');

    return { streamType, server, key };
  }

  private getFactory(mode: TStreamingMode): IStreamingFactory {
    if (mode === 'Advanced') return AdvancedStreamingFactory as unknown as IStreamingFactory;
    if (mode === 'Simple') return SimpleStreamingFactory as unknown as IStreamingFactory;
    throw new Error(`Unsupported OBS output mode: ${mode}`);
  }

  private destroyInstance(factory: IStreamingFactory, instance: IObsStreamingInstance) {
    try {
      factory.destroy(instance);
    } catch (error: unknown) {
      console.error('Failed to destroy OBS streaming output', error);
    }
  }

  private destroyStreamService() {
    if (!this.streamService) return;

    const serviceFactory = ServiceFactory as unknown as IObsServiceFactory;
    if (typeof serviceFactory.destroy === 'function') {
      try {
        serviceFactory.destroy(this.streamService);
      } catch (error: unknown) {
        console.error('Failed to destroy OBS stream service', error);
      }
    }
    this.streamService = null;
  }
}
