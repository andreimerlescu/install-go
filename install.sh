#!/bin/bash

set -e  # BEST PRACTICES: Exit immediately if a command exits with a non-zero status.
[ "${DEBUG:-0}" == "1" ] && set -x  # DEVELOPER EXPERIENCE: Enable debug mode, printing each command before it's executed.
set -u  # SECURITY: Exit if an unset variable is used to prevent potential security risks.
set -C  # SECURITY: Prevent existing files from being overwritten using the '>' operator.
[ "${VERBOSE:-0}" == "1" ] && set -v  # DEVELOPER EXPERIENCE: Print shell input lines as they are read, aiding in debugging.

safe_exit() {
  local msg="${1:-UnexpectedError}"
  echo "${msg}"
  exit 1
}

# No Windows support
[ "$(uname)" != "Linux" ] && safe_exit "HOLD UP: igo is only supported on Linux."

GODIR="${HOME:-"/home/$(whoami)"}/go" # where all things go go

# Create the Go Directory
mkdir -p "${GODIR}/backups" "${GODIR}/manifests" "${GODIR}/scripts" "${GODIR}/shims"

download_script() {
  local src="${1}"
  local pkg="${2}"
  local file="${3}"

  if [ ! -f "${pkg}" ] || [ ! -L "${pkg}" ]; then
    if [ -f "${GODIR}/install-go/${file}" ]; then
      [ -L "${pkg}" ] || ln -s "${GODIR}/install-go/${file}" "${pkg}"
    else
      wget --https-only --secure-protocol=auto --no-cache -O "$pkg" "$src" < /dev/null > /dev/null 2>&1 || echo "Downloading ${src}: FAILED, using local if available."
    fi
  else
    echo "Already have $(basename "${pkg}") installed at ${pkg}!"
  fi
}

download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/install_go.sh" "${GODIR}/scripts/igo" "install_go.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/switch_go.sh" "${GODIR}/scripts/sgo" "switch_go.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/backup_go.sh" "${GODIR}/scripts/bgo" "backup_go.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/remove_go.sh" "${GODIR}/scripts/rgo" "remove_go.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/functions.sh" "${GODIR}/scripts/functions.sh" "functions.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/shim_go.sh" "${GODIR}/shims/go" "shim_go.sh"
download_script "https://raw.githubusercontent.com/andreiwashere/install-go/main/shim_gofmt.sh" "${GODIR}/shims/gofmt" "shim_gofmt.sh"

chmod +x "${GODIR}/shims/go"
chmod +x "${GODIR}/shims/gofmt"
chmod +x "${GODIR}/scripts/bgo"
chmod +x "${GODIR}/scripts/igo"
chmod +x "${GODIR}/scripts/rgo"
chmod +x "${GODIR}/scripts/sgo"

current_shell=$(basename "${SHELL:-"/bin/bash"}")

case "$current_shell" in
  bash)
    rcfile="${HOME}/.bashrc"
    ;;
  zsh)
    rcfile="${HOME}/.zshrc"
    ;;
  *)
    echo "Unsupported shell: ${current_shell}."
    exit 1
    ;;
esac

if [ -f "$rcfile" ]; then
  ! grep -qxF "export PATH=${GODIR}/scripts:${GODIR}/shims" "$rcfile" && echo "export PATH=${GODIR}/scripts:${GODIR}/shims:\$PATH" >> "$rcfile"
  set +u
  # shellcheck disable=SC1090
  source "${rcfile}"
  set -u
  echo "Sourced $rcfile."
else
  echo "$rcfile not found."
fi

