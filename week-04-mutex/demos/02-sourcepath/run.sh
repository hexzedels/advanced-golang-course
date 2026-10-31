#!/bin/sh
# Usage: sh 02-sourcepath/run.sh [/path/to/goroot]
#
# Where sync.Mutex actually lives: the interesting words are all in the file the
# student did not open. Run from the demos module (the directory holding
# go.mod). With no argument it reads the pinned toolchain's tree, resolved as
# `go env GOROOT` from this module with any foreign GOROOT dropped; on the WSL
# box that is ~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.1.linux-amd64,
# never /usr/local/go (go1.25.3). It refuses a tree whose VERSION is not
# go1.27.1, and its first line names the tree it read.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Reconstructed 2026-10-25 from demos.md §2; not in the original document.
# demos.md gives an abridged script (the usage line, the SRC line, the first
# two printf lines and the last two greps) and the full captured output; the
# remaining lines are inferred from that output. Changed from the abridged
# copy: the default tree is the pinned GOROOT, not /tmp/gosrc, and the
# VERSION check is new. "occurrences" are grep -c counts, i.e. matching
# lines; at go1.27.1 the two agree for every word counted here.
set -eu
GO="${GO:-go}"

if [ $# -ge 1 ]; then
	ROOT=$1
else
	if [ -n "${GOROOT:-}" ] && [ "$(head -1 "$GOROOT/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
		unset GOROOT
	fi
	ROOT=$("$GO" env GOROOT)
fi
SRC="$ROOT/src"
V=$(head -1 "$ROOT/VERSION" 2>/dev/null || echo unknown)
if [ "$V" != "go1.27.1" ]; then
	echo "run.sh: $ROOT is $V, not go1.27.1. Pass the pinned GOROOT:" >&2
	echo "run.sh:     sh 02-sourcepath/run.sh \"\$(env -u GOROOT go env GOROOT)\"" >&2
	exit 1
fi

echo "GOROOT source: $SRC ($V)"
echo
echo "===== what a student opens ====="
printf 'src/sync/mutex.go        %s lines\n' "$(wc -l <"$SRC/sync/mutex.go" | tr -d ' ')"
for w in starvation CompareAndSwap lockSlow; do
	printf '  occurrences of "%s": %s\n' "$w" "$(grep -c "$w" "$SRC/sync/mutex.go" || :)"
done
echo
echo "===== where the algorithm is ====="
printf 'src/internal/sync/mutex.go  %s lines\n' "$(wc -l <"$SRC/internal/sync/mutex.go" | tr -d ' ')"
for w in starvation CompareAndSwap lockSlow; do
	printf '  occurrences of "%s": %s\n' "$w" "$(grep -c "$w" "$SRC/internal/sync/mutex.go" || :)"
done
echo
echo "===== the constant ====="
grep -n "starvationThresholdNs" "$SRC/internal/sync/mutex.go"
echo
echo "===== and sync.Map ====="
grep -n "isync\|HashTrieMap\|readOnly\|dirty" "$SRC/sync/map.go" | head -10
printf 'src/sync/map.go             %s lines\n' "$(wc -l <"$SRC/sync/map.go" | tr -d ' ')"
printf 'src/internal/sync/hashtriemap.go %s lines\n' "$(wc -l <"$SRC/internal/sync/hashtriemap.go" | tr -d ' ')"
