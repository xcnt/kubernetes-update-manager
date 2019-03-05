#!/usr/bin/env bash

find . -type f -name common_test.go -not -path "./vendor/*" -exec bash -c "cd \$(dirname '{}') && go test -coverprofile=cover.out -gocheck.vv |  go2xunit -gocheck -output xunit.xml" \;

workdir=`pwd`
profile="$workdir/cover.out"

mode=set
echo "mode: $mode" > "$profile"
find . -type f -name cover.out -not -path "./vendor/*" -not -path "./cover.out" -exec bash -c "grep -h -v \"^mode:\" \"{}\" >>\"$profile\"" \;
gocover-cobertura < $profile > "$workdir/coverage.xml"
