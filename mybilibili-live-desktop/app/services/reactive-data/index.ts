import { Service } from 'services';

export type TStateTreeLeaf = number | string;
export type TStateFlat = { [x: `${string}.${string}`]: TStateTreeLeaf };
export type TSchemaFlat = { [x: `${string}.${string}`]: { name: string; aliases?: string[] } };

export class ReactiveDataService extends Service {
  public sourceStateKeyInterest: Map<string, Set<string>> = new Map();

  async updateState(_changes: Partial<Record<string, number>>) {}

  async getUserState(_keys: string[]) {
    return undefined;
  }

  watchSourceState(_sourceId: string, _keys: string[]) {}

  unwatchSourceState(sourceId: string) {
    this.sourceStateKeyInterest.delete(sourceId);
  }
}
