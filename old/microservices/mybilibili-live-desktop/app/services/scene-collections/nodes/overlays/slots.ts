import { Scene, TSceneNode } from 'services/scenes';
import { TDisplayType } from 'services/settings-v2';
import { ArrayNode } from '../array-node';

interface IItemSchema {
  id: string;
  name: string;
  sceneNodeType: 'item';
  visible?: boolean;
  display?: TDisplayType;
  locked: boolean;
}

export interface IFolderSchema {
  id: string;
  name: string;
  sceneNodeType: 'folder';
  childrenIds: string[];
  display?: TDisplayType;
}

export type TSlotSchema = IItemSchema | IFolderSchema;

interface IContext {
  scene: Scene;
}

export class SlotsNode extends ArrayNode<TSlotSchema, IContext, TSceneNode> {
  schemaVersion = 2;

  getItems(context: IContext): TSceneNode[] {
    return context.scene.getRootNodes();
  }

  async saveItem(item: TSceneNode): Promise<TSlotSchema> {
    if (item.isFolder()) {
      return {
        id: item.id,
        name: item.name,
        sceneNodeType: 'folder',
        childrenIds: item.getNestedItems().map(child => child.id),
      };
    }

    return {
      id: item.id,
      name: item.name,
      sceneNodeType: 'item',
      visible: item.visible,
      locked: item.locked,
    };
  }

  async loadItem(_obj: TSlotSchema): Promise<void> {}
}
