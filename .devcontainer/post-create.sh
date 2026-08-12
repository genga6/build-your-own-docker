#!/usr/bin/env bash
set -euo pipefail

log() { echo "[$(date +%H:%M:%S)] $*"; }

# Claude Code CLI を導入（既に入っていればスキップ）
install_claude() {
  log "claude: start"
  if command -v claude >/dev/null 2>&1; then
    log "claude: already installed"
    return
  fi
  curl -fsSL https://claude.ai/install.sh | bash
  log "claude: done"
}

main() {
  install_claude
}

main "$@"
