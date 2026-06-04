export type TStreamErrorType =
  | 'LOGGED_OUT_ERROR'
  | 'SETTINGS_UPDATE_FAILED'
  | 'INVALID_ENCODER'
  | 'OUTPUT_START_FAILED'
  | 'UNKNOWN';

export interface IStreamError {
  type: TStreamErrorType;
  message: string;
  details?: unknown;
  platform?: string;
}

export class StreamError extends Error implements IStreamError {
  constructor(
    public type: TStreamErrorType = 'UNKNOWN',
    rejectedRequest?: { message?: string },
    public details?: unknown,
  ) {
    super(rejectedRequest?.message || type);
  }
}

export function createStreamError(
  type: TStreamErrorType = 'UNKNOWN',
  rejectedRequest?: { message?: string },
  details?: unknown,
): StreamError {
  return new StreamError(type, rejectedRequest, details);
}

export function throwStreamError(
  type: TStreamErrorType = 'UNKNOWN',
  rejectedRequest?: { message?: string },
  details?: unknown,
): never {
  throw createStreamError(type, rejectedRequest, details);
}

export function formatStreamErrorMessage(error?: TStreamErrorType | StreamError) {
  const type = error instanceof StreamError ? error.type : error || 'UNKNOWN';
  return {
    message: type,
    report: type,
  };
}

export function formatUnknownErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '未知推流错误';
}
