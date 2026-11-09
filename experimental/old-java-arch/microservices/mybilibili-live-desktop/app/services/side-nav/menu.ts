import { PersistentStatefulService, ViewHandler, mutation } from 'services/core';
import { EMenuItemKey, ENavName, IAppMenuItem, IMenu, SideBarBottomNavData, SideBarTopNavData } from './menu-data';

interface ISideNavServiceState {
  version: string;
  isOpen: boolean;
  showCustomEditor: boolean;
  hasLegacyMenu: boolean;
  compactView: boolean;
  currentMenuItem: EMenuItemKey | string;
  apps: IAppMenuItem[];
  [ENavName.TopNav]: IMenu;
  [ENavName.BottomNav]: IMenu;
}

class SideNavViews extends ViewHandler<ISideNavServiceState> {
  get isOpen() {
    return this.state.isOpen;
  }

  get compactView() {
    return this.state.compactView;
  }

  get hasLegacyMenu() {
    return this.state.hasLegacyMenu;
  }

  get currentMenuItem() {
    return this.state.currentMenuItem;
  }

  get apps() {
    return [];
  }

  get showCustomEditor() {
    return this.state.showCustomEditor;
  }
}

export class SideNavService extends PersistentStatefulService<ISideNavServiceState> {
  static defaultState: ISideNavServiceState = {
    version: 'mybilibili-local',
    isOpen: false,
    showCustomEditor: true,
    hasLegacyMenu: false,
    compactView: false,
    currentMenuItem: EMenuItemKey.Editor,
    apps: [],
    [ENavName.TopNav]: SideBarTopNavData(),
    [ENavName.BottomNav]: SideBarBottomNavData(),
  };

  get views() {
    return new SideNavViews(this.state);
  }

  toggleMenuStatus() {
    this.OPEN_CLOSE_MENU();
  }

  setCurrentMenuItem(key: EMenuItemKey | string) {
    this.SET_CURRENT_MENU_ITEM(key);
  }

  setCompactView(isCompact: boolean) {
    this.SET_COMPACT_VIEW(isCompact);
  }

  expandMenuItem() {}

  toggleSidebarSubmenu() {}

  setMenuItemStatus() {}

  toggleMenuItem() {}

  updateAllApps() {}

  toggleApp() {}

  replaceApp() {}

  @mutation()
  private OPEN_CLOSE_MENU() {
    this.state.isOpen = !this.state.isOpen;
  }

  @mutation()
  private SET_CURRENT_MENU_ITEM(key: EMenuItemKey | string) {
    this.state.currentMenuItem = key;
  }

  @mutation()
  private SET_COMPACT_VIEW(isCompact: boolean) {
    this.state.compactView = isCompact;
  }
}
