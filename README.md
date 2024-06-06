# cloud-functions-samples

Cloud Functions 活用のサンプルコードたち（ツッコミ歓迎）

コードコメントに日本語と英語が混ざっているのは気分です。特に理由はありません。

### 依存ツール

動かすには下記の CLI ツールが必要です。適宜インストールしてください。

- `direnv` 環境変数の設定（オプション）
- `gcloud` Cloud Functions 関数のデプロイやテスト
- `terraform` 依存リソースの管理

## gen2 ディレクトリ

Cloud Functions 第2世代で汎用的に使えそうなコードを置いています。（他のところでも現時点では特に明示しない限り第2世代を利用しています）

作成する関数:

- `functions-samples-start` (HTTP トリガー。 コマンドで呼び出す)
- `functions-samples-hook` (Pub/Sub トリガー。 `functions-samples-start` から呼び出される)

デプロイ方法:

````shell
./gen2/deploy.sh start
````

テスト方法:

```shell
gcloud functions call functions-samples-start --region=asia-northeast1 --gen2
```

## terraform ディレクトリ

Terraform は個別の関数ごとに分けずに、全ての依存リソースを一括で作成しています。

固定費が発生するリソースはありません。
