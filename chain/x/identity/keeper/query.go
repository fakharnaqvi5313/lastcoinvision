package keeper

import (
	"github.com/lastcoinvision/lcv/x/identity/types"
)

var _ types.QueryServer = Keeper{}
