import { Subject } from 'rxjs';
import { Service } from './core/service';

export type TSocketEvent = { type: string; message?: unknown; platform?: string };

export class WebsocketService extends Service {
  socketEvent = new Subject<TSocketEvent>();

  connect() {}

  disconnect() {}
}
