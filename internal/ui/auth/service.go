package authUI

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type Auth struct {
	UI            types.UI
	Authenticated bool
}

func NewAuthService() *Auth {
	return &Auth{}
}

func (a *Auth) IsAuthenticated() bool {
	return a.Authenticated
}

func (a *Auth) Authenticate(onComplete func(success bool, err error)) {
	a.UI.GetLogger().Info("🌏 Verifying DIS Connection", "url", txtServerURL.GetText())

	go func() {
		ctx := context.TODO()
		if err := a.UI.GetDIS().Connect(ctx); err != nil {
			a.UI.GetLogger().Error("DIS Connection Failed", "error", err)
		}

		err := a.UI.GetDIS().TestConnection(ctx)
		success := err == nil
		if err != nil {
			a.UI.GetLogger().Error("DIS Connection Failed", "error", err)
		} else {
			a.Authenticated = true
			a.UI.GetLogger().Info("DIS Connection Successful")
		}

		if onComplete != nil {
			onComplete(success, err)
		}
	}()
}
