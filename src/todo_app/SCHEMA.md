# Database Schema

このアプリケーションでは SQLite3 を使用しており、以下の 3 つのテーブルでデータを管理しています。

## 1. users テーブル
ユーザー情報を管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | ユーザーの一意なID |
| uuid | STRING | NOT NULL UNIQUE | 外部参照用のユニークな識別子 |
| name | STRING | | ユーザー名 |
| email | STRING | | Eメールアドレス |
| password | STRING | | ハッシュ化されたパスワード |
| created_at | DATETIME | | アカウント作成日時 |

## 2. todos テーブル
ユーザーが作成した TODO タスクを管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | TODOの一意なID |
| content | TEXT | | TODOの内容 |
| user_id | INTEGER | | 作成したユーザーのID (`users.id` と紐付け) |
| created_at | DATETIME | | タスク作成日時 |

## 3. sessions テーブル
ログインセッションを管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | セッションの一意なID |
| uuid | STRING | NOT NULL UNIQUE | ブラウザのクッキーに保存されるUUID |
| email | STRING | | ログイン中のユーザーのEメール |
| user_id | INTEGER | | ログイン中のユーザーID (`users.id` と紐付け) |
| created_at | DATETIME | | セッション作成日時 |
