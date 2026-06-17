import type { Account, AccountsClassNames } from './types';
import { useAccountsConfig } from './context';
import { mergeAccountClassNames } from './defaultClassNames';
import { accountFullName, accountInitials } from './names';

type AccountAvatarProps = {
  account: Account;
  className?: string;
  /** Header trigger: rounded square with initials (matches sign-in trigger shape). */
  variant?: 'default' | 'trigger';
  classNames?: AccountsClassNames;
};

export function AccountAvatar({
  account,
  className,
  variant = 'default',
  classNames: classNamesProp,
}: AccountAvatarProps) {
  const { classNames: ctxClassNames } = useAccountsConfig();
  const classNames = mergeAccountClassNames(ctxClassNames, classNamesProp);
  const rootClass = [classNames.avatar, className].filter(Boolean).join(' ');
  const label = accountFullName(account) || account.email;

  if (variant === 'trigger' || !account.avatarUrl) {
    return (
      <span className={rootClass} aria-hidden="true" title={label}>
        <span className={classNames.avatarFallback}>{accountInitials(account)}</span>
      </span>
    );
  }

  return (
    <span className={rootClass} aria-hidden="true" title={label}>
      <img
        src={account.avatarUrl}
        alt=""
        className={classNames.avatarImage}
        referrerPolicy="no-referrer"
      />
    </span>
  );
}
