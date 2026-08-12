# 実装ノート / 学習ロードマップ

Go でコンテナを自作する過程のメモ。**学習用かつ、あとで見返すリファレンス**。

このリポジトリの目的は2つある。片方は環境に強く依存するが、もう片方は依存しない:

- **A. 隔離のしくみを自作する** — namespaces / chroot / mount / cgroups / capabilities / seccomp
  → 一部は特権が必要（後述）
- **B. image と依存関係の実体を知る** — registry / manifest / layer / config
  → **特権が一切不要。ただの HTTP + tar + JSON。今の環境で全部できる**

---

## 学習ロードマップ

### A. 隔離のしくみ

| # | 柱 | 内容 | 状態 | Codespaces |
| --- | --- | --- | --- | --- |
| 0 | 自己再実行 | `run`/`child` ディスパッチ（`/proc/self/exe`） | 済 | ✅ |
| 1 | namespaces | user/uts/pid/mount + UID/GIDマッピング | 済 | ✅ |
| 2 | chroot | ファイルシステム隔離（Alpine rootfs） | 済 | ✅ |
| 3 | mount | `/proc` を mount して `ps` を動かす | コード済／**動かない** | ❌ |
| 4 | pivot_root | chroot を捨てて本物の root 切り替えへ | 未 | ❌ |
| 5 | cgroups v2 | メモリ/CPU の制限 | 未 | ❌ |
| 6 | **capabilities** | bounding set を削って「root だが root でない」を作る | 未 | ✅ |
| 7 | **seccomp** | syscall フィルタ（Docker の隔離のもう一本の柱） | 未 | ✅ |
| 8 | **network ns** | `CLONE_NEWNET` でネットワークを隔離 | 未 | ✅（veth は ❌） |
| 9 | chroot 脱出の実演 | なぜ `pivot_root` が必要かを体で理解する | 未 | ✅ |

### B. image と依存関係の実体（全部 Codespaces で可）

| # | 内容 | 状態 |
| --- | --- | --- |
| 10 | registry API を素手で叩く（token → manifest index → manifest） | 未 |
| 11 | config blob（JSON 1個）を読む: `Env` / `Cmd` / `Entrypoint` / `rootfs.diff_ids` / `history` | 未 |
| 12 | layer blob = ただの tar.gz であることを確認、digest を自分で `sha256` して照合 | 未 |
| 13 | **layer を順番に展開して rootfs を組む**（whiteout `.wh.` の処理を自作） | 未 |
| 14 | 13 + 既存の chroot を繋いで `run alpine:latest /bin/sh` 相当を完成させる | 未 |
| 15 | 依存関係: apk のメタデータ、musl vs glibc、静的リンクと `scratch` イメージ | 未 |

**13 が要点**: overlayfs（=mount）が使えなくても、**layer の tar を順番に展開すれば同じ rootfs になる**（重ね合わせをその場で flatten するだけ）。自分で書く必要があるのは削除の表現である whiteout（`.wh.foo` / `.wh..wh..opq`）の処理のみ。
→ **`fetch-rootfs.sh` を「本物の `docker pull` の自作版」に置き換えられる。** これで `docker run` の体感の8割が今の環境で完成する。欠けるのは中の `/proc`（`ps` が動かない）だけ。

---

## ⚠️ Codespaces / dev container の制約（重要）

**結論: `mount()` と cgroup 作成ができない。原因は capability ではなく AppArmor。**

### 実測ログ

`syscall.Mount` を直接叩くプローブで測った結果:

| 場所 | uid | capability | mount の結果 |
| --- | --- | --- | --- |
| そのまま | 0 (root) | `cap_sys_admin` **なし** | 全滅 `errno=EACCES(13)` |
| 新しい userns の中 | 0 | `=ep` = **全部持っている**（`cap_sys_admin` 込み） | 全滅 `errno=EACCES(13)` |

試したのは `tmpfs` / `proc` / `sysfs` / bind mount / `MS_REC|MS_PRIVATE` の5種類、**全部同じ EACCES**。

```
$ cat /proc/self/attr/current   → docker-default (enforce)
$ capsh --print                 → !cap_sys_admin / cap_sys_chroot あり
$ grep Seccomp /proc/self/status → Seccomp: 0        # seccompは無効（犯人ではない）
$ stat -fc %T /sys/fs/cgroup    → cgroup2fs（root は read-only）
$ cat /proc/sys/user/max_user_namespaces → 31749     # namespace自体は自由
```

### errno が犯人を教えてくれる

これが切り分けの決め手:

- **`EPERM(1)`** … capability 不足。カーネルの `may_mount()` → `ns_capable(CAP_SYS_ADMIN)` の失敗パス
- **`EACCES(13)`** … LSM の拒否。`security_sb_mount()` → AppArmor

返ってきたのは **EACCES**。つまり「権限が足りない」のではなく、**権限は足りているのに AppArmor が上から蹴っている**。

該当するのは Docker のデフォルトプロファイルにある文字どおりこの1行:

- [`moby/profiles` → `apparmor/template.go` の59行目 `deny mount,`](https://github.com/moby/profiles/blob/main/apparmor/template.go)

mount を使ったコンテナ脱出（`/proc/sys` の再マウント、ホストパスの bind mount 等）を防ぐためのもの。

### user namespace と capability の正しい理解

`create_user_ns()` は新しい namespace 内の credential を
`cap_permitted = cap_effective = cap_bset = CAP_FULL_SET` にする。

つまり **外側の bounding set に `cap_sys_admin` が無くても、新しい userns の中では持てる**。
実測の `Current: =ep` がそれ。rootless Podman が普通に `/proc` を mount できるのはこの仕組みのおかげ。

→ よくある誤解（このノートの旧版もそう書いていた）:
「土台のコンテナが CAP_SYS_ADMIN を持たないので userns を作っても超えられない」は**誤り**。
超えられないのは AppArmor であって capability ではない。
capability の欠如が効くのはホスト側リソースを触るとき（実デバイスの mount、`/sys/fs/cgroup` への書き込み）だけ。

### 中から外せるか → 無理

AppArmor プロファイルを付けているのは**この dev container を起動した外側の dockerd**（Codespaces 側）。
プロファイルを外す・遷移するには `CAP_MAC_ADMIN` が必要で、これは namespace 化されていないので userns を作っても手に入らない。
`devcontainer.json` の `runArgs` に `--security-opt apparmor=unconfined` を書いても Codespaces は基本無視する。

### 名前空間まわりの前提（記事との違い）

記事は実 root を前提に `CLONE_NEWUSER` 無しで書かれているが、コンテナの中（Codespaces 等）では
直接の名前空間作成が `Operation not permitted` になる。そこで **`CLONE_NEWUSER` を最初に作り**、
その中で root になってから他の名前空間を作る **rootless 方式**にしている（`SysProcAttr` の
`Cloneflags` に `CLONE_NEWUSER` と `UidMappings`/`GidMappings` を入れているのはこのため）。
実マシンで素の root なら user namespace 無しでも動く。

---

## 現在の既知の問題

**`main.go` は今の環境で panic する。** `child()` の

```go
must(syscall.Mount("proc", "/proc", "proc", 0, ""))
```

が無条件に実行され、Codespaces では必ず EACCES で落ちる。chroot（ステップ2）が成功しても
その後の mount で死ぬので、**下の「動作確認」コマンドはどれも通らない**。

→ **次の一手**: mount 失敗を `must` ではなく警告扱いにして、ステップ2までは今の環境で動くようにする。
「なぜ失敗するか」をエラーメッセージとしてコードに残しておけば、この調査結果とも繋がる。

その他、手を付けるときのメモ:

- mount した `/proc` の unmount（`defer`）がまだ無い
- `cmd.Run()` のところのコメントが `// メモリを上書き`（= `syscall.Exec` の説明）になっているが、
  実装は `cmd.Run()` で fork している。「なぜ Exec ではなく Run か / Exec にすると PID 1 がどう変わるか」は
  掘ると面白いところ

---

## 特権環境の入手方法

ステップ3〜5（mount / pivot_root / cgroup）を実演するには、特権付きの Linux 環境が要る。

### 方法 A: Codespaces のまま privileged にする（**未検証・試す価値あり**）

`ghcr.io/devcontainers/features/docker-in-docker` feature は、feature 定義自体が
`"privileged": true` を宣言していて、**Codespaces はこれを公式にサポートしている**。
docker を使いたいからではなく、**dev container を privileged で起動させるレバーとして**足す、という手が理屈上ある。
privileged になれば AppArmor は unconfined になり `cap_sys_admin` も付くので、mount がそのまま通るはず。

リビルド後の確認は2発:

```bash
cat /proc/self/attr/current        # → unconfined になっていれば勝ち
capsh --print | grep -o cap_sys_admin
```

### 方法 B: ローカル VS Code の Dev Container を特権化

`.devcontainer/devcontainer.json` に追記してリビルド（**ローカルの Docker で。Codespaces では効かない**）:

```jsonc
{
  "runArgs": ["--privileged"]
  // よりピンポイントにするなら:
  // "runArgs": ["--cap-add=SYS_ADMIN", "--security-opt", "apparmor=unconfined"]
}
```

`--privileged` で CAP_SYS_ADMIN が付き、AppArmor の拘束も外れて `mount` が通る。
cgroup v2 への書き込みも可能になる（cgroup namespace 委譲込み）。

### 方法 C: 普通の Linux VM / 実マシン

`sudo` できる Linux なら、そのまま動く。実 root であれば `CLONE_NEWUSER` 無しでも可。
一番素直で、cgroups まで含めて全部動く。

---

## 今の環境で使える／使えない syscall の早見表

| やりたいこと | 可否 | 備考 |
| --- | --- | --- |
| `clone(CLONE_NEWUSER/UTS/PID/NS/NET)` | ✅ | `max_user_namespaces=31749` |
| `chroot` | ✅ | `cap_sys_chroot` あり |
| `sethostname`（UTS ns 内） | ✅ | |
| `capset`（bounding set を削る） | ✅ | `cap_setpcap` あり。実測で drop 成功 |
| `seccomp(SECCOMP_SET_MODE_FILTER)` | ✅ | `actions_avail` に全アクションあり、`Seccomp: 0` なので自分で付けられる |
| `setrlimit` | ✅ | cgroups が使えない環境での粗い代替 |
| `mknod` | ✅ | `cap_mknod` あり（layer 展開時のデバイスノードに必要） |
| HTTP で registry を叩く | ✅ | Docker Hub の token 取得・manifest 取得を実測で確認 |
| `mount`（proc/tmpfs/bind/overlayfs） | ❌ | AppArmor `deny mount,` |
| `pivot_root` | ❌ | mount 依存 |
| cgroup の作成・書き込み | ❌ | `/sys/fs/cgroup` が read-only |
| veth でホストと接続 | ❌ | host netns で `CAP_NET_ADMIN` が必要 |

---

## セットアップ手順（clone 後 / 環境を移したとき）

```bash
cd container

# 1. Alpine rootfs を用意（rootfs/ は .gitignore 済みなので毎回必要）
./fetch-rootfs.sh

# 2. 動作確認  ※ 現状は mount で panic する（上の「既知の問題」参照）
go run . run /bin/ls /                  # → Alpine の中身が見える（chroot）
go run . run /bin/sh                    # → コンテナのシェルに入る
go run . run /bin/cat /etc/os-release   # → Alpine Linux

# 3. 特権環境なら（mount 後）
go run . run /bin/ps aux                # → コンテナ内プロセスだけが見える
```
