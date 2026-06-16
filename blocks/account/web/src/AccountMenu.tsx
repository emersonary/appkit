import { useContext, useEffect, useId, useRef, useState, type ReactNode } from 'react';
import { AccountAvatar } from './AccountAvatar';
import { AccountSignInIcon } from './AccountSignInIcon';
import { mergeAccountClassNames } from './defaultClassNames';
import { useAccountsConfig } from './context';
import { accountUserFullName } from './names';
import { AccountsSessionCtx } from './AccountsProvider';
import type { AccountUser, AccountsClassNames, AccountsUILabels } from './types';

function useOptionalAccountsSession() {
  return useContext(AccountsSessionCtx);
}

export type AccountMenuProps = {
  user?: AccountUser | null;
  isAuthenticated?: boolean;
  onLogout?: () => void | Promise<void>;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  renderPanelExtra?: (user: AccountUser, close: () => void) => ReactNode;
  signInHref?: string;
  onSignInClick?: () => void;
  renderSignInTrigger?: (className?: string) => ReactNode;
  settingsHref?: string;
  onSettingsClick?: () => void;
  hidePanelHeader?: boolean;
};

export function AccountMenu({
  user: userProp,
  isAuthenticated: isAuthenticatedProp,
  onLogout: onLogoutProp,
  labels: labelsProp,
  classNames: classNamesProp,
  renderPanelExtra,
  signInHref,
  onSignInClick,
  renderSignInTrigger,
  settingsHref: settingsHrefProp,
  onSettingsClick: onSettingsClickProp,
  hidePanelHeader = false,
}: AccountMenuProps) {
  const sessionCtx = useOptionalAccountsSession();
  const ctx = useAccountsConfig();
  const user = userProp ?? sessionCtx?.user ?? null;
  const isAuthenticated = isAuthenticatedProp ?? sessionCtx?.isAuthenticated ?? !!user;
  const logout = onLogoutProp ?? sessionCtx?.logout;
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = mergeAccountClassNames(ctx.classNames, classNamesProp);
  const settingsHref = settingsHrefProp ?? ctx.settingsHref;
  const onSettingsClick = onSettingsClickProp ?? ctx.onSettingsClick;
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useEffect(() => {
    if (!isAuthenticated) setOpen(false);
  }, [isAuthenticated]);

  useEffect(() => {
    if (!open) return;
    function handlePointerDown(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false);
    }
    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [open]);

  if (!isAuthenticated || !user) {
    const signInClass = classNames.signInTrigger ?? classNames.trigger;
    if (renderSignInTrigger) {
      return <div className="appkit-account-menu">{renderSignInTrigger(signInClass)}</div>;
    }
    const icon = <AccountSignInIcon className={classNames.signInIcon} />;
    if (onSignInClick) {
      return (
        <div className="appkit-account-menu">
          <button
            type="button"
            className={signInClass}
            onClick={onSignInClick}
            title={labels.signIn}
            aria-label={labels.signIn}
          >
            {icon}
          </button>
        </div>
      );
    }
    if (signInHref) {
      return (
        <div className="appkit-account-menu">
          <a
            className={signInClass}
            href={signInHref}
            title={labels.signIn}
            aria-label={labels.signIn}
          >
            {icon}
          </a>
        </div>
      );
    }
    return null;
  }

  const name = accountUserFullName(user);

  return (
    <div className="appkit-account-menu" ref={rootRef}>
      <button
        type="button"
        className={classNames.trigger}
        aria-expanded={open}
        aria-haspopup="true"
        aria-controls={menuId}
        title={labels.accountMenu}
        onClick={() => setOpen((v) => !v)}
      >
        <AccountAvatar user={user} variant="trigger" classNames={classNames} />
      </button>
      {open && (
        <div
          id={menuId}
          className={classNames.panel}
          role="dialog"
          aria-label={labels.accountMenu}
        >
          {!hidePanelHeader && (
            <div className={classNames.panelHeader}>
              <AccountAvatar user={user} classNames={classNames} />
              <div>
                {name && <div>{name}</div>}
                <div className={classNames.panelEmail}>{user.email}</div>
              </div>
            </div>
          )}
          {settingsHref && (
            <a className={classNames.menuItem} href={settingsHref} onClick={() => setOpen(false)}>
              Settings
            </a>
          )}
          {onSettingsClick && (
            <button
              type="button"
              className={classNames.menuItem}
              onClick={() => {
                onSettingsClick();
                setOpen(false);
              }}
            >
              Settings
            </button>
          )}
          {renderPanelExtra?.(user, () => setOpen(false))}
          <button
            type="button"
            className={classNames.menuItem}
            onClick={() => {
              setOpen(false);
              void logout?.();
            }}
          >
            {labels.signOut}
          </button>
        </div>
      )}
    </div>
  );
}

export function SignInButton({ href, onClick }: { href?: string; onClick?: () => void }) {
  const { labels, classNames } = useAccountsConfig();
  if (onClick) {
    return (
      <button type="button" className={classNames.trigger} onClick={onClick}>
        {labels.signIn}
      </button>
    );
  }
  if (href) {
    return (
      <a className={classNames.trigger} href={href}>
        {labels.signIn}
      </a>
    );
  }
  return null;
}

export function SignOutButton() {
  const { labels, classNames } = useAccountsConfig();
  const session = useOptionalAccountsSession();
  if (!session) return null;
  return (
    <button type="button" className={classNames.menuItem} onClick={() => session.logout()}>
      {labels.signOut}
    </button>
  );
}
