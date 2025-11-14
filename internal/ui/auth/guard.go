package authUI

import (
	"context"
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// AuthGuard wraps DISReaderService to enforce authentication transparently.
type AuthGuard struct {
	inner types.DISReaderService
	ui    types.UI
}

func NewAuthGuard(inner types.DISReaderService, ui types.UI) types.DISReaderService {
	return &AuthGuard{inner: inner, ui: ui}
}

func (a *AuthGuard) GetConfig() *types.DISConfig {
	return a.inner.GetConfig()
}

func (a *AuthGuard) GetLogger() *slog.Logger {
	return a.inner.GetLogger()
}

func (a *AuthGuard) SetLogger(logger *slog.Logger) {
	a.inner.SetLogger(logger)
}

func (a *AuthGuard) Connect(ctx context.Context) error {
	return a.inner.Connect(ctx)
}

func (a *AuthGuard) Disconnect(ctx context.Context) error {
	return a.inner.Disconnect(ctx)
}

func (a *AuthGuard) TestConnection(ctx context.Context) error {
	return a.inner.TestConnection(ctx)
}

func (a *AuthGuard) RunDebugSearch(term string, opts types.DebugSearchOptions, pCh chan<- types.ProgressStatus, eCh chan<- types.TableEvent) {
	if !a.ui.GetAuth().IsAuthenticated() {
		optsCopy := opts
		a.ui.SetPendingAction(func() { a.inner.RunDebugSearch(term, optsCopy, pCh, eCh) })
		a.ui.GetAuth().ShowAuthModal()
		return
	}
	a.inner.RunDebugSearch(term, opts, pCh, eCh)
}

func (a *AuthGuard) Shutdown() error {
	return a.inner.Shutdown()
}

func (a *AuthGuard) AttachShutdownHook() {
	a.inner.AttachShutdownHook()
}

func (a *AuthGuard) UnitService() types.UnitService {
	return a.inner.UnitService()
}

func (a *AuthGuard) InvoiceService() types.InvoiceService {
	return a.inner.InvoiceService()
}
