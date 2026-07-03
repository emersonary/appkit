package runtime

import (
	"context"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/ai"
	"github.com/emersonary/appkit/currency"
	"github.com/emersonary/appkit/email"
	"github.com/emersonary/appkit/language"
	"github.com/emersonary/appkit/menu"
	"github.com/emersonary/appkit/permissions"
	"github.com/emersonary/appkit/tenants"
	"github.com/emersonary/appkit/weather"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func (a *Application[T]) wireEmail() {
	base := a.Base()
	a.Email = email.NewHandler(base.Email, base.MailProvisioning, a.Logger.Named("email"))
}

func (a *Application[T]) wireBlocks(ctx context.Context, opts Options[T]) error {
	base := a.Base()
	if !base.Accounts.Enabled && !base.Tenants.Enabled && !base.Currency.Enabled && !base.Language.Enabled && !base.Permissions.Enabled && !base.Menu.Enabled && !base.AI.Enabled && !base.Weather.Enabled {
		return nil
	}

	// sqlDB wraps the application pool; do not Close here — Shutdown closes a.Pool.
	sqlDB := stdlib.OpenDBFromPool(a.Pool)

	if base.Permissions.Enabled {
		svc, err := permissions.Wire(ctx, sqlDB, base.Permissions, permissions.WireOptions{})
		if err != nil {
			return err
		}
		a.Permissions = svc
		if svc != nil {
			a.Logger.Info("permissions block enabled",
				zap.String("default_profile", base.Permissions.DefaultProfile),
			)
		}
	}

	if base.Menu.Enabled {
		svc, err := menu.Wire(ctx, base.Menu, menu.WireOptions{Permissions: a.Permissions})
		if err != nil {
			return err
		}
		a.Menu = svc
		if svc != nil {
			a.Logger.Info("menu block enabled",
				zap.Int("menus", len(svc.Setup().Menus)),
			)
		}
	}

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
		if accOpts.AfterCreate == nil && a.Permissions != nil {
			perm := a.Permissions
			accOpts.AfterCreate = func(ctx context.Context, account accounts.Account, registerAsAdmin bool) error {
				return perm.AssignNewAccountProfile(ctx, account.ID, registerAsAdmin || account.IsAdmin)
			}
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

	if base.Tenants.Enabled {
		svc, err := tenants.Wire(ctx, sqlDB, base.Tenants)
		if err != nil {
			return err
		}
		a.Tenants = svc
		if svc != nil {
			a.Logger.Info("tenants block enabled")
		}
	}

	if base.Currency.Enabled {
		workerCtx, _ := a.RegisterWorker()

		curOpts := currency.WireOptions{Logger: a.Logger.Named("currency"), WorkerCtx: workerCtx}
		if opts.CurrencyWire != nil {
			var err error
			curOpts, err = opts.CurrencyWire(ctx, a, workerCtx)
			if err != nil {
				a.stopWorkers()
				return err
			}
		}

		svc, err := currency.Wire(ctx, sqlDB, base.Currency, curOpts)
		if err != nil {
			a.stopWorkers()
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

	if base.Language.Enabled {
		svc, err := language.Wire(ctx, sqlDB, base.Language, language.WireOptions{})
		if err != nil {
			return err
		}
		a.Language = svc
		if svc != nil {
			a.Logger.Info("language block enabled",
				zap.String("default", base.Language.DefaultLanguage),
			)
		}
	}

	if base.AI.Enabled {
		svc, err := ai.Wire(ctx, base.AI, ai.WireOptions{})
		if err != nil {
			return err
		}
		a.AI = svc
		if svc != nil {
			a.Logger.Info("ai block enabled",
				zap.Any("routes", svc.RouteSummary()),
			)
		}
	}

	if base.Weather.Enabled {
		workerCtx, _ := a.RegisterWorker()
		svc, err := weather.Wire(ctx, base.Weather, weather.WireOptions{
			Redis:     a.Redis,
			Logger:    a.Logger.Named("weather"),
			WorkerCtx: workerCtx,
		})
		if err != nil {
			a.stopWorkers()
			return err
		}
		a.Weather = svc
		if svc != nil {
			a.Logger.Info("weather block enabled",
				zap.String("key_prefix", svc.Config().KeyPrefix),
				zap.Duration("refresh_interval", svc.Config().RefreshInterval),
			)
		}
	}

	return nil
}
