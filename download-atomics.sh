#!/bin/bash

repo=redcanaryco/master

if [[ -n $1 ]]; then
  repo=$1
fi

IFS='/' read -ra repo <<< "$repo"

if (( ${#repo[@]} != 2 )); then
  echo "invalid repo provided"
  exit 1
fi

[ -d include/atomics ] && echo "\
  Atomics already exist - not overwriting.
  Delete include/atomics directory and rerun if you want to reinstall." \
  && exit 0

url=https://github.com/${repo[0]}/atomic-red-team/archive/${repo[1]}.zip

echo "Downloading repo archive from $url"

curl -L -o art.zip $url

echo "Unarchiving repo"

unzip art.zip && rm art.zip

echo "Copying archived atomics to 'include' directory"

cp -a atomic-red-team-${repo[1]}/atomics include

echo "Deleting unarchived repo"

rm -rf atomic-red-team-${repo[1]}
