#!/bin/bash
echo "## $0 received NUM ARGS : " $#
if [[ $# -lt 1 ]]; then
  echo "## 💥💥 Usage: $0 <executable> [env_file] [extra args...]"
  exit 1
fi
BIN_FILENAME=${1}
shift
ENV_FILENAME='.env'
if [[ $# -ge 1 && "${1}" == *.env* ]]; then
  ENV_FILENAME=${1}
  shift
fi
# Any remaining arguments will be forwarded to the executable
echo "## will try to run : ${BIN_FILENAME} with env variables in ${ENV_FILENAME} and extra args: $*"
if [[ -r "$ENV_FILENAME" ]]; then
  if [[ -x "$BIN_FILENAME" ]]; then
    echo "## will execute $BIN_FILENAME $*"
    set -a
    source <(sed -e '/^#/d;/^\s*$/d' -e "s/'/'\\\\''/g" -e "s/=\(.*\)/='\1'/g" "$ENV_FILENAME" )
    "${BIN_FILENAME}" "$@"
    set +a
  else
    echo "## 💥💥 expecting first argument to be an executable path"
    exit 1
  fi
else
  echo "## 💥💥 env path argument : ${ENV_FILENAME} was not found"
  exit 1
fi
