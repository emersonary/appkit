package runtime

import (
	"context"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/currency"
	"github.com/emersonary/appkit/email"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func (a *Application[T]) wireEmail() {
	base := a.Base()
	a.Email = email.NewHandler(base.Email, base.MailProvisioning, a.Logger.Named("email"))
}

func (a *Application[T]) wireBlocks(ctx context.Context, opts Options[T]) error {
	base := a.Base()
	if !base.Accounts.Enabled && !base.Currency.Enabled {
		return nil
	}

	sqlDB := stdlib.OpenDBFromPool(a.Pool)
	defer sqlDB.Close()

	if base.Accounts.Enabled {
		accOpts := accounts.Options{}
		if opts.AccountsWire != nil {
			var err error
			accOpts, err = opts.AccountsWire(ctx, a)
			if err != nil {
				return err
			}
		}
		if accOpts.Mailer == nil && a.Email != nil {
			accOpts.Mailer = a.Email.AccountMailer()
		}

		svc, err := accounts.Wire(ctx, sqlDB, base.Accounts, accOpts)
		if err != nil {
			return err
		}
		a.Accounts = svc
		if svc != nil {
			a.Logger.Info("accounts block enabled")
		}
	}

	if base.Currency.Enabled {
		workerCtx, workerCancel := context.WithCancel(context.Background())
		a.workerCancel = workerCancel

		curOpts := currency.WireOptions{Logger: a.Logger.Named("currency"), WorkerCtx: workerCtx}
		if opts.CurrencyWire != nil {
			var err error
			curOpts, err = opts.CurrencyWire(ctx, a, workerCtx)
			if err != nil {
				if a.workerCancel != nil {
					a.workerCancel()
					a.workerCancel = nil
				}
				return err
			}
		}

		svc, err := currency.Wire(ctx, sqlDB, base.Currency, curOpts)
		if err != nil {
			if a.workerCancel != nil {
				a.workerCancel()
				a.workerCancel = nil
			}
			return err
		}
		a.Currency = svc
		if svc != nil {
			a.Logger.Info("currency block enabled",
				zap.Duration("interval", base.Currency.UpdateInterval),
				zap.String("api", base.Currency.APIURL),
			)
		}
	}

	return nil
}
