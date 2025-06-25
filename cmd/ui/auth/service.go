package authUI

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
)

type Auth struct {
	UI typesUI.UI
}

func NewAuthService() *Auth {
	return &Auth{}
}
