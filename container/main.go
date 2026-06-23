package main

import (
	"fmt"
	"os"
	"os/exec"
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

	// "/proc/self/exe" = 今まさに動いている自分自身の実行ファイルを指すカーネルのリンク。
	// append([]string{"child"}, os.Args[2:]...) で
	//   ["child", "echo", "hello"] のような引数列を組み立てる。
	//   os.Args[2:] は "run" の後ろ全部（例: echo hello）。
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

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