#!/bin/bash

usage="usage: $(basename "$0") [-d] [-h] [-v]

Download latest version of atomic-red-team atomic tests from the given
repository owner and branch (defaults to redcanaryco/master). This script
assumes the repository name is 'atomic-red-team'.

This script will exit early if the include/atomics directory already exists
unless the 'force' option is provided.

where:
    -f      force download of latest atomics
    -h      show this help text
    -r      repo owner and branch to download from"


repo=redcanaryco/master
force=false


# loop through positional options/arguments
while getopts ':fhr:' option; do
    case "$option" in
        f)  force=true             ;;
        h)  echo -e "$usage"; exit ;;
        r)  repo="$OPTARG"         ;;
        \?) echo -e "illegal option: -$OPTARG\n" >$2
            echo -e "$usage" >&2
            exit 1 ;;
    esac
done

IFS='/' read -ra repo <<< "$repo"

if (( ${#repo[@]} != 2 )); then
  echo "invalid repo provided"
  exit 1
fi

[ -d include/atomics ] && [ "$force" = false ] && echo "\
  Atomics already exist - not overwriting.
  Delete include/atomics directory and rerun if you want to reinstall." \
  && exit 0

url=https://github.com/${repo[0]}/atomic-red-team/archive/${repo[1]}.zip

echo "Downloading repo archive from $url"

curl -L -o art.zip $url

echo "Unarchiving repo"

unzip art.zip
rm art.zip

echo "Copying archived atomics to 'include' directory"

cp -a atomic-red-team-${repo[1]}/atomics include

echo "Deleting unarchived repo"

rm -rf atomic-red-team-${repo[1]}
