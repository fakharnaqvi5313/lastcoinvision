package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// FuzzMsgTaxedTransfer fuzzes the tax-rate/amount split at the core of
// MsgTaxedTransfer — the invariant that must hold for every reachable
// (rate, amount) pair is tax + remainder == amount exactly, with neither
// half ever negative (a negative half would panic inside sdk.NewCoin,
// since Coin construction itself is guaranteed to reject negative
// amounts). Rates are only ever fuzzed within RegisterTaxRule's own
// validated [0,1] range, since a rate outside that range can never reach
// a live jurisdiction's rule through the real message-passing path (see
// TestRegisterTaxRuleRejectsRateOutOfBounds).
func FuzzMsgTaxedTransfer(f *testing.F) {
	f.Add(uint64(0), uint64(100))
	f.Add(uint64(10000), uint64(100))
	f.Add(uint64(1), uint64(1))
	f.Add(uint64(9999), uint64(1))
	f.Add(uint64(5000), uint64(0))
	f.Add(uint64(3333), uint64(999_999_999_999))

	f.Fuzz(func(t *testing.T, rateBasisPoints uint64, amount uint64) {
		// Clamp to RegisterTaxRule's own accepted range: 0 <= rate <= 1,
		// expressed here as 0..10000 basis points (four decimal digits of
		// precision) — a live rule can never carry a rate outside this
		// range, so fuzzing outside it wouldn't be testing anything a real
		// jurisdiction rule could ever hit.
		rateBasisPoints %= 10001
		rate := math.LegacyNewDecWithPrec(int64(rateBasisPoints), 4)

		k, ctx, bank := keepertest.TaxrailKeeperWithBank(t)
		ms := keeper.NewMsgServerImpl(k)

		_, err := ms.RegisterTaxRule(ctx, &types.MsgRegisterTaxRule{
			Authority:        k.GetAuthority(),
			JurisdictionCode: testJurisdictionCode,
			Rate:             rate,
		})
		require.NoError(t, err)

		senderAddr, err := sdk.AccAddressFromBech32(testSenderAddr)
		require.NoError(t, err)
		amountInt := math.NewIntFromUint64(amount)
		bank.FundAccount(senderAddr, sdk.NewCoins(sdk.NewCoin(types.DefaultDenom, amountInt)))

		resp, err := ms.TaxedTransfer(ctx, &types.MsgTaxedTransfer{
			Sender:           testSenderAddr,
			Recipient:        testRecipientAddr,
			JurisdictionCode: testJurisdictionCode,
			Amount:           sdk.NewCoin(types.DefaultDenom, amountInt),
		})
		require.NoError(t, err)

		require.False(t, resp.TaxAmount.Amount.IsNegative(), "tax amount must never be negative (rate=%s amount=%s)", rate, amountInt)
		require.False(t, resp.RemainderAmount.Amount.IsNegative(), "remainder must never be negative (rate=%s amount=%s)", rate, amountInt)
		require.True(t, resp.TaxAmount.Amount.Add(resp.RemainderAmount.Amount).Equal(amountInt),
			"tax + remainder must equal the input amount exactly: got tax=%s remainder=%s amount=%s",
			resp.TaxAmount.Amount, resp.RemainderAmount.Amount, amountInt)

		msg, broken := keeper.LedgerMatchesModuleBalanceInvariant(k)(ctx)
		require.False(t, broken, "ledger/module-balance invariant broken: %s", msg)
	})
}
