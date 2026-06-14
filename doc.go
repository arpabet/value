/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package value provides an immutable, dynamically-typed value tree with
// deterministic, canonical MessagePack serialization.
//
// The defining feature is determinism: a given value always encodes to exactly
// the same bytes, so those bytes can be hashed and signed reproducibly. This
// makes the package well suited to cryptographic protocols and blockchain
// systems, where content addressing, Merkle trees, and signature verification
// all depend on a stable canonical byte representation. See CANONICAL.md for
// the specification.
package value

