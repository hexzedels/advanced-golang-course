#!/bin/sh
# Show that sync.Mutex.Lock inlines all the way down to one LOCK CMPXCHGL,
# and Unlock to one LOCK XADDL. Run from the module root:
#   sh 01-fastpath/dis.sh
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Reconstructed 2026-10-25 from demos.md §1 ("The disassembly"); not in the
# original document. demos.md gives this header and main.go; the rest is
# inferred from the captured output: a throwaway module, `-gcflags='-m -m'`
# for the inlining decisions, `-gcflags=-S` for the listing, FUNCDATA/PCDATA
# filtered, the decimal-offset column, hex dump and relocations dropped, and
# the absolute paths shortened to main.go:N and src/...:N.
#
# The pinned toolchain must answer. On the WSL box the shell exports
# GOROOT=/usr/local/go (go1.25.3); a GOROOT whose VERSION is not go1.27.1 is
# dropped below, and the toolchain is resolved from this module (go 1.27.1 in
# go.mod), then called by path with GOTOOLCHAIN=local, because the throwaway
# module lives outside the repository (docs/environment-notes.md §0.3).
# GO=<a go binary> overrides the command used to resolve it.
set -eu
GO="${GO:-go}"

if [ -n "${GOROOT:-}" ] && [ "$(head -1 "$GOROOT/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	unset GOROOT
fi
G=$("$GO" env GOROOT)
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "dis.sh: go env GOROOT is $G, whose VERSION is not go1.27.1; run from the demos module." >&2
	exit 1
fi
echo "dis.sh: toolchain $G ($(head -1 "$G/VERSION"))" >&2

tmp=$(mktemp -d "${TMPDIR:-/tmp}/w04dis.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
printf 'module dis\n\ngo 1.27.1\n' >"$tmp/go.mod"
cat >"$tmp/main.go" <<'EOF'
package main

import "sync"

var mu sync.Mutex
var n int

//go:noinline
func Crit() {
	mu.Lock()
	n++
	mu.Unlock()
}

func main() { Crit() }
EOF

cd "$tmp"
export GOTOOLCHAIN=local
unset GOFLAGS

echo "===== inlining decisions ====="
"$G/bin/go" build -gcflags='-m -m' -o /dev/null . 2>&1 | grep 'inlining call to'

echo
echo "===== generated code for Crit ====="
# -S lines are: TAB "0xHHHH DDDDD (path:line)" TAB op [TAB args]. Keep main.Crit's
# block only, drop FUNCDATA/PCDATA, the decimal offset, the hex dump and the
# relocations, and shorten the paths.
"$G/bin/go" build -gcflags='-S' -o /dev/null . 2>&1 |
	awk -F'\t' -v g="$G/" -v t="$tmp/" '
	/^main\.Crit STEXT/ { on = 1; print; next }
	/^[^\t]/ { on = 0 }
	!on || NF < 3 || $2 !~ /^0x[0-9a-f]+ [0-9]+ \(/ { next }
	$3 == "FUNCDATA" || $3 == "PCDATA" { next }
	{
		split($2, a, " ")
		pos = substr($2, index($2, "("))
		sub(/^\(/, "", pos); sub(/\)$/, "", pos)
		if (index(pos, g) == 1) pos = substr(pos, length(g) + 1)
		if (index(pos, t) == 1) pos = substr(pos, length(t) + 1)
		line = sprintf("%s %-33s %-9s %s", a[1], "(" pos ")", $3, $4)
		sub(/ +$/, "", line)
		print line
	}'
