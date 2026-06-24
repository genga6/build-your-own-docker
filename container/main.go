package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// 使い方: go run . run <コマンド> [引数...]
//   例)  go run . run echo hello
//
// このプログラムは2つの顔を持つ:
//   run   ... あなたが叩く入口。自分自身を child として起動し直す。
//   child ... 起動し直された側。最終的に目的のコマンドを実行する。

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run . run <command> [args...]")
		os.Exit(1)
	}

	// os.Args[0] は実行ファイル名。os.Args[1] が最初の引数 (run / child)。
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("unknown command: " + os.Args[1])
	}
}

// run: あなたが最初に叩く側。
func run() {
	fmt.Printf("[run]	pid=%d args=%v\n", os.Getpid(), os.Args[2:])

	// ステップ0：自分自身を別プロセスとして起動しなおす。
	// "/proc/self/exe" = 今まさに動いている自分自身の実行ファイルを指すカーネルのリンク。
	// append([]string{"child"}, os.Args[2:]...) で
	//   ["child", "echo", "hello"] のような引数列を組み立てる。
	//   os.Args[2:] は "run" の後ろ全部（例: echo hello）。
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// ステップ1：新しい名前空間つきで子プロセスを起動する設定。
	// ここでは「プロセスIDの名前空間」を作る。つまり、子プロセスは自分が PID=1 だと思う。
	// これにより、子プロセスは自分のプロセスツリーの中で init プロセスのように振る舞える。
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | // rootlessの起点
			syscall.CLONE_NEWUTS | // ホスト名を隔離
			syscall.CLONE_NEWPID | // PID を隔離
			syscall.CLONE_NEWNS, // マウントを隔離

		// rootless のための UID/GID マッピングを設定する。
		// 「中のUIDを外のどのUIDに対応させるか」の翻訳表
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1}, // 子プロセスの root(0) を、親の自分の UID に対応させる
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1}, // 子プロセスの root(0) を、親の自分の GID に対応させる
		},
	}

	// 標準入出力を親（このプロセス）とつなぎ、
	// child の出力が見る／入力を渡す。
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// cmd.Run() は「起動して終了を待つ」。エラーは戻り値で返る（Go 流）。
	must(cmd.Run())
}

// child: 起動し直された側。ここで目的のコマンドを実行する。
func child() {
	fmt.Printf("[child] pid=%d args=%v\n", os.Getpid(), os.Args[2:])

	// UTS namespace 仲なので、ここでホスト名を変えてもホスト側には影響しない。
	must(syscall.Sethostname([]byte("container")))

	// os.Args[2:] = 目的のコマンドとその引数（例: echo hello）。
	// os.Args[2]  = "echo"（実行するコマンド）
	// os.Args[3:] = ["hello"]（その引数）
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// メモリを上書き
	must(cmd.Run())
}

// must: Go ではエラーは例外ではなく戻り値で返る（if err != nil で扱うのが基本）。
// 学習用に「エラーが返ったら即パニックで止める」小さなヘルパにまとめておく。
// 本番のコードでは握りつぶさず、文脈を付けて返すのが望ましい。
func must(err error) {
	if err != nil {
		panic(err)
	}
}
