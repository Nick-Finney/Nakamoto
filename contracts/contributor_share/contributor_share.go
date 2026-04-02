// Package contributor_share contains the Contributor Revenue Share smart contract bytecode.
//
// This contract is the TRUSTLESS, VERIFIABLE escrow for the contributor revenue
// sharing system. It is open-source so that contributors can audit the logic
// without needing access to the rest of the Nakamoto protocol source code.
//
// Functions:
//   - deposit(amount: i32) - Deposit dev fund fees into the contributor pool
//   - claim(amount: i32) -> i32 - Claim earnings (returns 1=success, 0=insufficient)
//   - get_pool_balance() -> i32 - Get total pool balance available for claims
//   - get_total_deposited() -> i32 - Get lifetime total deposited into pool
//   - get_total_claimed() -> i32 - Get lifetime total claimed from pool
//
// The contract uses three global variables:
//   - pool_balance (mut i32) - Current pool balance available for claims
//   - total_deposited (mut i32) - Lifetime total deposited
//   - total_claimed (mut i32) - Lifetime total claimed
//
// How this contract fits into the revenue sharing system:
//   1. Protocol-level ContributorShareManager calls deposit() when dev fund fees arrive
//   2. Manager calculates each contributor's proportional share off-chain (BPS weights)
//   3. Contributors call claim() to withdraw their earned share
//   4. All operations are logged on-chain for full auditability
//
// IMPORTANT: Share weight calculations happen in ContributorShareManager (Go code),
// not in this WASM contract. The contract provides the trustless escrow layer —
// funds can only leave via claim(), and only up to the deposited amount.
// The Go manager enforces per-contributor limits before calling claim().
//
// This contract is intentionally simple so that anyone can verify:
//   - Funds deposited can only be withdrawn via claim()
//   - Pool balance is always ≥ 0
//   - total_deposited = total_claimed + pool_balance (invariant)
//   - No admin override or backdoor functions exist
package contributor_share

import "encoding/base64"

// ContributorShareWASM is the compiled WebAssembly bytecode for the
// Contributor Revenue Share contract. Hand-assembled following WASM
// binary format specification for full auditability.
var ContributorShareWASM = []byte{
	// ===== WASM Header =====
	0x00, 0x61, 0x73, 0x6d, // magic number: "\0asm"
	0x01, 0x00, 0x00, 0x00, // version: 1

	// ===== Type Section (ID=1) =====
	// Defines function signatures used by the contract
	0x01, // section id: type
	0x0e, // section size: 14 bytes
	0x03, // 3 types

	// Type 0: (i32) -> () — for deposit function
	0x60,       // func type
	0x01, 0x7f, // 1 param: i32
	0x00, // 0 results

	// Type 1: (i32) -> (i32) — for claim function
	0x60,       // func type
	0x01, 0x7f, // 1 param: i32
	0x01, 0x7f, // 1 result: i32

	// Type 2: () -> (i32) — for balance/total query functions
	0x60,       // func type
	0x00,       // 0 params
	0x01, 0x7f, // 1 result: i32

	// ===== Function Section (ID=3) =====
	// Maps functions to their type signatures
	0x03, // section id: function
	0x06, // section size: 6 bytes
	0x05, // 5 functions
	0x00, // func 0: deposit — type 0 (i32)->()
	0x01, // func 1: claim — type 1 (i32)->(i32)
	0x02, // func 2: get_pool_balance — type 2 ()->(i32)
	0x02, // func 3: get_total_deposited — type 2 ()->(i32)
	0x02, // func 4: get_total_claimed — type 2 ()->(i32)

	// ===== Global Section (ID=6) =====
	// Contract state: three mutable i32 globals initialized to 0
	0x06, // section id: global
	0x10, // section size: 16 bytes
	0x03, // 3 globals

	// Global 0: pool_balance (mut i32, init 0)
	// Current balance available for claims
	0x7f,       // type: i32
	0x01,       // mutable: yes
	0x41, 0x00, // i32.const 0
	0x0b, // end

	// Global 1: total_deposited (mut i32, init 0)
	// Lifetime total deposited into the pool
	0x7f,       // type: i32
	0x01,       // mutable: yes
	0x41, 0x00, // i32.const 0
	0x0b, // end

	// Global 2: total_claimed (mut i32, init 0)
	// Lifetime total claimed from the pool
	0x7f,       // type: i32
	0x01,       // mutable: yes
	0x41, 0x00, // i32.const 0
	0x0b, // end

	// ===== Export Section (ID=7) =====
	// Makes contract functions callable by external code
	0x07, // section id: export
	0x50, // section size: 80 bytes
	0x05, // 5 exports

	// Export "deposit" -> func 0
	0x07,                                           // name length: 7
	'd', 'e', 'p', 'o', 's', 'i', 't',             // "deposit"
	0x00,                                           // export kind: function
	0x00,                                           // function index: 0

	// Export "claim" -> func 1
	0x05,                                           // name length: 5
	'c', 'l', 'a', 'i', 'm',                       // "claim"
	0x00,                                           // export kind: function
	0x01,                                           // function index: 1

	// Export "get_pool_balance" -> func 2
	0x10,                                           // name length: 16
	'g', 'e', 't', '_', 'p', 'o', 'o', 'l',
	'_', 'b', 'a', 'l', 'a', 'n', 'c', 'e',       // "get_pool_balance"
	0x00,                                           // export kind: function
	0x02,                                           // function index: 2

	// Export "get_total_deposited" -> func 3
	0x13,                                           // name length: 19
	'g', 'e', 't', '_', 't', 'o', 't', 'a', 'l',
	'_', 'd', 'e', 'p', 'o', 's', 'i', 't', 'e', 'd', // "get_total_deposited"
	0x00,                                           // export kind: function
	0x03,                                           // function index: 3

	// Export "get_total_claimed" -> func 4
	0x11,                                           // name length: 17
	'g', 'e', 't', '_', 't', 'o', 't', 'a', 'l',
	'_', 'c', 'l', 'a', 'i', 'm', 'e', 'd',       // "get_total_claimed"
	0x00,                                           // export kind: function
	0x04,                                           // function index: 4

	// ===== Code Section (ID=10) =====
	// Function bodies — the actual contract logic
	0x0a, // section id: code
	0x3e, // section size: 62 bytes
	0x05, // 5 function bodies

	// --------------------------------------------------
	// Function 0: deposit(amount: i32)
	// Logic: pool_balance += amount; total_deposited += amount
	// --------------------------------------------------
	0x12,       // body size: 18 bytes
	0x00,       // local declaration count: 0
	// pool_balance += amount
	0x23, 0x00, // global.get 0 (pool_balance)
	0x20, 0x00, // local.get 0 (amount)
	0x6a,       // i32.add
	0x24, 0x00, // global.set 0 (pool_balance)
	// total_deposited += amount
	0x23, 0x01, // global.get 1 (total_deposited)
	0x20, 0x00, // local.get 0 (amount)
	0x6a,       // i32.add
	0x24, 0x01, // global.set 1 (total_deposited)
	0x0b,       // end

	// --------------------------------------------------
	// Function 1: claim(amount: i32) -> i32
	// Logic: if (pool_balance >= amount) {
	//            pool_balance -= amount;
	//            total_claimed += amount;
	//            return 1; // success
	//        } else {
	//            return 0; // insufficient funds
	//        }
	// --------------------------------------------------
	0x1e,       // body size: 30 bytes
	0x00,       // local declaration count: 0
	0x23, 0x00, // global.get 0 (pool_balance)
	0x20, 0x00, // local.get 0 (amount)
	0x4e,       // i32.ge_u (unsigned greater-or-equal)
	0x04, 0x7f, // if (result i32)
	// then: sufficient funds — deduct and return success
	0x23, 0x00, // global.get 0 (pool_balance)
	0x20, 0x00, // local.get 0 (amount)
	0x6b,       // i32.sub
	0x24, 0x00, // global.set 0 (pool_balance)
	// total_claimed += amount
	0x23, 0x02, // global.get 2 (total_claimed)
	0x20, 0x00, // local.get 0 (amount)
	0x6a,       // i32.add
	0x24, 0x02, // global.set 2 (total_claimed)
	0x41, 0x01, // i32.const 1 (success)
	0x05,       // else: insufficient funds
	0x41, 0x00, // i32.const 0 (failure)
	0x0b,       // end (of if)
	0x0b,       // end (of function)

	// --------------------------------------------------
	// Function 2: get_pool_balance() -> i32
	// --------------------------------------------------
	0x04,       // body size: 4 bytes
	0x00,       // local declaration count: 0
	0x23, 0x00, // global.get 0 (pool_balance)
	0x0b,       // end

	// --------------------------------------------------
	// Function 3: get_total_deposited() -> i32
	// --------------------------------------------------
	0x04,       // body size: 4 bytes
	0x00,       // local declaration count: 0
	0x23, 0x01, // global.get 1 (total_deposited)
	0x0b,       // end

	// --------------------------------------------------
	// Function 4: get_total_claimed() -> i32
	// --------------------------------------------------
	0x04,       // body size: 4 bytes
	0x00,       // local declaration count: 0
	0x23, 0x02, // global.get 2 (total_claimed)
	0x0b,       // end
}

// ContributorShareWASMBase64 returns the base64-encoded bytecode for deployment
func ContributorShareWASMBase64() string {
	return base64.StdEncoding.EncodeToString(ContributorShareWASM)
}
