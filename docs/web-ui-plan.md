# Web UI 構想（コンテナの動きをブラウザから観察する）

`container/`（自作コンテナ = byod）に Web UI を被せて、**コンテナの隔離を可視化する**ための構想メモ。
今すぐ作るのではなく、ローカル環境移行後に着手する想定の**設計の置き場**。

## なぜやると良いか：本物の Docker と同じ構造になる

```
このリポジトリで作るもの        本物の Docker
┌──────────────┐            ┌──────────────┐
│ React UI      │            │ docker CLI /   │
│ (一覧・起動)   │            │ Docker Desktop │  ← クライアント
└──────┬───────┘            └──────┬───────┘
       │ HTTP/JSON (/api)           │ HTTP API
┌──────▼───────┐            ┌──────▼───────┐
│ Go backend    │            │ dockerd        │  ← デーモン
│ = byod を起動  │            │ (containerd)   │
└──────┬───────┘            └──────┬───────┘
       │ clone/chroot…             │ clone/chroot…
   コンテナ                    コンテナ
```

- `frontend/`（React + Vite, 既存）= クライアント
- `backend/`（Go net/http, AGENTS.md が予告。未作成）= デーモン役
- `container/`（byod 本体, 実装中）= ランタイム

→ Docker が「デーモン + クライアント + ランタイム」でできていることを、自作で体感できる。

## UI で見せたいもの

| パネル | 内容 | Codespaces |
| --- | --- | --- |
| ホスト視点 vs コンテナ視点 | hostname / PID / whoami / `ls /` を左右に並べる。コンテナ側は `container` / PID=1 / Alpine | ✅ |
| 名前空間の ID 比較 | `/proc/<pid>/ns/*` の inode 番号を表で。作った ns(uts/pid/mnt/user) は番号が違い、作ってない ns(net 等) は同じ | ✅ |
| プロセスツリー | ホストからは「ただの子プロセス(PID 30123)」、中からは「PID 1」 | ✅ |
| コマンド実行 & 出力 | フォームから `run /bin/ls /` 等を実行して結果表示 | ✅ |
| リソース制限 (cgroup) | プロセス数上限・fork ボム防止の実演 | ❌ 特権環境のみ |

**一番おいしい「隔離の可視化」は Codespaces でも動く**（mount/cgroup 以外）のがポイント。

### 「名前空間 ID 比較」が特に効く

各プロセスには `/proc/<pid>/ns/uts` のようなリンクがあり、中身は `uts:[4026531838]` という inode 番号
（＝どの名前空間に属すかの ID）。Go なら `os.Readlink("/proc/<pid>/ns/uts")` で読める。

```
              host shell        container
uts   :       4026531838        4026532301   ← 違う（作ったから）
pid   :       4026531836        4026532303   ← 違う
mnt   :       4026531840        4026532299   ← 違う
user  :       4026531837        4026532298   ← 違う
net   :       4026531999        4026531999   ← 同じ（作ってない）
```

これを並べると「名前空間って結局ただの ID（inode）の振り分けなんだ」が視覚的に伝わる。

## 最小の第一歩（刻み方）

1. **Go backend を作る**（`backend/`、net/http）。エンドポイント 1 個
   `POST /api/run` → body の `{cmd, args}` で `byod run` 相当を実行し、stdout/stderr を JSON で返す。
2. **frontend** にフォーム 1 個（コマンド入力 + 実行ボタン）と結果表示エリア。
   Vite のプロキシ（`/api` → 8080）で繋ぐ。
3. 動いたら「ホスト視点 vs コンテナ視点」パネル、続いて「名前空間 ID 表」を足す。

この過程で React の `useState`/`fetch`、Go の `net/http`/`os/exec`、CORS/プロキシ といった
AGENTS.md が挙げる学習ポイントを自然に踏める。

## 注意 / 前提

- backend が byod を起動するので、byod のバイナリ化（`go build -o byod .`）か、`go run` の呼び出し方を決める。
- 名前空間の作成は Codespaces でも動くが、cgroup/mount 系は特権環境が必要（`container/NOTES.md` 参照）。
- `backend/` は独立 Go モジュール（module path は `github.com/genga6/build-your-own-docker/backend` を想定。AGENTS.md 準拠）。
