#!/usr/bin/env bash

find . -type f -name common_test.go -not -path "./vendor/*" -exec bash -c "cd \$(dirname '{}') && go test -gocheck.vv |  go2xunit -gocheck -output xunit.xml" \;
