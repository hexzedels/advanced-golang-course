// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

package fastpath

import (
	"sync"
	"testing"
	"unsafe"
)

// Reconstructed 2026-10-25 from demos.md §1 ("The sizes"); not in the original document.
func TestSizes(t *testing.T) {
	t.Logf("sync.Mutex        %2d bytes", unsafe.Sizeof(sync.Mutex{}))
	t.Logf("sync.RWMutex      %2d bytes", unsafe.Sizeof(sync.RWMutex{}))
	t.Logf("sync.Once         %2d bytes", unsafe.Sizeof(sync.Once{}))
	t.Logf("sync.WaitGroup    %2d bytes", unsafe.Sizeof(sync.WaitGroup{}))
	t.Logf("sync.Map          %2d bytes", unsafe.Sizeof(sync.Map{}))
	t.Logf("sync.Cond         %2d bytes", unsafe.Sizeof(sync.Cond{}))
	t.Logf("sync.Pool         %2d bytes", unsafe.Sizeof(sync.Pool{}))
	if got := unsafe.Sizeof(sync.Mutex{}); got != 8 {
		t.Fatalf("sync.Mutex is %d bytes, want 8", got)
	}
}

// Not a test, a measurement. The seven Logf calls sit on lines 14-20 so that
// the output's sizes_test.go:NN prefixes match demos.md's capture. Only the
// Mutex size is asserted: demo 3 reads the state word through an
// unsafe.Pointer, and if the layout ever changes this fails first.
