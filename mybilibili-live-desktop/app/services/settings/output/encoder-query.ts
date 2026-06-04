import { Inject } from 'services/core/injector';
import { Service } from 'services/core/service';
import { IObsListOption } from 'components/obs/inputs/ObsInput';
import {
  SimpleStreamingFactory,
  AdvancedStreamingFactory,
  SimpleRecordingFactory,
  AdvancedRecordingFactory,
  ERecordingFormat,
} from '../../../../obs-api';
import { StreamingService } from 'services/streaming';

/**
 * The module.ts in obs-studio-node is out of sync with module.d.ts and is
 * missing getAvailableEncoders and IEncoderOption.  TypeScript resolves
 * .ts over .d.ts, so we declare the shapes we need locally.
 */
interface IEncoderOption {
  title: string;
  name: string;
}

interface IWithAvailableEncoders {
  getAvailableEncoders(): IEncoderOption[];
}

function mapEncoders(encoders: IEncoderOption[]): IObsListOption<string>[] {
  return encoders.map(e => ({ description: e.title, value: e.name }));
}

function hasGetAvailableEncoders(instance: any): instance is IWithAvailableEncoders {
  return typeof instance?.getAvailableEncoders === 'function';
}

interface ICacheEntry {
  key: string;
  value: IObsListOption<string>[];
}

export class EncoderQueryService extends Service {
  @Inject() private streamingService: StreamingService;

  private streamingEncoderCache: ICacheEntry | null = null;
  private recordingEncoderCache: ICacheEntry | null = null;

  getAvailableStreamingEncoders(
    mode: 'Simple' | 'Advanced',
    streamSettings?: Record<string, any>,
  ): IObsListOption<string>[] {
    const cacheKey = `${mode}:custom-rtmp`;

    if (this.streamingEncoderCache?.key === cacheKey) {
      return this.streamingEncoderCache.value;
    }

    try {
      const existing = this.streamingService.getStreamingInstance();
      if (existing && hasGetAvailableEncoders(existing)) {
        const result = mapEncoders(existing.getAvailableEncoders());
        this.streamingEncoderCache = { key: cacheKey, value: result };
        return result;
      }

      if (mode === 'Simple') {
        const instance = SimpleStreamingFactory.create();
        let service: any = null;
        try {
          if (!hasGetAvailableEncoders(instance)) return [];
          const result = mapEncoders(instance.getAvailableEncoders());
          this.streamingEncoderCache = { key: cacheKey, value: result };
          return result;
        } finally {
          SimpleStreamingFactory.destroy(instance);
        }
      } else {
        const instance = AdvancedStreamingFactory.create();
        let service: any = null;
        try {
          if (!hasGetAvailableEncoders(instance)) return [];
          const result = mapEncoders(instance.getAvailableEncoders());
          this.streamingEncoderCache = { key: cacheKey, value: result };
          return result;
        } finally {
          AdvancedStreamingFactory.destroy(instance);
        }
      }
    } catch (e: unknown) {
      console.error('Error querying available streaming encoders', e);
      return [];
    }
  }

  getAvailableRecordingEncoders(
    mode: 'Simple' | 'Advanced',
    format: ERecordingFormat,
  ): IObsListOption<string>[] {
    const cacheKey = `${mode}:${format}`;

    if (this.recordingEncoderCache?.key === cacheKey) {
      return this.recordingEncoderCache.value;
    }

    try {
      const existing = this.streamingService.getRecordingInstance();
      if (existing && hasGetAvailableEncoders(existing)) {
        const result = mapEncoders(existing.getAvailableEncoders());
        this.recordingEncoderCache = { key: cacheKey, value: result };
        return result;
      }

      if (mode === 'Simple') {
        const instance = SimpleRecordingFactory.create();
        try {
          instance.format = format;
          if (!hasGetAvailableEncoders(instance)) return [];
          const result = mapEncoders(instance.getAvailableEncoders());
          this.recordingEncoderCache = { key: cacheKey, value: result };
          return result;
        } finally {
          SimpleRecordingFactory.destroy(instance);
        }
      } else {
        const instance = AdvancedRecordingFactory.create();
        try {
          instance.format = format;
          if (!hasGetAvailableEncoders(instance)) return [];
          const result = mapEncoders(instance.getAvailableEncoders());
          this.recordingEncoderCache = { key: cacheKey, value: result };
          return result;
        } finally {
          AdvancedRecordingFactory.destroy(instance);
        }
      }
    } catch (e: unknown) {
      console.error('Error querying available recording encoders', e);
      return [];
    }
  }

}
