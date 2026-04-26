#! /bin/sh

target=git-ai-commit
go build -o ${target} .
# cp ${target} .git/hooks/prepare-commit-msg
# cp ${target} ~/gopath/bin
# chmod +x .git/hooks/prepare-commit-msg
