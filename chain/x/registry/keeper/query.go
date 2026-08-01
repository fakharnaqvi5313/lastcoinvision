package keeper

import (
	"github.com/lastcoinvision/lcv/x/registry/types"
)

var _ types.QueryServer = Keeper{}
