# Kubernetes Deployment

このディレクトリには、TACOKUMO Admin APIをKubernetesにデプロイするためのマニフェストファイルが含まれています。

## 前提条件

- Kubernetesクラスタが稼働していること
- `kubectl` コマンドがインストールされていること
- Dockerイメージがビルドされ、コンテナレジストリにプッシュされていること

## デプロイ前の準備

### 1. TLS証明書の準備

```bash
# TLS証明書とキーを準備し、Secretを作成
kubectl create secret tls admin-api-tls \
  --cert=path/to/api-server.crt \
  --key=path/to/api-server.key \
  -n tacokumo-admin
```

### 2. Secret の更新

`secret.yaml` ファイルを編集して、以下の値を実際の値に置き換えてください：

- `AUTH0_CLIENT_ID`: Auth0のクライアントID
- `AUTH0_CLIENT_SECRET`: Auth0のクライアントシークレット
- `POSTGRES_PASSWORD`: PostgreSQLのパスワード（必要に応じて変更）

### 3. ConfigMap の更新

必要に応じて `configmap.yaml` の設定値を環境に合わせて調整してください。

### 4. Deployment の更新

`admin-api-deployment.yaml` のイメージを実際のコンテナレジストリのイメージに変更してください：

```yaml
image: your-registry.example.com/tacokumo/admin-api:tag
```

## デプロイ方法

### kubectl apply を使用する場合

```bash
# すべてのマニフェストを適用
kubectl apply -f k8s/

# または個別に適用
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres-pvc.yaml
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/postgres-service.yaml
kubectl apply -f k8s/admin-api-deployment.yaml
kubectl apply -f k8s/admin-api-service.yaml
```

### kustomize を使用する場合

```bash
kubectl apply -k k8s/
```

## 確認

```bash
# Podの状態を確認
kubectl get pods -n tacokumo-admin

# Serviceの状態を確認
kubectl get services -n tacokumo-admin

# Deploymentの状態を確認
kubectl get deployments -n tacokumo-admin

# ログを確認
kubectl logs -n tacokumo-admin -l app=admin-api

# PostgreSQLのログを確認
kubectl logs -n tacokumo-admin -l app=postgresql
```

## アクセス方法

### LoadBalancer の場合

```bash
# External IPを確認
kubectl get service admin-api -n tacokumo-admin

# 表示されたEXTERNAL-IPを使用してアクセス
# https://<EXTERNAL-IP>:8444
```

### Port Forwarding の場合

```bash
kubectl port-forward -n tacokumo-admin service/admin-api 8444:8444

# https://localhost:8444 でアクセス可能
```

## トラブルシューティング

### Podが起動しない場合

```bash
# Pod の詳細を確認
kubectl describe pod -n tacokumo-admin <pod-name>

# イベントを確認
kubectl get events -n tacokumo-admin --sort-by='.lastTimestamp'
```

### データベース接続エラーの場合

```bash
# PostgreSQL Podが起動しているか確認
kubectl get pods -n tacokumo-admin -l app=postgresql

# PostgreSQL Podのログを確認
kubectl logs -n tacokumo-admin -l app=postgresql

# PostgreSQL Podに接続してテスト
kubectl exec -it -n tacokumo-admin <postgresql-pod-name> -- psql -U admin_api -d tacokumo_admin_db
```

## 削除方法

```bash
# すべてのリソースを削除
kubectl delete -f k8s/

# または
kubectl delete namespace tacokumo-admin
```

## カスタマイズ

### レプリカ数の変更

```bash
kubectl scale deployment admin-api -n tacokumo-admin --replicas=3
```

### リソース制限の調整

`admin-api-deployment.yaml` の `resources` セクションを編集してください。

### ストレージクラスの指定

`postgres-pvc.yaml` の `storageClassName` をクラスタに適したものに変更してください。

## 本番環境での考慮事項

1. **Ingress の設定**: LoadBalancer の代わりに Ingress を使用することを推奨
2. **HPA の設定**: Horizontal Pod Autoscaler を設定してオートスケーリングを有効化
3. **NetworkPolicy**: ネットワークポリシーを設定してセキュリティを強化
4. **データベースのバックアップ**: PostgreSQL のバックアップ戦略を実装
5. **監視とログ**: Prometheus, Grafana, ELKスタックなどを導入
6. **Secret 管理**: External Secrets Operator や Sealed Secrets を使用して Secret を安全に管理
