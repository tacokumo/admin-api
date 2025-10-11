-- Create dedicated schema for tacokumo admin
CREATE SCHEMA IF NOT EXISTS tacokumo_admin;

-- プロジェクト情報を保持するテーブル
CREATE TABLE tacokumo_admin.projects (
  id   BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  name VARCHAR(64) NOT NULL, -- プロジェクト名
  description VARCHAR(256) NOT NULL, -- プロジェクトの説明 
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (name) -- プロジェクト名はユニーク
);

-- ユーザ情報を保持するテーブル
CREATE TABLE tacokumo_admin.users (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  email VARCHAR(256) NOT NULL, -- アカウントのメールアドレス
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (email) -- メールアドレスはユニーク
);

-- プロジェクトのオーナー情報を保持するテーブル
CREATE TABLE tacokumo_admin.project_owners(
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  project_id BIGINT NOT NULL REFERENCES tacokumo_admin.projects(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES tacokumo_admin.users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (project_id, user_id) -- 同じプロジェクトとユーザの組み合わせはユニーク
);

-- ユーザがGitHub連携している場合の情報を保持するテーブル
CREATE TABLE tacokumo_admin.github_accounts (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auth0によって提供されるIdPの情報と、Admin DBのユーザ情報を紐づけて管理する
-- パスワードは管理しない
CREATE TABLE tacokumo_admin.account_identities (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  user_id BIGINT NOT NULL REFERENCES tacokumo_admin.users(id) ON DELETE CASCADE,
  email_verified BOOLEAN NOT NULL,
  issuer VARCHAR(256) NOT NULL, -- Auth0のドメイン
  sub VARCHAR(64) NOT NULL, -- Auth0のユーザーID
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (issuer, sub) -- Auth0のユーザーIDはユニーク
);


-- ユーザやユーザグループに割り当てられるロールの定義
CREATE TABLE tacokumo_admin.roles (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  project_id BIGINT NOT NULL REFERENCES tacokumo_admin.projects(id) ON DELETE CASCADE,
  name VARCHAR(32) NOT NULL, -- ロール名 (例: admin, editor, viewer)
  description VARCHAR(256) NOT NULL, -- ロールの説明
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (project_id, name) -- ロール名はプロジェクト内でユニーク
);

-- ロールに関連付けられる属性 (例: 権限の詳細設定)
-- 現状事前定義された属性しか挿入されないため､あくまでも実装上の都合でテーブルを作成しているだけ
CREATE TABLE tacokumo_admin.role_attributes (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  name VARCHAR(64) NOT NULL, -- 属性名
  description VARCHAR(256) NOT NULL, -- 属性の説明
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (name) -- 属性名はユニーク
);

-- ロールと属性の多対多の関係を管理する中間テーブル
CREATE TABLE tacokumo_admin.role_attributes_relations (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  role_id BIGINT NOT NULL REFERENCES tacokumo_admin.roles(id) ON DELETE CASCADE,
  role_attribute_id BIGINT NOT NULL REFERENCES tacokumo_admin.role_attributes(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (role_id, role_attribute_id) -- 同じロールと属性の組
);

-- ユーザグループ情報を保持するテーブル
CREATE TABLE tacokumo_admin.usergroups (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  project_id BIGINT NOT NULL REFERENCES tacokumo_admin.projects(id) ON DELETE CASCADE,
  name VARCHAR(64) NOT NULL, -- ユーザグループ名
  description VARCHAR(256) NOT NULL, -- ユーザグループの説明
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (project_id, name) -- ユーザグループ名はプロジェクト内でユニーク
);

-- ユーザとユーザグループの多対多の関係を管理する中間テーブル
CREATE TABLE tacokumo_admin.user_usergroups_relations (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  user_id BIGINT NOT NULL REFERENCES tacokumo_admin.users(id) ON DELETE CASCADE,
  usergroup_id BIGINT NOT NULL REFERENCES tacokumo_admin.usergroups(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, usergroup_id) -- 同じユーザとユーザグループの組み合わせはユニーク
);

-- ユーザに割り当てられたロールを管理するテーブル
CREATE TABLE tacokumo_admin.user_role_relations (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  user_id BIGINT NOT NULL REFERENCES tacokumo_admin.users(id) ON DELETE CASCADE,
  role_id BIGINT NOT NULL REFERENCES tacokumo_admin.roles(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, role_id) -- 同じユーザとロール
);

-- ユーザグループに割り当てられたロールを管理するテーブル
CREATE TABLE tacokumo_admin.usergroup_role_relations (
  id BIGSERIAL PRIMARY KEY, -- ひとまず主キーはBIGSERIALで、UUIDv7に移行することもあるかもしれない
  usergroup_id BIGINT NOT NULL REFERENCES tacokumo_admin.usergroups(id) ON DELETE CASCADE,
  role_id BIGINT NOT NULL REFERENCES tacokumo_admin.roles(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (usergroup_id, role_id) -- 同じユーザグループとロール
);