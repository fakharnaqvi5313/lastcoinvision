package keeper

import (
	"github.com/lastcoinvision/lcv/x/voting/types"
)

var _ types.QueryServer = Keeper{}
