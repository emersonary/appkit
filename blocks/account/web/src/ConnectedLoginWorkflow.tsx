import { LoginWorkflow, type LoginWorkflowProps } from './LoginWorkflow';
import { useLoginWorkflowHandlers, type UseLoginWorkflowHandlersOptions } from './useLoginWorkflowHandlers';

export type ConnectedLoginWorkflowProps = Omit<
  LoginWorkflowProps,
  | 'onRequestReset'
  | 'onResetPassword'
  | 'onFetchVerificationStatus'
  | 'onResendVerificationEmail'
  | 'onGoogleClick'
  | 'onLogin'
  | 'onRegister'
  | 'showGoogle'
> & {
  apiBaseUrl?: string;
  onAccountError?: UseLoginWorkflowHandlersOptions['onAccountError'];
  mapResetError?: UseLoginWorkflowHandlersOptions['mapResetError'];
  onLogin?: LoginWorkflowProps['onLogin'];
  onRegister?: LoginWorkflowProps['onRegister'];
};

export function ConnectedLoginWorkflow({
  apiBaseUrl = '',
  onAccountError,
  mapResetError,
  classifyResetError,
  onLogin: onLoginOverride,
  onRegister: onRegisterOverride,
  isLoading: isLoadingOverride,
  ...workflowProps
}: ConnectedLoginWorkflowProps) {
  const handlers = useLoginWorkflowHandlers({
    apiBaseUrl,
    onAccountError,
    mapResetError: classifyResetError ?? mapResetError,
  });

  return (
    <LoginWorkflow
      {...workflowProps}
      isLoading={isLoadingOverride ?? handlers.isLoading}
      classifyResetError={classifyResetError ?? handlers.classifyResetError}
      onLogin={onLoginOverride ?? handlers.onLogin}
      onRegister={
        onRegisterOverride ??
        (async (email, password, firstName, lastName) => {
          await handlers.onRegister(email, password, firstName, lastName);
        })
      }
      onRequestReset={handlers.onRequestReset}
      onResetPassword={handlers.onResetPassword}
      onFetchVerificationStatus={handlers.onFetchVerificationStatus}
      onResendVerificationEmail={handlers.onResendVerificationEmail}
      onGoogleClick={handlers.onGoogleClick}
      showGoogle={handlers.showGoogle && workflowProps.step === 'sign-in'}
    />
  );
}
