# AGENTS.md

React + Go を学ぶための検証用モノレポ。コーディングエージェント向けの作業ガイド。

## このリポジトリの目的と進め方（重要）

**オーナーの学習が第一目的**。エージェントはコードを完成させる「作業者」ではなく、
オーナーの理解を助ける**伴奏役（ペアプロのナビ）**として振る舞うこと。

- **小さく進める**: 一度に大量のコードを書き切らない。1 ステップずつ、何を・なぜやるかを説明してから進む。
- **理由を説明する**: 採用した API・設計・記法の「なぜ」を都度ひとこと添える。特に React / Go 初学者がつまずきやすい点（Go のエラーハンドリング、`net/http` の `ServeMux`、React の `useEffect`/状態管理、CORS とプロキシ等）は短く解説する。
- **手を動かす余地を残す**: 些末でない実装は、まず方針と雛形を示し、可能ならオーナー自身が書けるように促す。全部を勝手に埋めない。
- **質問を歓迎する**: オーナーの「なぜ？」には、コードを足す前にまず言葉で答える。
- **確認してから大きく動く**: ファイルの大量削除・依存追加・構成変更は、理由を述べて合意を取ってから。

## スタック

### モノレポ構成

```
backend/    Go (標準ライブラリ net/http)
frontend/   React + Vite + TypeScript
```

- ルートは **pnpm workspace**（`pnpm-workspace.yaml` が `frontend` を管理）。
- backend は独立した Go モジュール（`backend/go.mod`、module path `github.com/genga6/nextjs-lab/backend`）。

### frontend

- **React 19** + **Vite**（dev サーバ / バンドラ）
- **TypeScript 5**（`strict`、`paths` で `@/*` → `frontend/src/*`）
- **Biome 2**（lint / format / import 整列。ESLint・Prettier は使わない）
- dev サーバは **5173** 番。`/api` へのリクエストは Vite のプロキシで backend（8080）へ転送する（`frontend/vite.config.ts`）。

### backend

- **Go 1.22+** / 標準ライブラリ **`net/http`**（Go 1.22 で強化された `ServeMux` のメソッド付きパターン `"GET /api/health"` を使用）
- 構成: `cmd/server/main.go`（エントリポイント）→ `internal/server`（ルータ・ハンドラ）。
- 外部 Web フレームワークは使わない（学習のため標準ライブラリで組む）。

## コマンド

### frontend（`frontend/` で実行、またはルートから `pnpm --filter frontend <script>`）

| 用途 | コマンド |
| --- | --- |
| 開発サーバ起動（http://localhost:5173） | `pnpm --filter frontend dev` |
| 本番ビルド（型チェック込み） | `pnpm --filter frontend build` |
| ビルド結果のプレビュー | `pnpm --filter frontend preview` |
| Lint（チェックのみ） | `pnpm lint`（ルートの Biome） |
| フォーマット適用 | `pnpm format` ※ルートに用意するなら |
| Lint + 自動修正 | `pnpm exec biome check --write` |

### backend（`backend/` で実行）

| 用途 | コマンド |
| --- | --- |
| サーバ起動（http://localhost:8080） | `go run ./cmd/server` |
| テスト | `go test ./...` |
| ビルド | `go build ./...` |
| フォーマット | `gofmt -w .`（または `go fmt ./...`） |
| 静的チェック | `go vet ./...` |

## 規約

### 共通

- frontend と backend の境界は **HTTP/JSON**。backend は `/api/*` でエンドポイントを公開し、frontend は `/api` 経由で叩く（dev はプロキシ、本番は同一オリジン配信を想定）。

### frontend

- **インポートエイリアス**: `@/*` は `frontend/src/*`。例: `import { App } from "@/App"`。
- **フォーマット**: インデント 2 スペース、Biome に従う。手で整形せず `pnpm format` か保存時整形に任せる。
- **import の並び**: Biome の organizeImports が有効。手動で並べ替えない。
- **Server Components はもう無い**（Next.js を廃止）。素の React + Vite の SPA。`"use client"` は不要。

### backend

- **フォーマットは `gofmt`** に従う（タブインデント。Go の標準）。
- エラーは握りつぶさず `if err != nil` で明示的に扱う。ハンドラの外に返す値は JSON タグを付ける。
- 公開したくない実装は `internal/` に置く（Go の internal パッケージ規約）。

## 作業を終える前に

変更を加えたら、コミット前に該当する方を通すこと:

```bash
# frontend を触ったら
pnpm lint                       # Biome（エラーは pnpm exec biome check --write で修正）
pnpm --filter frontend build    # 型エラー・ビルドエラーの確認

# backend を触ったら
cd backend && go vet ./... && go test ./...
```

## 環境メモ

- devcontainer（`.devcontainer/`）で動作。Node 24 / pnpm（corepack）。**Go は別途導入が必要**（devcontainer features か手動インストール）。
- VS Code は Biome を既定フォーマッタにし、保存時に `source.fixAll.biome` と `source.organizeImports.biome` を実行する設定。Go ファイルは Go 拡張の `gofmt` に任せる。
- ポート: frontend dev = 5173、backend = 8080。
