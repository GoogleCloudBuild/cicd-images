# shellcheck disable=SC2148
# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PACKAGES="$(dirname "$(readlink $0 -f)")/packages.yaml"
SRC=${SRC:-/src}
TOOLS_DIR="/opt/tools"

export PACKAGES SRC TOOLS_DIR

err() {
  echo "ERROR [$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" 1>&2
}
err_exit() {
  echo "ERROR [$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" 1>&2
  exit 1
}

# Copies source files to a destination
# First param is destination
# therefore it expects at least 2 param
copy_sources(){
  if [[ "$#" -lt 2 ]]; then
    err_exit "copy_sources expect at least 2 params given $#"
  fi
  # first param is a
  destination=$1
  if [[ ! "$destination" =~ ^/.* ]]; then
    err_exit "copy_sources expects destination to be absolute path $destination"
  fi
  shift
  (cd $SRC || err_exit "Unable to switch pwd to $SRC" ;
   cp -p "$@" $destination)

}


update_env(){
  if [[ $# != 2 ]]; then
    err "update_env: {variable_to_update} {new_value}"
    return 1
  fi
  env_key=$1
  value_to_set=$2
  if [[ -n "$value_to_set" ]]; then
    export $env_key="$value_to_set"
    # Update environment file
    sed -i "/^$env_key=/{h;s#=.*#=${!env_key}#};\${x;/^\$/{s##${env_key}=${!env_key}#;H};x}" \
        /etc/environment
  fi
  source /etc/environment
}

is_version_supported(){
  local version=$1
  local key=$2
  SUPPORTED=$(yq ".${key}.supported-versions" <$PACKAGES)
  is_supported=$(echo "$SUPPORTED" | grep -c "$version" || true)
  if [[ "$is_supported" -ne "1" ]]; then
    # Version is not found in 'supported-versions'
    return 1
  fi
  # Version is support
  return 0
}


copy_licenses(){
  if [[ $# != 1 ]] || [[ ! -d $1 ]] ; then
    err_exit "copy_licenses: {INSTALL_DIR} is required"
  fi

  install_dir=$1
  tool_name=$(basename $install_dir)
  mkdir -p /THIRD_PARTY_NOTICES/$tool_name;
  find $install_dir -type f \( -iname '*license*' -o -iname '*notice*' \) | \
  while read file
  do
    target_name=$(echo $file|tr '/' '_')
    cp $file /THIRD_PARTY_NOTICES/${tool_name}/${target_name}
  done
}
