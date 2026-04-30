# Database Schema

このアプリケーションでは SQLite3 を使用しており、以下の 3 つのテーブルでデータを管理しています。

## 1. users テーブル
ユーザー情報を管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | ユーザーの一意なID |
| uuid | TEXT | NOT NULL UNIQUE | 外部参照用のユニークな識別子 |
| name | TEXT | | ユーザー名 |
| email | TEXT | UNIQUE | Eメールアドレス（ログインに使用） |
| password | TEXT | | ハッシュ化されたパスワード |
| created_at | DATETIME | | アカウント作成日時 |

## 2. todos テーブル
ユーザーが作成した TODO タスクを管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | TODOの一意なID |
| content | TEXT | | TODOの内容 |
| user_id | INTEGER | FOREIGN KEY (users.id) ON DELETE CASCADE | 作成したユーザーのID |
| created_at | DATETIME | | タスク作成日時 |

## 3. sessions テーブル
ログインセッションを管理するテーブルです。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | セッションの一意なID |
| uuid | TEXT | NOT NULL UNIQUE | ブラウザのクッキーに保存されるUUID |
| email | TEXT | | ログイン中のユーザーのEメール |
| user_id | INTEGER | FOREIGN KEY (users.id) ON DELETE CASCADE | ログイン中のユーザーID |
| created_at | DATETIME | | セッション作成日時 |
