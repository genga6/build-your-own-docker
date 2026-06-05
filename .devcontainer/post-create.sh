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

# lefthook が devDependencies に入っていれば git hooks を登録する。
# プロジェクト未作成 / lefthook 未導入の場合はスキップ。
setup_lefthook() {
  log "lefthook: start"
  if [ ! -f package.json ]; then
    log "lefthook: skip (no package.json)"
    return
  fi
  if [ ! -f lefthook.yml ] && [ ! -f lefthook.yaml ] && [ ! -f .lefthook.yml ]; then
    log "lefthook: skip (no lefthook config)"
    return
  fi
  if ! pnpm exec lefthook --version >/dev/null 2>&1; then
    log "lefthook: skip (binary not found — add lefthook to devDependencies)"
    return
  fi
  pnpm exec lefthook install
  log "lefthook: done"
}

# セットアップ処理をまとめて実行
main() {
  install_claude
  setup_lefthook
}

main "$@"
