# ローカル開発

このドキュメントでは、Admin APIのローカル開発環境を構築する手順について説明します。

## 概要

ローカル開発環境には以下のサービスが含まれています：

- **Admin API** (http://localhost:8080) - メインのGo API サーバー
- **PostgreSQL** (localhost:5432) - データベース
- **Valkey** (localhost:6379) - Redis互換のセッションストレージ

## 事前準備

### 必要なソフトウェア

- Docker & Docker Compose
- Make
- Git

### GitHub OAuth アプリケーションの設定

1. GitHub にて OAuth アプリケーションを作成します
   - Settings > Developer settings > OAuth Apps > New OAuth App
   - Application name: `Tacokumo Admin API (Local)`
   - Homepage URL: `http://localhost:8080`
   - Authorization callback URL: `http://localhost:8080/v1alpha1/auth/callback`

2. Client ID と Client Secret をメモしておきます

## 環境構築手順

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd admin-api
```

### 2. 環境変数の設定

```bash
# 環境変数テンプレートをコピー
make setup-env

# .envファイルを編集して GitHub OAuth の設定を追加
vim .env
```

`.env` ファイルで設定が必要な項目：

```bash
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
GITHUB_CALLBACK_URL=http://localhost:8080/v1alpha1/auth/callback
GITHUB_ALLOWED_ORGS=your-org-1,your-org-2
SESSION_TTL=24h
```

### 3. 開発環境の起動

```bash
# すべてのサービスを起動（ビルド、マイグレーション含む）
make docker-compose-up
```

このコマンドは以下を自動的に実行します：
- Dockerイメージのビルド
- すべてのサービスの起動
- データベースマイグレーションの実行

### 4. 環境の検証

```bash
# 環境が正しく設定されているか確認
make verify-setup
```

## APIエンドポイント

主要なエンドポイント：

- `GET /health` - ヘルスチェック
- `GET /v1alpha1/auth/login` - GitHub OAuth ログイン
- `POST /v1alpha1/auth/logout` - ログアウト
- `GET /v1alpha1/user` - ユーザー情報取得
- `GET /v1alpha1/projects` - プロジェクト一覧

## 開発コマンド

### 基本操作

```bash
# 開発環境の起動
make docker-compose-up

# 開発環境の停止
make docker-compose-down

# 環境の検証
make verify-setup

# 環境のリセット（すべてのデータが削除されます）
make reset-dev-env
```

### ビルドとテスト

```bash
# コード生成
make generate

# テスト実行
make test

# ビルド
make build

# フォーマット
make format

# リント
make lint

# 全部実行
make all
```

### Docker Compose操作

```bash
# 特定のサービスのログを確認
docker compose logs admin_api
docker compose logs postgresql
docker compose logs valkey

# サービスの再起動
docker compose restart admin_api

# データベースに接続
docker compose exec postgresql psql -U admin_api -d tacokumo_admin_db
```

## トラブルシューティング

### よくある問題

#### 1. データベース接続エラー

```bash
# データベースの状態確認
docker compose logs postgresql

# マイグレーションの再実行
make migrate
```

#### 2. Redis接続エラー

```bash
# Valkeyの状態確認
docker compose logs valkey
docker compose exec valkey valkey-cli ping
```

#### 3. ポート競合

デフォルトポートが使用されている場合：
- Admin API: 8080番ポート
- PostgreSQL: 5432番ポート
- Valkey: 6379番ポート

使用中のプロセスを確認：

```bash
# macOS/Linux
lsof -i :8080
lsof -i :5432

# プロセス終了
kill -9 <PID>
```

#### 4. 環境の完全リセット

```bash
# すべてをリセット（データ損失注意）
make reset-dev-env

# 環境変数を再設定
make setup-env
vim .env

# 再起動
make docker-compose-up
```

### ログの確認

```bash
# すべてのサービスのログ
docker compose logs -f

# 特定のサービス
docker compose logs -f admin_api

# エラーのみ
docker compose logs --tail=50 admin_api | grep ERROR
```

## 開発Tips

### API テストについて

1. curl や Postman で API の動作テストを実行
2. ログは `docker compose logs` で確認

### データベース操作

```bash
# データベースシェルに接続
docker compose exec postgresql psql -U admin_api -d tacokumo_admin_db

# テーブル一覧
\dt

# スキーマ確認
\d table_name
```

### Hot reload

Go コードを変更した場合：

```bash
# APIサーバーのみ再ビルド・再起動
docker compose up -d --build admin_api
```


## セキュリティ注意事項

- `.env`ファイルには機密情報が含まれるため、Gitにコミットしないでください
- 本番環境では異なる設定が必要です

## より詳細な設定

詳細な設定については以下のファイルを参照してください：

- `compose.yaml` - Docker Compose設定
- `develop/server.yaml` - API サーバー設定
- `.env.example` - 環境変数テンプレート
- `scripts/verify-setup.sh` - 環境検証スクリプト