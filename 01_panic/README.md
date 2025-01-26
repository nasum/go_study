# panic

## goのversion 1.22と1.23での違い

goのversion 1.23ではpanicないで改行があると、インデントされて表示される

```cmd
panic: a problem
        hoge

goroutine 1 [running]:
main.main()
        /Users/masaya.nasu/src/github.com/nasum/go_study/01_panic/main.go:4 +0x2c
exit status 2
```

goのversion 1.22ではインデントされない。

```cmd
panic: a problem
hoge

goroutine 1 [running]:
main.main()
        /Users/masaya.nasu/src/github.com/nasum/go_study/01_panic/go_122/main.go:4 +0x2c
exit status 2
```

修正はこのCL
https://github.com/golang/go/commit/69e75c8581e15328454bb6e2f1dc347f73616b37

コードとしてはただインデントしているだけ

```
// printindented prints s, replacing "\n" with "\n\t".
func printindented(s string) {
	for {
		i := bytealg.IndexByteString(s, '\n')
		if i < 0 {
			break
		}
		i += len("\n")
		print(s[:i])
		print("\t")
		s = s[i:]
	}
	print(s)
}
```

panicのコードはこちら
https://github.com/golang/go/blob/40b3c0e58a0ae8dec4684a009bf3806769e0fc41/src/runtime/panic.go

panicはプリミティブな機能だと思っていたのでgoで実装されているとは思っていなかった。ただ関数名がgopanicなので普段使っているpanicとは違う関数名である。この違いは一体何なのか気になった。

## panicはSSA形式に変換されるときruntime.panicとして使われる

goのコードをコンパイルするとき、一旦中間表現としてSSA形式に変換される。コンパイルする対象にpanicの実行があったらそれをruntime.panicに変換してSSA形式にする。

以下のようなSSA形式になる。

```
main func()
  b1:
    (?) v1 = InitMem <mem>
    (?) v3 = SB <uintptr>
    (+4) v4 = InlMark <void> [0] v1
    (?) v5 = Addr <*uint8> {type:string} v3
    (?) v6 = Addr <*string> {main..stmp_0} v3
    (8) v7 = IMake <interface {}> v5 v6
    (8) v8 = StaticLECall <mem> {AuxCall{runtime.gopanic}} [16] v7 v1
    (8) v9 = SelectN <mem> [0] v8
    Exit v9

```

SSA形式のファイルにはruntime.gopanicのSSA形式のファイルが出てこないが、これはruntime.panicがプリコンパイルされているからである。コンパイル時にコンパイラが使ってビルドする。

TODO: コンパイラ周りについて調べたい

panicをコードとして書くとruntime.gopanicになりgopanicの実装がそのまま使われているということがわかった。

## panicを手動で実行しない場合panicはどう起こるのか

panicを起こすコード

```
package main

func test() int {
	var arr = []int{1, 2, 3}
	return arr[3]
}
```

をSSA形式に出力した

```
test func() int
  b1:
    (?) v1 = InitMem <mem>
    (?) v2 = SP <uintptr>
    (?) v5 = Const64 <int> [0]
    (-4) v6 = LocalAddr <*[3]int> {.autotmp_3} v2 v1
    (4) v7 = Zero <mem> {[3]int} [24] v6 v1
    (-4) v8 = LocalAddr <*[3]int> {.autotmp_3} v2 v7
    (?) v9 = Const64 <int> [1]
    (4) v10 = NilCheck <*[3]int> v8 v7
    (?) v11 = Const64 <int> [3]
    (4) v12 = PtrIndex <*int> v10 v5
    (4) v13 = Store <mem> {int} v12 v9 v7
    (?) v14 = Const64 <int> [2]
    (4) v15 = NilCheck <*[3]int> v8 v13
    (4) v16 = PtrIndex <*int> v15 v9
    (4) v17 = Store <mem> {int} v16 v14 v13
    (4) v18 = NilCheck <*[3]int> v8 v17
    (4) v19 = PtrIndex <*int> v18 v14
    (4) v20 = Store <mem> {int} v19 v11 v17
    (4) v21 = NilCheck <*[3]int> v8 v20
    (-4) v22 = Copy <*int> v21
    (4) v23 = IsSliceInBounds <bool> v5 v11
    If v23 -> b2 b3 (likely)
  b2: <- b1
    (4) v26 = Sub64 <int> v11 v5
    (4) v27 = SliceMake <[]int> v22 v26 v26 (arr[[]int])
    (5) v28 = SliceLen <int> v27
    (5) v29 = IsInBounds <bool> v11 v28
    If v29 -> b4 b5 (likely)
  b3: <- b1
    (-4) v24 = Copy <mem> v20
    (4) v25 = PanicBounds <mem> [6] v5 v11 v24
    Exit v25
  b4: <- b2
    (5) v32 = SlicePtr <*int> v27
    (5) v33 = PtrIndex <*int> v32 v11
    (-5) v34 = Copy <mem> v20
    (5) v35 = Load <int> v33 v34
    (5) v36 = MakeResult <int,mem> v35 v34
    Ret v36
  b5: <- b2
    (-5) v30 = Copy <mem> v20
    (5) v31 = PanicBounds <mem> [0] v11 v28 v30
    Exit v31
name arr[[]int]: [v27]

```

コンパイラがPanicBoundsの分岐をSSAに追加。各アーキテクチャのssa.goで
BoundsCheckFunc[...] = typecheck.LookupRuntimeFunc("panicIndex")などの設定をがあり
https://github.com/golang/go/blob/40b3c0e58a0ae8dec4684a009bf3806769e0fc41/src/cmd/compile/internal/amd64/ssa.go#L1143

コード生成するときどのランタイム関数を実行するか確定する
https://github.com/golang/go/blob/40b3c0e58a0ae8dec4684a009bf3806769e0fc41/src/cmd/compile/internal/ssagen/ssa.go#L181

このコンパイラの動きでpanic.goの実装が実行されるようになる。