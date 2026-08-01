package types_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

func TestStubZKVerifierAlwaysValid(t *testing.T) {
	var verifier types.ZKVerifier = types.StubZKVerifier{}

	valid, err := verifier.VerifyProof(context.Background(), types.Credential{}, nil)
	require.NoError(t, err)
	require.True(t, valid, "the stub is a placeholder that never rejects a proof — see the TODO on ZKVerifier")
}
