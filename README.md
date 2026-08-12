# build-your-own-docker

Dockerの実装 ＋ Docker imageの実体 ＋ 依存関係・バージョン

実装の進捗・環境の制約・学習ロードマップは [`container/NOTES.md`](container/NOTES.md) に置いてある。

## 参考

### 入口（自作コンテナのチュートリアル）

- [What even is a container? (Julia Evans)](https://jvns.ca/blog/2016/10/10/what-even-is-a-container/) — 15分で全体像
- [Build Your Own Container Using Less than 100 Lines of Go (InfoQ)](https://www.infoq.com/articles/build-a-container-golang/) — このリポジトリの出発点
- [Containers From Scratch • Liz Rice • GOTO 2018](https://www.youtube.com/watch?v=8fi7uSYlOdc)
- [lizrice/containers-from-scratch](https://github.com/lizrice/containers-from-scratch) — 上の動画の完成コード。`pivot_root` や cgroups の書き方の答え合わせに
- [Container Learning Path (iximiuz)](https://iximiuz.com/en/posts/container-learning-path/) — 何をどの順で学ぶかの地図
- [How Containers Work (wizardzines)](https://wizardzines.com/zines/containers/) — 全体像を絵で掴む

### 名前空間・カーネル側の一次情報

チュートリアルが「こう書けば動く」で済ませるところの理由が全部ここにある。

- [Namespaces in operation (LWN, Michael Kerrisk)](https://lwn.net/Articles/531114/) — 全7回シリーズの第1回。namespace の決定版解説
- [`namespaces(7)`](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [`user_namespaces(7)`](https://man7.org/linux/man-pages/man7/user_namespaces.7.html) — UID/GIDマッピングと capability の扱い。**このリポジトリの rootless 方式の根拠**
- [`capabilities(7)`](https://man7.org/linux/man-pages/man7/capabilities.7.html) — permitted / effective / bounding / inheritable の違い
- [`pivot_root(2)`](https://man7.org/linux/man-pages/man2/pivot_root.2.html) — なぜ `chroot` では駄目なのか
- [`seccomp(2)`](https://man7.org/linux/man-pages/man2/seccomp.2.html)
- [`cgroups(7)`](https://man7.org/linux/man-pages/man7/cgroups.7.html) / [cgroup v2 (kernel docs)](https://docs.kernel.org/admin-guide/cgroup-v2.html)
- [Rootless Containers](https://rootlesscontaine.rs/) — rootless の制約と回避策のまとめ

### image / registry の実体（目的Bの一次情報）

- [OCI Image Spec](https://github.com/opencontainers/image-spec) — image = manifest + config JSON + layer tar
- [OCI Image Spec: layer.md](https://github.com/opencontainers/image-spec/blob/main/layer.md) — **whiteout（`.wh.` による削除の表現）の仕様**。layer を自分で展開するならここ
- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec) — `docker pull` が実際に叩いている HTTP API
- [OCI Runtime Spec](https://github.com/opencontainers/runtime-spec) — `config.json` に何を書けばコンテナになるのか（自作コードとの対応表として読める）

### 手を動かすツール（自作実装の答え合わせ用）

- [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md) — `crane manifest alpine` / `crane config alpine` / `crane blob` で manifest・config・layer を直接覗く
- [dive](https://github.com/wagoodman/dive) — レイヤごとの増減を見る
- [skopeo](https://github.com/containers/skopeo) — registry の inspect / コピー

### Docker が実際に何を設定しているか（答え合わせ用）

- [`moby/profiles` → `seccomp/default.json`](https://github.com/moby/profiles/blob/main/seccomp/default.json) — Docker のデフォルト seccomp プロファイル。塞いでいる syscall の一覧
- [`moby/profiles` → `apparmor/template.go`](https://github.com/moby/profiles/blob/main/apparmor/template.go) — `docker-default` プロファイルの中身。59行目の `deny mount,` が Codespaces で mount できない直接の原因（詳細は NOTES.md）
