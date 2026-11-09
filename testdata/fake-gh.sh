#!/bin/sh
echo "$@" >> "$FAKE_GH_CALLS"
if [ "$2" = "view" ]; then
  state=$(grep "^$3 " "$FAKE_GH_STATES" | cut -d' ' -f2)
  if [ -n "$state" ]; then
    echo "{\"state\":\"$state\",\"url\":\"https://github.com/foreverfl/first/pull/1\"}"
    exit 0
  fi
  echo 'no pull requests found' >&2
  exit 1
fi
# like real gh, a PR just created reads as OPEN on the next view; gh runs in the PR's worktree
if [ "$1" = "pr" ] && [ "$2" = "create" ]; then
  echo "$(git branch --show-current) OPEN" >> "$FAKE_GH_STATES"
fi
echo https://github.com/foreverfl/first/pull/1
