import type { AccountUser, AccountsClassNames } from './types';
import { useAccountsConfig } from './context';
import { mergeAccountClassNames } from './defaultClassNames';
import { accountUserFullName, accountUserInitials } from './names';

type AccountAvatarProps = {
  user: AccountUser;
  className?: string;
  /** Header trigger: rounded square with initials (matches sign-in trigger shape). */
  variant?: 'default' | 'trigger';
  classNames?: AccountsClassNames;
};

export function AccountAvatar({
  user,
  className,
  variant = 'default',
  classNames: classNamesProp,
}: AccountAvatarProps) {
  const { classNames: ctxClassNames } = useAccountsConfig();
  const classNames = mergeAccountClassNames(ctxClassNames, classNamesProp);
  const rootClass = [classNames.avatar, className].filter(Boolean).join(' ');
  const label = accountUserFullName(user) || user.email;

  if (variant === 'trigger' || !user.avatarUrl) {
    return (
      <span className={rootClass} aria-hidden="true" title={label}>
        <span className={classNames.avatarFallback}>{accountUserInitials(user)}</span>
      </span>
    );
  }

  return (
    <span className={rootClass} aria-hidden="true" title={label}>
      <img
        src={user.avatarUrl}
        alt=""
        className={classNames.avatarImage}
        referrerPolicy="no-referrer"
      />
    </span>
  );
}
