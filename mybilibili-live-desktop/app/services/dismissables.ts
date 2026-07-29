import { PersistentStatefulService } from 'services/core/persistent-stateful-service';
import { mutation, ViewHandler } from './core/stateful-service';
import Vue from 'vue';

export enum EDismissable {
  SceneCollectionsHelpTip = 'scene_collections_help_tip',
  HighlighterNotification = 'highlighter_notification',
  SourceSelectorFolders = 'source_selector_folders',
  NewSideNav = 'new_side_nav',
  CustomMenuSettings = 'custom_menu_settings',
  LoginPrompt = 'login_prompt',
  EnhancedBroadcasting = 'enhanced_broadcasting',
}

interface IDismissablesServiceState {
  [key: string]: boolean;
}

class DismissablesViews extends ViewHandler<IDismissablesServiceState> {
  shouldShow(key: EDismissable) {
    return !this.state[key];
  }
}

/**
 * A dismissable is anything that can be dismissed and should
 * never show up again, like a help tip.
 */
export class DismissablesService extends PersistentStatefulService<IDismissablesServiceState> {
  initialize() {}

  get views() {
    return new DismissablesViews(this.state);
  }

  dismiss(key: EDismissable) {
    this.DISMISS(key);
  }

  dismissAll() {
    Object.keys(EDismissable).forEach((key: keyof typeof EDismissable) =>
      this.dismiss(EDismissable[key]),
    );
  }

  /**
   * Resets all dismissables. Useful for testing.
   * @deprecated For testing use only
   */
  reset() {
    this.RESET();
  }

  @mutation()
  DISMISS(key: EDismissable) {
    Vue.set(this.state, key, true);
  }

  @mutation()
  RESET() {
    this.state = {};
  }
}
