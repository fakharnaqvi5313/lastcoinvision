package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/group"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

// ComputeTier computes an address's account tier. Precedence (highest first):
//
//  1. institutional — address is a x/group group-policy account with
//     non-empty metadata (the group module's own metadata field is reused
//     as the "role" descriptor rather than inventing a parallel registry).
//  2. attested — address holds at least one ACTIVE, unrevoked credential of
//     params.AttestedCredentialType issued by a *currently* authorized
//     issuer (an issuer removed after issuance no longer counts).
//  3. pseudonymous — the default for everyone else.
//
// roleMetadata is only non-empty when tier is institutional.
func (k Keeper) ComputeTier(ctx context.Context, address string) (tier types.AccountTier, roleMetadata string) {
	if info, ok := k.groupPolicyInfo(ctx, address); ok && info.Metadata != "" {
		return types.AccountTier_ACCOUNT_TIER_INSTITUTIONAL, info.Metadata
	}

	params := k.GetParams(ctx)
	for _, cred := range k.GetCredentialsBySubject(ctx, address) {
		if cred.Status != types.CredentialStatus_CREDENTIAL_STATUS_ACTIVE {
			continue
		}
		if cred.CredentialType != params.AttestedCredentialType {
			continue
		}
		if !k.IsAuthorizedIssuer(ctx, cred.Issuer) {
			continue
		}
		return types.AccountTier_ACCOUNT_TIER_ATTESTED, ""
	}

	return types.AccountTier_ACCOUNT_TIER_PSEUDONYMOUS, ""
}

func (k Keeper) groupPolicyInfo(ctx context.Context, address string) (*group.GroupPolicyInfo, bool) {
	if k.groupKeeper == nil {
		return nil, false
	}
	res, err := k.groupKeeper.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{Address: address})
	if err != nil || res == nil || res.Info == nil {
		return nil, false
	}
	return res.Info, true
}
