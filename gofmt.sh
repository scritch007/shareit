#!/bin/bash
for f in $(find . -name "*.go")
do
  echo "Running command on $(dirname $f)"
  go fmt $(dirname $f)
done
