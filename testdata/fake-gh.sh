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
echo https://github.com/foreverfl/first/pull/1
