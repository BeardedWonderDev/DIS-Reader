package authUI

import (
	"context"
	"fmt"

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
		if err := a.UI.GetDIS().TestConnection(ctx); err != nil {
			if err := a.UI.GetDIS().Connect(ctx); err != nil {
				a.UI.GetLogger().Error(fmt.Sprintf("DIS Connection Failed: %s", err))
				if onComplete != nil {
					onComplete(false, err)
				}
				return
			}

			err := a.UI.GetDIS().TestConnection(ctx)
			success := err == nil
			if err != nil {
				a.UI.GetLogger().Error(fmt.Sprintf("DIS Connection Failed: %s", err))
			} else {
				a.Authenticated = true
				a.UI.GetLogger().Info("DIS Connection Successful")
			}

			if onComplete != nil {
				onComplete(success, err)
			}
		} else {
			a.Authenticated = true
			if onComplete != nil {
				onComplete(true, err)
			}
		}
	}()
}
