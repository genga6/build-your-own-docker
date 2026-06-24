# 実装ノート / 環境メモ

Go でコンテナを自作する過程のメモ。**学習用かつ、あとで見返すリファレンス**。

## 進捗（ステップ別）

| ステップ | 内容 | Codespaces | 特権環境 |
| --- | --- | --- | --- |
| 0 | `run`/`child` の自己再実行ディスパッチ | ✅ | ✅ |
| 1 | 名前空間で隔離（user/uts/pid/mount + UID/GIDマッピング） | ✅ | ✅ |
| 2 | `chroot` でファイルシステム隔離（Alpine rootfs） | ✅ | ✅ |
| 3 | `/proc` を `mount` | ❌（mount不可） | ✅ |
| 4 | `pivot_root` | ❌（mount不可） | ✅ |
| 5 | cgroups v2 でリソース制限 | ❌（cgroup作成不可） | ✅ |

## ⚠️ Codespaces / dev container の制約（重要）

このリポジトリの Codespaces 環境では **`mount()` と cgroup 作成ができない**。調査で確定した原因:

```
$ grep -i seccomp /proc/self/status   → Seccomp: 0          # seccompは無効（犯人ではない）
$ capsh --print                       → !cap_sys_admin      # mount に必要な権限が無い
                                        cap_sys_chroot あり  # だから chroot は動く
$ cat /proc/self/attr/current         → docker-default (enforce)  # AppArmorが mount を禁止
$ stat -fc %T /sys/fs/cgroup          → cgroup2fs（root は read-only）
```

理由は 2 つ重なっている:

1. **CAP_SYS_ADMIN が無い** — `mount(2)` に必須。Codespaces のコンテナに付与されていない。
   user namespace を作っても、土台のコンテナがこの権限を持たないので超えられない。
2. **AppArmor `docker-default` (enforce)** — capability とは別に mount 操作そのものを禁止。

`chroot` が動くのは `cap_sys_chroot` だけは許可されているから。

→ **名前空間（ステップ0-2）は Codespaces でも動く。mount/cgroup（ステップ3以降）は特権環境が必要。**

### 名前空間まわりの前提（記事との違い）

記事は実 root を前提に `CLONE_NEWUSER` 無しで書かれているが、コンテナの中（Codespaces 等）では
直接の名前空間作成が `Operation not permitted` になる。そこで **`CLONE_NEWUSER` を最初に作り**、
その中で root になってから他の名前空間を作る **rootless 方式**にしている（`SysProcAttr` の
`Cloneflags` に `CLONE_NEWUSER` と `UidMappings`/`GidMappings` を入れているのはこのため）。
実マシンで素の root なら user namespace 無しでも動く。

## ローカル（特権環境）での動かし方

ステップ3以降（mount / cgroup）を実演するには、特権付きの Linux 環境が要る。

### 方法 A: ローカル VS Code の Dev Container を特権化

`.devcontainer/devcontainer.json` に以下を足してリビルド（**Codespaces では効かない可能性が高い。ローカルの Docker で**）:

```jsonc
{
  // 既存の設定に追記
  "runArgs": ["--privileged"]
  // よりピンポイントにするなら:
  // "runArgs": ["--cap-add=SYS_ADMIN", "--security-opt", "apparmor=unconfined", "--security-opt", "seccomp=unconfined"]
}
```

`--privileged` で CAP_SYS_ADMIN が付き、AppArmor の拘束も外れて `mount` が通る。
cgroup v2 への書き込みも可能になる（cgroup namespace 委譲込み）。

### 方法 B: 普通の Linux VM / 実マシン

`sudo` できる Linux なら、そのまま動く。実 root であれば `CLONE_NEWUSER` 無しでも可。

## セットアップ手順（clone 後 / 環境を移したとき）

```bash
cd container

# 1. Alpine rootfs を用意（rootfs/ は .gitignore 済みなので毎回必要）
./fetch-rootfs.sh

# 2. 動作確認
go run . run /bin/ls /              # → Alpine の中身が見える（chroot）
go run . run /bin/sh               # → コンテナのシェルに入る
go run . run /bin/cat /etc/os-release   # → Alpine Linux

# 3. 特権環境なら（mount 後）
go run . run /bin/ps aux           # → コンテナ内プロセスだけが見える
```
