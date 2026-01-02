# ローカル開発

このドキュメントでは､Admin APIのローカル開発環境を構築する手順について説明します｡

## 事前準備

<https://github.com/pepabo/tacokumo> をクローンし､
terraformをdevelop環境に対して実行しておきます｡

## 開発環境

```shell
envchain tacokumo-admin make docker-compose-up
envchain tacokumo-cli ./bin/client project list
```