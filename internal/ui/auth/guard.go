package authUI

import (
	"context"

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

func (a *AuthGuard) TestConnection(ctx context.Context) error {
	return a.inner.TestConnection(ctx)
}

func (a *AuthGuard) RunDebugSearch(term, file string, pCh chan<- types.ProgressStatus, eCh chan<- types.TableEvent) {
	if !a.ui.GetAuth().IsAuthenticated() {
		a.ui.SetPendingAction(func() { a.inner.RunDebugSearch(term, file, pCh, eCh) })
		a.ui.GetAuth().ShowAuthModal()
		return
	}
	a.inner.RunDebugSearch(term, file, pCh, eCh)
}

func (a *AuthGuard) Shutdown() error {
	return a.inner.Shutdown()
}

func (a *AuthGuard) AttachShutdownHook() {
	a.inner.AttachShutdownHook()
}
