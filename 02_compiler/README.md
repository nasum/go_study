# compiler

compilerの[README](https://github.com/golang/go/tree/master/src/cmd/compile)

コンパイルの流れ。
* パース
* 型チェック
* 中間表現の処理
* SSA形式への変換
* 機械語への翻訳

パスしてASTを作り、型チェックをし、ASTの最適化をし、SSA形式に変換し、SSAの最適化をした後機械語に翻訳