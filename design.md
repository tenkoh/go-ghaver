## コード規約
- Go言語のベストプラクティスに基づく簡潔な実装
- テストしやすいインターフェース設計。ただし複雑にしすぎない。
- 主要ロジックはプロジェクトルートにフラットに配置。配布用のCLIコードはcmd/ghaver/main.goとして実装。

## 技術選定
- コマンドラインパーサー。`alecthomas/kong`
- 曖昧検索.`koki-develop/go-fzf`
- バージョン情報取得にはGitHubのAPIを使用。
- 著名なGitHub Actions一覧の管理はJSONファイルを用いる。`cmd/ghaver/actions.json`。embedでファイルをCLIに埋め込む。
- JSONファイルの内容は以下の通り。ユーザーとリポジトリの階層構造。

```json
{
    "actions": [
        "cache",
        "checkout"
    ],
    "aws-actions": [ 
        "configure-aws-credential"
     ]
}
```

## 詳細使用
- `--sha`オプションを使用すると`abc123def456ghq789123123 # v1.2.3`のように、`SHA # バージョン記号`の形式を出力する。
- オプションなしではバージョン記号だけを出力する。
- 出力先が標準出力の場合は末尾改行あり、パイプの場合は末尾改行なしで出力する。