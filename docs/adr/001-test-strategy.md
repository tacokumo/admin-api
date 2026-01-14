# ADR-001: テスト戦略の軽量化

## 状態

採用

## 決定日

2026-01-14

## コンテキスト

プロジェクトでは以前、testcontainersを使用したE2E/シナリオテストが `test/scenario/` に存在していたが、以下の課題があった：

- CI実行時間が長い（PostgreSQLコンテナの起動など）
- テスト実行が重く、開発中の頻繁な実行に向かない
- 外部依存が多く、テスト環境の整合性確保が困難

現在のプロジェクト構成：
- **アプリケーション**: Admin API (GoサーバーとCLIクライアント)
- **アーキテクチャ**: Echo + PostgreSQL + Redis
- **主要機能**: GitHub OAuth認証、プロジェクト管理、ユーザ・ロール管理
- **コード生成**: OpenAPI (ogen) + SQLCを使用

既存テスト：
- `pkg/pg/dsn_test.go`: PostgreSQL DSN生成のテーブル駆動テスト
- `api-spec/appconfig_test.go`: 設定バリデーションのテーブル駆動テスト

## 決定

E2E/シナリオテストを最小化し、以下のテストピラミッド構造を採用する：

### テスト分類と比重

1. **ユニットテスト (70%)**
   - 純粋関数、ビジネスロジック、バリデーション
   - 外部依存なし、高速実行
   - テーブル駆動テストを積極活用

2. **統合テスト (25%)**
   - DB操作、外部API連携（モック使用）
   - in-memoryデータベースやモックを活用した軽量テスト

3. **E2Eテスト (5%)**
   - 重要なユーザーフローのみ
   - CIで高速実行可能な最小限のテスト

### 具体的な実装方針

#### ユニットテスト対象
- **認証・セッション**: `pkg/auth/session/session.go`, `pkg/auth/oauth/github.go`
- **データベース**: `pkg/db/admindb/` (SQLC生成コード)
- **設定・環境**: `pkg/config/config.go`, `pkg/envconfig/envconfig.go`
- **ミドルウェア**: `pkg/middleware/`

#### 統合テスト対象
- **API層**: `pkg/apis/v1alpha1/server.go`
- **サービス層**: 認証フロー全体

#### E2Eテスト対象
- 認証フロー（ログイン成功）
- プロジェクト作成・取得の基本操作

### 使用技術・ツール

- **モック**: `github.com/stretchr/testify/mock`, `github.com/pashagolub/pgxmock`
- **軽量DB**: SQLite（メモリ）、pgxmock
- **HTTPテスト**: `net/http/httptest`
- **並列実行**: `t.Parallel()` でユニットテスト高速化

### テスト実行の分離

```bash
# 高速テスト（ユニット・統合）
go test -short ./...

# E2Eテスト（必要時のみ）
go test -tags=e2e ./...
```

## 結果

### 期待される利点

1. **開発速度向上**
   - 高速なユニットテストによる即座のフィードバック
   - CI実行時間の大幅短縮

2. **品質向上**
   - テーブル駆動による網羅的テスト
   - モックによる異常系テストの充実

3. **保守性向上**
   - 軽量テストによる変更時の影響範囲特定
   - 明確なテスト階層

### 実装優先度

- **Phase 1**: 認証・セッション管理、DB操作、設定処理のユニットテスト
- **Phase 2**: API層統合テスト、ミドルウェアテスト、クライアントテスト
- **Phase 3**: 最小限のE2Eテスト

### 成功指標

- `make test` の実行時間が従来のシナリオテスト時代より短縮
- テストカバレッジ向上（特に異常系）
- CI/CDパイプラインの高速化

## 関連する決定

- 今後のテスト追加は本ADRの方針に従う
- パフォーマンステストやセキュリティテストは別途検討

## 参考資料

- [Test Pyramid - Martin Fowler](https://martinfowler.com/articles/practical-test-pyramid.html)
- 既存の良いテスト例: `pkg/pg/dsn_test.go`, `api-spec/appconfig_test.go`