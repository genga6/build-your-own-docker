#!/usr/bin/env bash
# Alpine の mini rootfs をダウンロードして ./rootfs に展開する。
#   rootfs/ は .gitignore 済み（コミットしない）。clone 後はこのスクリプトで用意する。
#
# 使い方:
#   cd container && ./fetch-rootfs.sh
set -euo pipefail

ARCH="${ARCH:-x86_64}"   # 必要なら ARCH=aarch64 ./fetch-rootfs.sh
BASE="https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/${ARCH}"
DEST="$(dirname "$0")/rootfs"

echo "==> latest-releases.yaml から minirootfs のファイル名を取得"
FILE="$(curl -fsSL "${BASE}/latest-releases.yaml" \
  | grep -m1 -oE "alpine-minirootfs-[0-9.]+-${ARCH}\.tar\.gz")"
if [ -z "${FILE}" ]; then
  echo "ファイル名の取得に失敗しました" >&2
  exit 1
fi
echo "    file = ${FILE}"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "==> ダウンロード"
curl -fsSL -o "${TMP}/alpine.tar.gz" "${BASE}/${FILE}"

echo "==> ${DEST} に展開"
rm -rf "${DEST}"
mkdir -p "${DEST}"
tar -xzf "${TMP}/alpine.tar.gz" -C "${DEST}"

echo "==> 完了: $(ls "${DEST}" | tr '\n' ' ')"
