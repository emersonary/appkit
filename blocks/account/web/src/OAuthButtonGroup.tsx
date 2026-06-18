import { useAccountsConfig } from './context';
import { GoogleIcon } from './fieldIcons';
import type { AccountsClassNames, AccountsUILabels } from './types';

type OAuthButtonGroupProps = {
  onGoogleClick?: () => void;
  googleOAuthUrl?: string;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
};

export function OAuthButtonGroup({
  onGoogleClick,
  googleOAuthUrl: googleOAuthUrlProp,
  labels: labelsProp,
  classNames: classNamesProp,
}: OAuthButtonGroupProps) {
  const ctx = useAccountsConfig();
  const googleOAuthUrl = googleOAuthUrlProp ?? ctx.googleOAuthUrl;
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };
  const oauthEnabled = ctx.oauth.enabled !== false;
  const googleEnabled = ctx.oauth.google !== false;
  const canUseGoogle =
    ctx.registrationEnabled && oauthEnabled && googleEnabled && Boolean(googleOAuthUrl || onGoogleClick);
  if (!canUseGoogle) return null;

  function handleGoogle() {
    if (onGoogleClick) {
      onGoogleClick();
      return;
    }
    if (googleOAuthUrl) {
      window.location.href = googleOAuthUrl;
    }
  }

  return (
    <div className={classNames.oauthGroup}>
      <button type="button" className={classNames.oauthButton} onClick={handleGoogle}>
        <GoogleIcon className="appkit-login-workflow__oauth-icon" />
        {labels.continueWithGoogle}
      </button>
    </div>
  );
}
