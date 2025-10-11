-- Create dedicated schema for tacokumo admin
CREATE SCHEMA IF NOT EXISTS tacokumo_admin;

CREATE TABLE tacokumo_admin.projects (
  id   BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  name VARCHAR(64) NULL, -- プロジェクト名
  bio VARCHAR(256) NULL, -- プロジェクトの説明 
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (name) -- プロジェクト名はユニーク
);

-- Adminアカウント
-- Auth0認証を前提としているが、admin側でもデータを持っておくことで、認可の実装を可能にする
CREATE TABLE tacokumo_admin.accounts (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  email VARCHAR(256) NOT NULL, -- アカウントのメールアドレス
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (email) -- メールアドレスはユニーク
);

-- Auth0のユーザテーブルと、Admin DBのアカウント情報を紐づけて管理する
-- パスワードは管理しない
CREATE TABLE tacokumo_admin.account_identities (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  account_id BIGINT NOT NULL REFERENCES tacokumo_admin.accounts(id) ON DELETE CASCADE,
  email_verified BOOLEAN NOT NULL,
  issuer VARCHAR(256) NOT NULL, -- Auth0のドメイン
  sub VARCHAR(64) NOT NULL, -- Auth0のユーザーID
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (issuer, sub) -- Auth0のユーザーIDはユニーク
);

-- Adminアカウントとプロジェクトの関連付け
CREATE TABLE tacokumo_admin.project_account_relationships (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  project_id BIGINT NOT NULL REFERENCES tacokumo_admin.projects(id) ON DELETE CASCADE,
  account_id BIGINT NOT NULL REFERENCES tacokumo_admin.accounts(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (project_id, account_id) -- 同じアカウントが同じプロジェクトに複数回関連付けられないようにする
);