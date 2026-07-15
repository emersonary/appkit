package runtime

import (
	"context"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/ai"
	"github.com/emersonary/appkit/currency"
	"github.com/emersonary/appkit/email"
	"github.com/emersonary/appkit/language"
	"github.com/emersonary/appkit/menu"
	"github.com/emersonary/appkit/migrate"
	"github.com/emersonary/appkit/permissions"
	"github.com/emersonary/appkit/tenants"
	"github.com/jackc/pgx/v5/stdlib"
)

func (a *Application[T]) wireEmail() {
	base := a.Base()
	a.Email = email.NewService(base.Email, base.MailProvisioning, a.Logger.Named("email"))
}

func (a *Application[T]) wireBlocks(ctx context.Context, opts Options[T]) error {
	base := a.Base()
	if !base.Accounts.Enabled && !base.Tenants.Enabled && !base.Currency.Enabled && !base.Language.Enabled && !base.Permissions.Enabled && !base.Menu.Enabled && !base.AI.Enabled && !base.DBHist.Enabled {
		return nil
	}

	// sqlDB wraps the application pool; do not Close here — Shutdown closes a.Pool.
	sqlDB := stdlib.OpenDBFromPool(a.Pool)

	var dbhistWorker context.Context
	if base.DBHist.Enabled {
		dbhistWorker, _ = a.RegisterWorker()
	}
	svc, err := migrate.Wire(ctx, sqlDB, base.DBHist, migrate.WireOptions{
		Logger:    a.Logger.Named("migrate"),
		WorkerCtx: dbhistWorker,
	})
	if err != nil {
		return err
	}
	a.DBHist = svc

	permSvc, err := permissions.Wire(ctx, sqlDB, base.Permissions, permissions.WireOptions{
		Logger: a.Logger.Named("permissions"),
	})
	if err != nil {
		return err
	}
	a.Permissions = permSvc

	menuSvc, err := menu.Wire(ctx, base.Menu, menu.WireOptions{
		Permissions: a.Permissions,
		Logger:      a.Logger.Named("menu"),
	})
	if err != nil {
		return err
	}
	a.Menu = menuSvc

	// AccountsWire / default hooks are caller-side prep; only build when accounts will wire.
	accOpts := accounts.Options{}
	if base.Accounts.Enabled {
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
		if accOpts.Logger == nil {
			accOpts.Logger = a.Logger.Named("accounts")
		}
	}

	accSvc, err := accounts.Wire(ctx, sqlDB, base.Accounts, accOpts)
	if err != nil {
		return err
	}
	a.Accounts = accSvc

	tenSvc, err := tenants.Wire(ctx, sqlDB, base.Tenants, tenants.WireOptions{
		Logger: a.Logger.Named("tenants"),
	})
	if err != nil {
		return err
	}
	a.Tenants = tenSvc

	// RegisterWorker has side effects; only prepare when currency will actually wire.
	curOpts := currency.WireOptions{Logger: a.Logger.Named("currency")}
	if base.Currency.Enabled {
		workerCtx, _ := a.RegisterWorker()
		curOpts.WorkerCtx = workerCtx
		if opts.CurrencyWire != nil {
			var err error
			curOpts, err = opts.CurrencyWire(ctx, a, workerCtx)
			if err != nil {
				a.stopWorkers()
				return err
			}
		}
		if curOpts.Logger == nil {
			curOpts.Logger = a.Logger.Named("currency")
		}
	}

	curSvc, err := currency.Wire(ctx, sqlDB, base.Currency, curOpts)
	if err != nil {
		a.stopWorkers()
		return err
	}
	a.Currency = curSvc

	langSvc, err := language.Wire(ctx, sqlDB, base.Language, language.WireOptions{
		Logger: a.Logger.Named("language"),
	})
	if err != nil {
		return err
	}
	a.Language = langSvc

	aiSvc, err := ai.Wire(ctx, base.AI, ai.WireOptions{
		Logger: a.Logger.Named("ai"),
	})
	if err != nil {
		return err
	}
	a.AI = aiSvc

	return nil
}
