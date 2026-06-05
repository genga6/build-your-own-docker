# AGENTS.md

Next.js の検証用リポジトリ（nextjs-lab）。コーディングエージェント向けの作業ガイド。

## スタック

- **Next.js 16** / App Router（`app/` ディレクトリ）
- **React 19** + **React Compiler**（`next.config.ts` の `reactCompiler: true`、`babel-plugin-react-compiler`）
- **TypeScript 5**
- **Tailwind CSS v4**（`@tailwindcss/postcss`、設定は `app/globals.css` の `@import "tailwindcss"`）
- **Biome 2**（lint / format / import 整列。ESLint・Prettier は使わない）
- **pnpm**（パッケージマネージャ）

## コマンド

| 用途 | コマンド |
| --- | --- |
| 開発サーバ起動（http://localhost:3000） | `pnpm dev` |
| 本番ビルド | `pnpm build` |
| 本番サーバ起動 | `pnpm start` |
| Lint（チェックのみ） | `pnpm lint` |
| フォーマット適用 | `pnpm format` |
| Lint + 自動修正 | `pnpm exec biome check --write` |

## 規約

- **インポートエイリアス**: `@/*` はリポジトリルート（`./*`）を指す。例: `import { x } from "@/app/lib/x"`。`src/` ディレクトリは使っていない。
- **フォーマット**: インデントは 2 スペース。Biome に従う。手で整形せず `pnpm format` か保存時整形に任せる。
- **import の並び**: Biome の organizeImports が有効。手動で並べ替えない。
- **React Compiler が有効**: 不要な `useMemo` / `useCallback` / `React.memo` を新規に足さない（コンパイラが最適化する）。既存コードを壊さない範囲で従う。
- **Server Components が既定**: クライアント側の状態・イベントが必要なときだけファイル先頭に `"use client"` を付ける。
- **Lint ルール**: Biome recommended ＋ `next` / `react` ドメインを有効化済み。

## 作業を終える前に

変更を加えたら、コミット前に必ず以下を通すこと:

```bash
pnpm lint      # biome check（エラーが出たら pnpm exec biome check --write で修正）
pnpm build     # 型エラー・ビルドエラーの確認
```

## 環境メモ

- devcontainer（`.devcontainer/`）で動作。Node 24 / pnpm（corepack）。
- VS Code は Biome を既定フォーマッタにし、保存時に `source.fixAll.biome` と `source.organizeImports.biome` を実行する設定。
- dev サーバはコンテナの 3000 番ポートを転送して使う（`WATCHPACK_POLLING=true` で Fast Refresh 対応）。
