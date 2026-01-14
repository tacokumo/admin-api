## 概要

このプロジェクトは､Herokuやfly.ioなどのPlatform as a Serviceに影響を受けたTACOKUMOというサービスで､
テナントやユーザの管理を行うための"Admin"というサービスにおけるAPIサーバ実装です｡

./cmd/serverにAPIサーバのエントリポイントが､
./cmd/clientにCLIクライアントのエントリポイントがあります｡

## 基本

あらゆる変更を行ったあとは、以下のコマンドが動くことを確認してください。

```bash
make generate # to generate API code and sqlc files
make # to format/test/build the project
```

## テスト戦略

docs/adr/001-test-strategy.md を参考に記述してください。

## エラーハンドリング

すべてのエラーは握りつぶさず、ハンドリングしてください。
golangci-lintの出力は無視せず、すべてしたがってください。
