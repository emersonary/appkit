import { Code, ConnectError } from '@connectrpc/connect';
import { AccountsError, AccountsErrorCode } from './errors';

export function fromConnectError(err: unknown): AccountsError {
  if (err instanceof ConnectError) {
    const message = err.message.toLowerCase();
    if (message.includes('unmarshal') && message.includes('loginrequest')) {
      return new AccountsError(AccountsErrorCode.UNAVAILABLE);
    }

    switch (err.code) {
      case Code.InvalidArgument:
        return new AccountsError(AccountsErrorCode.INVALID_ARGUMENT);
      case Code.Unauthenticated:
        return new AccountsError(AccountsErrorCode.UNAUTHENTICATED);
      case Code.PermissionDenied:
        return new AccountsError(AccountsErrorCode.PERMISSION_DENIED);
      case Code.AlreadyExists:
        return new AccountsError(AccountsErrorCode.ALREADY_EXISTS);
      case Code.NotFound:
        return new AccountsError(AccountsErrorCode.NOT_FOUND);
      case Code.Unavailable:
        return new AccountsError(AccountsErrorCode.UNAVAILABLE);
      case Code.FailedPrecondition:
        return new AccountsError(AccountsErrorCode.EMAIL_NOT_VERIFIED);
      default:
        return new AccountsError(AccountsErrorCode.INTERNAL);
    }
  }

  return new AccountsError(AccountsErrorCode.UNAVAILABLE);
}
