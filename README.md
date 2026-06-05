# nextjs-lab

Next.js の検証用リポジトリ。devcontainer 上で動かす前提で構成している。

## スタック

- **Next.js 16** / App Router
- **React 19** + **React Compiler**（`next.config.ts` の `reactCompiler: true`）
- **TypeScript 5**
- **Tailwind CSS v4**
- **Biome 2**（lint / format / import 整列。ESLint・Prettier は使わない）
- **pnpm**（corepack 経由）

## 開発環境（devcontainer）

`.devcontainer/` を使う前提。VS Code / Codespaces で「Reopen in Container」すると以下が自動で行われる。

- corepack で pnpm を有効化（`onCreateCommand`）
- `package.json` があれば依存をインストール（`updateContentCommand`）
- Claude Code CLI の導入、lefthook があれば git hooks 登録（`post-create.sh`）
- 起動時に dev サーバを立ち上げ、3000 番ポートを転送

エディタは Biome を既定フォーマッタにし、保存時に自動修正・import 整列を実行する。

## セットアップ（ローカルで動かす場合）

```bash
corepack enable
pnpm install
pnpm dev
```

[http://localhost:3000](http://localhost:3000) を開く。`app/page.tsx` を編集すると自動でリロードされる。

## コマンド

| 用途 | コマンド |
| --- | --- |
| 開発サーバ起動 | `pnpm dev` |
| 本番ビルド | `pnpm build` |
| 本番サーバ起動 | `pnpm start` |
| Lint（チェックのみ） | `pnpm lint` |
| フォーマット適用 | `pnpm format` |
| Lint + 自動修正 | `pnpm exec biome check --write` |

## コーディングエージェント向け

作業ガイドは [AGENTS.md](./AGENTS.md) を参照（`CLAUDE.md` もこれを参照している）。
