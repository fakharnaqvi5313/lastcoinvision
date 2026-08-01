package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

const testJurisdictionCode = "us-ca"

var (
	testSenderAddr    = sdk.AccAddress([]byte("taxrail-sender______")).String()
	testRecipientAddr = sdk.AccAddress([]byte("taxrail-recipient___")).String()
	testTreasuryAddr  = sdk.AccAddress([]byte("taxrail-treasury____")).String()
)

// TestAcceptance_TaxedTransferLifecycle is the Phase C acceptance flow for
// x/taxrail: a governance-registered tax rule -> a taxed transfer that
// atomically routes the tax share to the jurisdiction's collector treasury
// and the remainder to the recipient -> a gov-gated spend of the collected
// tax out of the treasury.
func TestAcceptance_TaxedTransferLifecycle(t *testing.T) {
	k, ctx, bank := keepertest.TaxrailKeeperWithBank(t)
	ms := keeper.NewMsgServerImpl(k)

	senderAddr, err := sdk.AccAddressFromBech32(testSenderAddr)
	require.NoError(t, err)
	bank.FundAccount(senderAddr, sdk.NewCoins(sdk.NewCoin(types.DefaultDenom, math.NewInt(1000))))

	// a non-authority signer cannot register a tax rule
	_, err = ms.RegisterTaxRule(ctx, &types.MsgRegisterTaxRule{
		Authority:        testSenderAddr,
		JurisdictionCode: testJurisdictionCode,
		Rate:             math.LegacyNewDecWithPrec(10, 2),
	})
	require.ErrorIs(t, err, types.ErrInvalidSigner)

	// "governance-registered tax rule": 10% for us-ca
	_, err = ms.RegisterTaxRule(ctx, &types.MsgRegisterTaxRule{
		Authority:        k.GetAuthority(),
		JurisdictionCode: testJurisdictionCode,
		Rate:             math.LegacyNewDecWithPrec(10, 2),
	})
	require.NoError(t, err)

	// a transfer with no registered rule for the jurisdiction is rejected
	_, err = ms.TaxedTransfer(ctx, &types.MsgTaxedTransfer{
		Sender:           testSenderAddr,
		Recipient:        testRecipientAddr,
		JurisdictionCode: "unregistered-jurisdiction",
		Amount:           sdk.NewCoin(types.DefaultDenom, math.NewInt(100)),
	})
	require.ErrorIs(t, err, types.ErrNoTaxRule)

	// "taxed transfer atomically routes tax share + remainder"
	transferResp, err := ms.TaxedTransfer(ctx, &types.MsgTaxedTransfer{
		Sender:           testSenderAddr,
		Recipient:        testRecipientAddr,
		JurisdictionCode: testJurisdictionCode,
		Amount:           sdk.NewCoin(types.DefaultDenom, math.NewInt(100)),
	})
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10), transferResp.TaxAmount.Amount)
	require.Equal(t, math.NewInt(90), transferResp.RemainderAmount.Amount)

	// tax routed + remainder = input amount, by construction and cross-checked here
	require.True(t, transferResp.TaxAmount.Amount.Add(transferResp.RemainderAmount.Amount).Equal(math.NewInt(100)))

	recipientAddr, err := sdk.AccAddressFromBech32(testRecipientAddr)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(90), bank.GetBalance(ctx, recipientAddr, types.DefaultDenom).Amount)
	require.Equal(t, math.NewInt(900), bank.GetBalance(ctx, senderAddr, types.DefaultDenom).Amount)

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	require.Equal(t, math.NewInt(10), bank.GetBalance(ctx, moduleAddr, types.DefaultDenom).Amount)

	ledger := k.GetJurisdictionLedger(ctx, testJurisdictionCode)
	require.Equal(t, math.NewInt(10), ledger.Balance)

	// only the module denom is accepted
	_, err = ms.TaxedTransfer(ctx, &types.MsgTaxedTransfer{
		Sender:           testSenderAddr,
		Recipient:        testRecipientAddr,
		JurisdictionCode: testJurisdictionCode,
		Amount:           sdk.NewCoin("uatom", math.NewInt(100)),
	})
	require.ErrorIs(t, err, types.ErrUnsupportedDenom)

	// "collector outflows require gov proposals": non-authority cannot spend
	_, err = ms.SpendCollectedTax(ctx, &types.MsgSpendCollectedTax{
		Authority:        testSenderAddr,
		JurisdictionCode: testJurisdictionCode,
		Recipient:        testTreasuryAddr,
		Amount:           sdk.NewCoin(types.DefaultDenom, math.NewInt(10)),
		Reason:           "unauthorized attempt",
	})
	require.ErrorIs(t, err, types.ErrInvalidSigner)

	// spending more than collected fails
	_, err = ms.SpendCollectedTax(ctx, &types.MsgSpendCollectedTax{
		Authority:        k.GetAuthority(),
		JurisdictionCode: testJurisdictionCode,
		Recipient:        testTreasuryAddr,
		Amount:           sdk.NewCoin(types.DefaultDenom, math.NewInt(100)),
		Reason:           "over-spend attempt",
	})
	require.ErrorIs(t, err, types.ErrInsufficientLedger)

	// gov-authorized spend succeeds
	_, err = ms.SpendCollectedTax(ctx, &types.MsgSpendCollectedTax{
		Authority:        k.GetAuthority(),
		JurisdictionCode: testJurisdictionCode,
		Recipient:        testTreasuryAddr,
		Amount:           sdk.NewCoin(types.DefaultDenom, math.NewInt(10)),
		Reason:           "annual budget allocation",
	})
	require.NoError(t, err)

	ledger = k.GetJurisdictionLedger(ctx, testJurisdictionCode)
	require.True(t, ledger.Balance.IsZero())

	treasuryAddr, err := sdk.AccAddressFromBech32(testTreasuryAddr)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10), bank.GetBalance(ctx, treasuryAddr, types.DefaultDenom).Amount)
	require.True(t, bank.GetBalance(ctx, moduleAddr, types.DefaultDenom).Amount.IsZero())

	// invariant holds throughout
	msg, broken := keeper.LedgerMatchesModuleBalanceInvariant(k)(ctx)
	require.False(t, broken)
	require.Empty(t, msg)
}

func TestRegisterTaxRuleRejectsRateOutOfBounds(t *testing.T) {
	k, ctx, _ := keepertest.TaxrailKeeperWithBank(t)
	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.RegisterTaxRule(ctx, &types.MsgRegisterTaxRule{
		Authority:        k.GetAuthority(),
		JurisdictionCode: testJurisdictionCode,
		Rate:             math.LegacyNewDec(2),
	})
	require.ErrorIs(t, err, types.ErrInvalidRate)
}

func TestLedgerMatchesModuleBalanceInvariantCatchesDivergence(t *testing.T) {
	k, ctx, _ := keepertest.TaxrailKeeperWithBank(t)

	msg, broken := keeper.LedgerMatchesModuleBalanceInvariant(k)(ctx)
	require.False(t, broken)
	require.Empty(t, msg)

	// directly corrupt the ledger without moving any real bank funds
	require.NoError(t, k.AddToJurisdictionLedger(ctx, testJurisdictionCode, math.NewInt(50)))

	msg, broken = keeper.LedgerMatchesModuleBalanceInvariant(k)(ctx)
	require.True(t, broken)
	require.Contains(t, msg, "does not equal")
}
