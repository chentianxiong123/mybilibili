import { ObjectSchema } from 'realm';
import { Service } from 'services/core';
import { RealmObject } from 'services/realm';

export interface IConnectedDevice {
  socketId: string;
  deviceName: string;
  clientType: string;
}

class RemoteControlEphemeralState extends RealmObject {
  devices: IConnectedDevice[];

  static schema: ObjectSchema = {
    name: 'RemoteControlEphemeralState',
    properties: {
      devices: {
        type: 'list',
        objectType: 'string',
        default: [] as string[],
      },
    },
  };
}

RemoteControlEphemeralState.register();

class RemoteControlPersistentState extends RealmObject {
  enabled: boolean;

  static schema: ObjectSchema = {
    name: 'RemoteControlPersistentState',
    properties: {
      enabled: { type: 'bool', default: false },
    },
  };
}

RemoteControlPersistentState.register({ persist: true });

export class RemoteControlService extends Service {
  state = RemoteControlPersistentState.inject();
  connectedDevices = RemoteControlEphemeralState.inject();

  disconnect() {
    this.setEnableRemoteConnection(false);
    this.setConnectedDevices([]);
  }

  disconnectDevice(socketId: string) {
    this.removeConnectedDevice(socketId);
  }

  async createRemoteConnection() {
    this.setEnableRemoteConnection(false);
  }

  sendMessage(_event: unknown) {}

  setEnableRemoteConnection(val: boolean) {
    this.state.db.write(() => {
      this.state.enabled = val;
    });
  }

  setConnectedDevices(_devices: IConnectedDevice[]) {
    this.connectedDevices.db.write(() => {
      this.connectedDevices.devices = [];
    });
  }

  removeConnectedDevice(_socketId: string) {
    this.setConnectedDevices([]);
  }
}
