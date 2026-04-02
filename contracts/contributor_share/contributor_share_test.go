package contributor_share

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContributorShareWASM_ValidHeader(t *testing.T) {
	require.GreaterOrEqual(t, len(ContributorShareWASM), 8, "Bytecode too short")

	// WASM magic number: \0asm
	assert.Equal(t, byte(0x00), ContributorShareWASM[0])
	assert.Equal(t, byte(0x61), ContributorShareWASM[1])
	assert.Equal(t, byte(0x73), ContributorShareWASM[2])
	assert.Equal(t, byte(0x6d), ContributorShareWASM[3])

	// Version 1
	assert.Equal(t, byte(0x01), ContributorShareWASM[4])
	assert.Equal(t, byte(0x00), ContributorShareWASM[5])
	assert.Equal(t, byte(0x00), ContributorShareWASM[6])
	assert.Equal(t, byte(0x00), ContributorShareWASM[7])
}

func TestContributorShareWASM_Base64Encoding(t *testing.T) {
	b64 := ContributorShareWASMBase64()
	assert.NotEmpty(t, b64)
	t.Logf("Contributor Share WASM Base64: %s", b64)
	t.Logf("Bytecode length: %d bytes", len(ContributorShareWASM))
}

func TestContributorShareWASM_HasRequiredExports(t *testing.T) {
	bytecodeStr := string(ContributorShareWASM)

	assert.Contains(t, bytecodeStr, "deposit", "Should export 'deposit' function")
	assert.Contains(t, bytecodeStr, "claim", "Should export 'claim' function")
	assert.Contains(t, bytecodeStr, "get_pool_balance", "Should export 'get_pool_balance' function")
	assert.Contains(t, bytecodeStr, "get_total_deposited", "Should export 'get_total_deposited' function")
	assert.Contains(t, bytecodeStr, "get_total_claimed", "Should export 'get_total_claimed' function")
}

func TestContributorShareWASM_SectionStructure(t *testing.T) {
	// Verify the bytecode has the expected section IDs after the 8-byte header
	offset := 8
	seenSections := make(map[byte]bool)

	for offset < len(ContributorShareWASM)-1 {
		sectionID := ContributorShareWASM[offset]
		offset++

		if offset >= len(ContributorShareWASM) {
			break
		}

		// Read section size (single-byte LEB128 for our small contract)
		sectionSize := int(ContributorShareWASM[offset])
		offset++

		seenSections[sectionID] = true

		if offset+sectionSize > len(ContributorShareWASM) {
			break
		}
		offset += sectionSize
	}

	// Must have: type(1), function(3), global(6), export(7), code(10)
	assert.True(t, seenSections[1], "Missing Type section")
	assert.True(t, seenSections[3], "Missing Function section")
	assert.True(t, seenSections[6], "Missing Global section")
	assert.True(t, seenSections[7], "Missing Export section")
	assert.True(t, seenSections[10], "Missing Code section")
}
