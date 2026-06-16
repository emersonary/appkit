import type { AccountsClassNames } from './types';

export const defaultAccountClassNames: Required<
  Pick<
    AccountsClassNames,
    | 'trigger'
    | 'signInTrigger'
    | 'signInIcon'
    | 'avatar'
    | 'avatarImage'
    | 'avatarFallback'
    | 'panel'
    | 'panelHeader'
    | 'panelEmail'
    | 'menuItem'
  >
> = {
  trigger: 'appkit-account-menu__trigger',
  signInTrigger: 'appkit-account-menu__sign-in-trigger',
  signInIcon: 'appkit-account-menu__sign-in-icon',
  avatar: 'appkit-account-menu__avatar',
  avatarImage: 'appkit-account-menu__avatar-img',
  avatarFallback: 'appkit-account-menu__avatar-initials',
  panel: 'appkit-account-menu__panel',
  panelHeader: 'appkit-account-menu__panel-header',
  panelEmail: 'appkit-account-menu__panel-email',
  menuItem: 'appkit-account-menu__menu-item',
};

export function mergeAccountClassNames(...sources: Array<AccountsClassNames | undefined>): AccountsClassNames {
  return Object.assign({}, defaultAccountClassNames, ...sources);
}
