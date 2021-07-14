#!/bin/sh
set -e

SED="sed"
if which gsed >/dev/null 2>&1; then
	SED="gsed"
fi

rm -rf ./docs/*.md

git checkout -- go.*
go mod edit -replace github.com/spf13/cobra=github.com/caarlos0/cobra@completions-md
go mod tidy
go run . docs

"$SED" \
	-i'' \
	-e 's/SEE ALSO/See also/g' \
	-e 's/^## /# /g' \
	-e 's/^### /## /g' \
	-e 's/^#### /### /g' \
	-e 's/^##### /#### /g' \
	./docs/*.md


git checkout -- go.*
