export const AccountsErrorCode = {
  INVALID_ARGUMENT: 'INVALID_ARGUMENT',
  UNAUTHENTICATED: 'UNAUTHENTICATED',
  PERMISSION_DENIED: 'PERMISSION_DENIED',
  ALREADY_EXISTS: 'ALREADY_EXISTS',
  NOT_FOUND: 'NOT_FOUND',
  UNAVAILABLE: 'UNAVAILABLE',
  EMAIL_NOT_VERIFIED: 'EMAIL_NOT_VERIFIED',
  INTERNAL: 'INTERNAL',
} as const;

export type AccountsErrorCode = (typeof AccountsErrorCode)[keyof typeof AccountsErrorCode];

export class AccountsError extends Error {
  readonly code: AccountsErrorCode;

  constructor(code: AccountsErrorCode, message?: string) {
    super(message ?? code);
    this.name = 'AccountsError';
    this.code = code;
  }
}
