#!/bin/sh
# cues.sh -- L-04, the week 4 lecture: the run cues of demos.md "Run order for
# the lecture", as a runner for a presenter who drives alone with one hand. One
# key per cue: nothing is typed during the lecture. Copied from week 3's
# cues.sh (the pin and GOROOT checks, the key loop, --list/--present/--short and
# the single-cue mode are week 3's); the CUES table and the warm-up are this
# week's.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Usage:  sh cues.sh            warm the build cache, then walk the cues one key each
#         sh cues.sh --short    the 60-minute deck's live cues only (none marked yet:
#                               mark rows short when that deck is cut)
#         sh cues.sh --present  hide the Expect and fallback lines and clear the
#                               screen at each cue, for a shared screen
#         sh cues.sh --list     print the cues, their expected output and the
#                               fallback line, and run nothing; combines with --short
#         sh cues.sh ID         run one cue once and exit, e.g. sh cues.sh D4
#         sh cues.sh --help     this text
#
# Keys at each cue: Space runs it, b goes back one cue, s skips it, q quits.
# Ctrl-C at the key prompt does nothing; only q quits. Ctrl-C while a cue
# runs stops that command and returns here. Keys pressed while a cue runs
# are discarded. After the last cue only b (back to it) and q (quit) act.
#
# Run cues are the commands of demos.md, verbatim, except D2, which passes no
# GOROOT and lets the script resolve the pinned tree. Every cue exits 0. The
# warm-up builds the four program demos into ./bin, as demos.md's run order
# requires (build first, never go run them). From fish: sh cues.sh, unchanged.
# Run it from anywhere; it cd's to its own directory, the demos module.

set -eu
cd "$(dirname "$0")"

usage() {
	sed -n '11,25s/^# \{0,1\}//p' "$(basename "$0")"
}

short=0
list=0
present=0
only=
for a in "$@"; do
	case $a in
	--short) short=1 ;;
	--list) list=1 ;;
	--present) present=1 ;;
	-h | --help)
		usage
		exit 0
		;;
	-*)
		echo "cues.sh: unknown argument: $a" >&2
		usage >&2
		exit 2
		;;
	*) only=$a ;;
	esac
done

# Where the Expect lines come from. NO WSL CAPTURE EXISTS YET (DM-04,
# 2026-10-25): every Expect line below gives the output's shape and the
# machine-independent facts only. Replace them with the ranges from
# demos/captures/<date>-wsl once the owner's measure.sh batch has run, and
# fill the fallback lines (the presenter's words) at the same time.
STAMP='no measured ranges yet: shapes only, from the DM-04 reconstruction (2026-10-25); the WSL batch is demos/measure.sh'

# The cues, in lecture order (demos.md "Run order for the lecture"). Fields:
# id | lecture section | short (in the 60-minute deck) or - | the command |
# what to expect | the line to say if the run falls outside it.
# D1b, D6 and D7 run for minutes (-count=5 at the default 1 s benchtime: at
# least 50, 120 and 200 benchmark runs): start them, then talk.
CUES='D4|2, 6.3|-|./bin/fairness|GOMAXPROCS=N, 8 workers, 500ms each; a table of three locks (sync.Mutex, unfair TAS, fifo chan) with acquires, streak90/99/Max, starve90/99/Max, ns/op; fifo chan reads streak 1 and starve 7 at p90 and p99. Ranges: pending the WSL batch. Wall about 1.5 s|TBD after the WSL batch
D2|3|-|sh 02-sourcepath/run.sh|GOROOT source names the go1.27.1 toolchain tree; src/sync/mutex.go 66 lines, 0 starvation, 0 CompareAndSwap, 0 lockSlow; src/internal/sync/mutex.go 234 lines, 11, 5, 5; lines 55 and 150 for starvationThresholdNs; map.go lines 8 and 41, 115 lines; hashtriemap.go 724 lines (go1.27.1, not machine-dependent)|TBD after the WSL batch
D1a|4.3|-|sh 01-fastpath/dis.sh|four inlining-call lines; main.Crit STEXT size=107; LOCK CMPXCHGL at internal/sync/mutex.go:63, LOCK XADDL at :194, four XCHGL AX, AX; byte-identical to demos.md lines 217-259 at go1.27.1 amd64|TBD after the WSL batch
D1s|4.1|-|go test -v -run TestSizes ./01-fastpath|Mutex 8, RWMutex 24, Once 12, WaitGroup 16, Map 48, Cond 56, Pool 40 bytes; PASS (default build, linux/amd64; also unchanged under -race)|TBD after the WSL batch
D1b|4.3|-|go test -run=X -bench=. -count=5 -cpu=1,2 ./01-fastpath|50 benchmark lines (5 benchmarks x -cpu=1,2 x 5), about a minute: start it, then read the source. Ranges: pending the WSL batch|TBD after the WSL batch
D3|6.1, 6.2|-|./bin/statword|layout check ... OK; GOMAXPROCS=N NumCPU=N; eight blocks (2/0 ... 64/10000), each with locked, woken, starving percentages, transitions, max waiters, critical sections. Ranges: pending the WSL batch. Wall about 3.5 s|TBD after the WSL batch
D5|6.4|-|./bin/threshold|three tables (work=0, 200, 2000), rows tas, sync.Mutex, fifo chan: n=20000, p50, p90, p99, p99.9, max, >1ms, greedy ops. Ranges: pending the WSL batch|TBD after the WSL batch
D6|8.2|-|go test -run=X -bench=. -count=5 -cpu=2 ./06-rwmutex|120 benchmark lines (Mutex and RWMutex x work0/10/100/1000 x readonly, write1in100, write1in10), several minutes: start it, then talk. Ranges: pending the WSL batch|TBD after the WSL batch
D7|9.3, 9.4|-|go test -run=X -bench=. -count=5 -cpu=1,2 ./07-maps|200 benchmark lines (4 mixes x 5 structures x -cpu=1,2 x 5); the sharded rows are named Sharded<GOMAXPROCS*8>, Sharded160 at 20 CPUs. Several minutes: start it, then talk. Ranges: pending the WSL batch|TBD after the WSL batch
D8|10.1, 10.2|-|./bin/contention|part 1: wall clock, profile entries, contention events, goroutine-nanoseconds lost and its ratio to wall clock; part 2: three runs of padded/packed ms and their ratio. Ranges: pending the WSL batch. Wall about 1 s|TBD after the WSL batch'

if [ "$short" = 1 ]; then
	CUES=$(printf '%s\n' "$CUES" | grep '^[^|]*|[^|]*|short|' || :)
	if [ -z "$CUES" ]; then
		echo "cues.sh: no cue is marked short yet; the 60-minute deck's set is not chosen." >&2
		exit 2
	fi
fi
if [ -n "$only" ]; then
	CUES=$(printf '%s\n' "$CUES" | grep "^$only|" || :)
	if [ -z "$CUES" ]; then
		echo "cues.sh: no cue $only; sh cues.sh --list shows the ids." >&2
		exit 2
	fi
fi
ncues=$(printf '%s\n' "$CUES" | wc -l | tr -d ' ')

# show N: print cue N (a line number in $CUES) and set id, kind, cmd, expect, fallback.
show() {
	IFS='|' read -r id sect tag cmd expect fallback <<-EOF
	$(printf '%s\n' "$CUES" | sed -n "${1}p")
	EOF
	if [ "$present" = 1 ] && [ "$list" = 0 ]; then
		printf '\033[2J\033[H'
	else
		echo
	fi
	echo "== cue $id   [§$sect]   $1 of $ncues"
	echo "   \$ $cmd"
	[ "$present" = 1 ] && [ "$list" = 0 ] && return 0
	echo "   Expect: $expect"
	[ "$list" = 1 ] && echo "   Если не так: $fallback"
	return 0
}

if [ "$list" = 1 ]; then
	echo "Cues for L-04, from demos.md \"Run order for the lecture\". Run from this directory."
	echo "Expected ranges: $STAMP."
	i=1
	while [ "$i" -le "$ncues" ]; do
		show "$i"
		i=$((i + 1))
	done
	exit 0
fi

# The pin is enforced by the go directive in go.mod and go.work, not by the
# command name. On the WSL box the shell exports GOROOT=/usr/local/go
# (go1.25.3). With the go1.27.1 binary first on PATH that GOROOT wins: go env
# GOROOT, go list and go doc read the 1.25.3 tree, and a build fails with
#   compile: version "go1.25.3" does not match go tool version "go1.27.1"
# (docs/environment-notes.md §0.3). Drop a GOROOT that is not the pinned tree.
if [ -n "${GOROOT:-}" ] && [ "$(head -1 "$GOROOT/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "cues.sh: your shell exports GOROOT=$GOROOT, not the go1.27.1 tree; unsetting it for this run."
	unset GOROOT
fi
v=$(go env GOVERSION)
if [ "$v" != "go1.27.1" ]; then
	echo "cues.sh: go here is $v, want go1.27.1." >&2
	echo "cues.sh: check GOTOOLCHAIN (want auto) and that this is inside the course repo." >&2
	exit 1
fi
G=$(go env GOROOT)
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "cues.sh: go env GOROOT is $G, whose VERSION is not go1.27.1." >&2
	exit 1
fi

# A GODEBUG, GOMAXPROCS or GORACE left in the shell would leak into every cue
# (GORACE also changes the race runtime's one-second exit sleep).
for e in GODEBUG GOMAXPROCS GORACE; do
	if eval "[ -n \"\${$e+x}\" ]"; then
		echo "cues.sh: unsetting $e, which your shell exports; the cues set their own."
		unset "$e"
	fi
done

# runcue: run $cmd and report a non-zero status. A child of a shell that
# ignores INT inherits SIG_IGN, so INT gets a handler (trap :), not trap ''.
runcue() {
	rc=0
	eval "$cmd" || rc=$?
	case $rc in
	0) ;;
	130) echo "== stopped by Ctrl-C (exit status 130)" ;;
	*) echo "== exit status $rc" ;;
	esac
	if [ "$present" = 0 ]; then
		echo "   Если не так: $fallback"
	fi
}

if [ -n "$only" ]; then
	# One cue, no warm-up: the program cues still need ./bin (build first,
	# never go run them; the builds are cache hits after a full warm-up).
	mkdir -p bin
	case $(printf '%s\n' "$CUES" | cut -d'|' -f4) in
	./bin/*)
		go build -o bin/statword ./03-statword && go build -o bin/fairness ./04-fairness &&
			go build -o bin/threshold ./05-threshold && go build -o bin/contention ./08-contention
		;;
	esac
	trap : INT
	show 1
	runcue
	exit 0
fi

if [ ! -t 0 ]; then
	echo "cues.sh: needs a terminal for the key prompt; use --list to print the cues." >&2
	exit 1
fi

tmp=$(mktemp -d "${TMPDIR:-/tmp}/cues.XXXXXX")
saved=$(stty -g)
trap 'stty "$saved" 2>/dev/null; rm -rf "$tmp"' EXIT
trap 'exit 129' HUP
trap 'exit 143' TERM

# Warm the build cache, so the first go run, go run -race and go test are not
# slow on screen. vet compiles every package, test-only ones included; the
# -race builds and the test links are separate cache entries.
echo "== warming the build cache with $v, $(go env GOOS)/$(go env GOARCH)"
mkdir -p bin
if ! {
	go vet ./... &&
		go build -o bin/statword ./03-statword &&
		go build -o bin/fairness ./04-fairness &&
		go build -o bin/threshold ./05-threshold &&
		go build -o bin/contention ./08-contention &&
		go test -c -o "$tmp/01.test" ./01-fastpath &&
		go test -c -o "$tmp/06.test" ./06-rwmutex &&
		go test -c -o "$tmp/07.test" ./07-maps &&
		go test -run TestSizes ./01-fastpath
} >"$tmp/warm.log" 2>&1; then
	cat "$tmp/warm.log" >&2
	echo "cues.sh: the warm-up failed." >&2
	exit 1
fi
echo "== ready"
echo "== keys: Space run, b back, s skip, q quit (Ctrl-C at the prompt does nothing)"

# key: wait for one keypress, without echo or Enter, and set k to it. Keys
# pressed while a cue ran are already buffered: drain them first (min 0
# time 0 makes the read return at once), so a stray Space cannot start the
# next cue.
key() {
	stty raw -echo min 0 time 0
	dd bs=256 count=1 >/dev/null 2>&1 || :
	stty raw -echo min 1 time 0
	k=$(dd bs=1 count=1 2>/dev/null)
	stty "$saved"
}

trap : INT

i=1
while :; do
	if [ "$i" -gt "$ncues" ]; then
		echo
		# Only b and q leave the end prompt: a stray key, Ctrl-C included,
		# shows it again rather than ending the runner after the last cue.
		while :; do
			echo "== end of cues. b goes back to the last one; q quits."
			key
			case $k in
			b)
				i=$ncues
				break
				;;
			q) exit 0 ;;
			esac
		done
		continue
	fi
	show "$i"
	echo "   [Space] run  [b] back  [s] skip  [q] quit"
	key
	case $k in
	' ') ;;
	b)
		[ "$i" -gt 1 ] && i=$((i - 1))
		continue
		;;
	s)
		i=$((i + 1))
		continue
		;;
	q)
		echo "== quit at cue $id."
		exit 0
		;;
	*) continue ;; # Ctrl-C (\003) and any other key: show the cue again
	esac
	runcue
	i=$((i + 1))
done
