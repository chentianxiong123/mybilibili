import { $t } from 'services/i18n';
import { TAppPage } from 'services/navigation';

/**
 * Update Menu Items
 * 1. EMenuItemKey: Add/update/remove enum in EMenuItemKey.
 * 2. Add string title to menuTitles.
 * 3. Add an entry to SideNavMenuItems. Use the subMenuItems property to add submenu items from SideBarSubMenuItems to the menu item.
 * 4. To show the menu item, add it to either SideBarTopNavData or SideBarBottomNavData.
 * 5. To show a menu item in the top nav to a logged out user, add it to loggedOutMenuItems.
 * 6. To show a menu item in the top nav compact menu, add it to compactMenuItemKeys.
 */
export enum EMenuItemKey {
  Editor = 'editor',
  LayoutEditor = 'layout-editor',
  StudioMode = 'studio-mode',
  RecordingHistory = 'recording-history',
  DevTools = 'dev-tools',
  Settings = 'settings',
  Login = 'login',
}

/**
 * Update SubMenu Items
 * 1. ESubMenuItemKey: Add/update/remove enum.
 * 2. Add string title to menuTitles.
 * 3. Add entry to SideBarSubMenuItems.
 * 4. To show the submenu item, add it to a menu item entry in SideNavMenuItems by using the subMenuItems property.
 */
export const ESubMenuItemKey = {};
export type ESubMenuItemKey = never;

export const ESideNavKey = { ...EMenuItemKey, ...ESubMenuItemKey };

/**
 * Update External Links
 * 1. Confirm external link parameter for url.
 * 2. Add/update/remove type in TExternalLinkType. The type is the url parameter.
 */
export type TExternalLinkType =
  never;

/**
 * Add custom side item targets here. This is for menu items that don't fit
 * the standard target/type model.
 */
export type TCustomSideItem = never;

/**
 * Update Protocal Link Map
 * 1. Confirm protocol link parameter for url.
 * 2. Add/update/remove entry in ProtocolLinkKeyMap. The ket is the url parameter.
 */
export const ProtocolLinkKeyMap = {};

type TSideNavItem = TAppPage | TExternalLinkType | 'NavTools' | TCustomSideItem;
export interface IAppMenuItem {
  id: string;
  name?: string;
  isActive: boolean;
  icon?: string;
}
export interface IMenu {
  name: string;
  menuItems: (IMenuItem | IParentMenuItem)[];
}

interface ISideNavItem {
  key: EMenuItemKey | ESubMenuItemKey;
  target?: TSideNavItem; // optional because menu item could be a toggle
}
export interface IMenuItem extends ISideNavItem {
  type?: TExternalLinkType | string;
  trackingTarget?: string;
  icon?: string;
  isExpanded: boolean;
  isActive?: boolean;
}

export interface IParentMenuItem extends IMenuItem {
  isToggled?: boolean;
  subMenuItems: IMenuItem[];
}

export enum ENavName {
  TopNav = 'top-nav',
  BottomNav = 'bottom-nav',
}

export const loggedOutMenuItems: ISideNavItem[] = [
  {
    key: EMenuItemKey.Editor,
    target: 'Studio',
  },
];

export const compactMenuItemKeys: Set<EMenuItemKey | ESubMenuItemKey> = new Set([
  EMenuItemKey.Editor,
]);

/**
 * The string titles for the menu items and submenu items
 * @param item - key for the menu item
 * @returns string title
 */
export const menuTitles = (item: EMenuItemKey | ESubMenuItemKey | string) => {
  return {
    [EMenuItemKey.Editor]: $t('Editor'),
    [EMenuItemKey.LayoutEditor]: $t('Layout Editor'),
    [EMenuItemKey.StudioMode]: $t('Studio Mode'),
    [EMenuItemKey.RecordingHistory]: $t('Recordings'),
    [EMenuItemKey.DevTools]: 'Dev Tools',
    [EMenuItemKey.Settings]: $t('Settings'),
    [EMenuItemKey.Login]: $t('Login'),
  }[item];
};

/**
 * Menu items in the top menu of the side nav
 */
export const SideBarTopNavData = () => {
  const menuItems = SideNavMenuItems();
  return {
    name: ENavName.TopNav,
    menuItems: [
      menuItems[EMenuItemKey.Editor],
      menuItems[EMenuItemKey.LayoutEditor],
      menuItems[EMenuItemKey.StudioMode],
    ],
  };
};

/**
 * Menu items in the bottom menu of the side nav
 */
export const SideBarBottomNavData = (): IMenu => {
  const menuItems = SideNavMenuItems();
  return {
    name: ENavName.BottomNav,
    menuItems: [
      menuItems[EMenuItemKey.DevTools],
      menuItems[EMenuItemKey.Settings],
    ],
  };
};

export type TMenuItems = {
  [MenuItem in Partial<EMenuItemKey>]: IMenuItem | IParentMenuItem;
};

/**
 * Data for menu items in the side nav
 */
export const SideNavMenuItems = (): TMenuItems => {
  return {
    [EMenuItemKey.Editor]: {
      key: EMenuItemKey.Editor,
      target: 'Studio',
      trackingTarget: 'editor',
      icon: 'icon-studio',
      isActive: true,
      isExpanded: false,
    },
    [EMenuItemKey.LayoutEditor]: {
      key: EMenuItemKey.LayoutEditor,
      target: 'LayoutEditor',
      trackingTarget: 'layout-editor',
      icon: 'fas fa-th-large',
      isActive: true,
      isExpanded: false,
    },
    [EMenuItemKey.StudioMode]: {
      key: EMenuItemKey.StudioMode,
      icon: 'icon-studio-mode-3',
      isActive: true,
      isExpanded: false,
    },
    [EMenuItemKey.RecordingHistory]: {
      key: EMenuItemKey.RecordingHistory,
      target: 'RecordingHistory',
      icon: 'icon-play-round',
      trackingTarget: 'recording-history',
      isActive: false,
      isExpanded: false,
    },
    [EMenuItemKey.DevTools]: {
      key: EMenuItemKey.DevTools,
      trackingTarget: 'devtools',
      icon: 'icon-developer',
      isExpanded: false,
    },
    [EMenuItemKey.Settings]: {
      key: EMenuItemKey.Settings,
      icon: 'icon-settings',
      isActive: true,
      isExpanded: false,
    },
    [EMenuItemKey.Login]: {
      key: EMenuItemKey.Login,
      icon: 'icon-user',
      isActive: true,
      isExpanded: false,
    },
  };
};

type TSubMenuItems = {
  [MenuItem in ESubMenuItemKey]?: IMenuItem | IParentMenuItem;
};

/**
 * Data for sub menu items in the side nav
 */
export const SideBarSubMenuItems = (): TSubMenuItems => ({});
