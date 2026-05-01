.PHONY: up up-build down stop build test

# APP_DIR はソースコードのルートディレクトリを指定します。
APP_DIR=src/todo_app
# COMPOSE は Docker Compose コマンドの共通ベースです。-f フラグで設定ファイルを明示することで、
# カレントディレクトリに関わらず Makefile から一貫した操作を可能にします。
COMPOSE=docker compose -f $(APP_DIR)/docker-compose.yml

# TIMESTAMP はデータベースのダンプファイル名に使用します。
# Windows 環境での実行を想定し、PowerShell を利用して OS に依存しない日付形式を取得しています。
TIMESTAMP=$(shell powershell -NoProfile -Command "Get-Date -Format 'yyyyMMdd_HHmmss'")
# LATEST_DUMP は最新のバックアップファイルを自動的に特定します。
# 開発の再開時に前回のデータを自動復旧させるための利便性向上を目的としています。
LATEST_DUMP=$(shell powershell -NoProfile -Command "Get-ChildItem -Filter '*_dump.sql' $(APP_DIR) | Sort-Object LastWriteTime -Descending | Select-Object -First 1 -ExpandProperty Name")

# up-build はコンテナを再ビルドして起動します。
# 開発初期や Dockerfile の変更後に、クリーンな状態で最新のバックアップを適用するために使用します。
up-build:
	$(COMPOSE) up -d --build
ifneq ($(LATEST_DUMP),)
	@echo "Restoring database from latest dump: $(LATEST_DUMP)..."
	-$(COMPOSE) exec -T app sqlite3 /app/data/webapp.sql "DROP TABLE IF EXISTS users; DROP TABLE IF EXISTS todos; DROP TABLE IF EXISTS sessions;"
	-$(COMPOSE) exec -T app sqlite3 /app/data/webapp.sql < $(APP_DIR)/$(LATEST_DUMP)
endif

# up は既存のイメージを使用してコンテナを起動します。
# 通常の開発ルーチンで使用し、前回の終了時の状態を LATEST_DUMP から復元します。
up:
	$(COMPOSE) up -d
ifneq ($(LATEST_DUMP),)
	@echo "Restoring database from latest dump: $(LATEST_DUMP)..."
	-$(COMPOSE) exec -T app sqlite3 /app/data/webapp.sql "DROP TABLE IF EXISTS users; DROP TABLE IF EXISTS todos; DROP TABLE IF EXISTS sessions;"
	-$(COMPOSE) exec -T app sqlite3 /app/data/webapp.sql < $(APP_DIR)/$(LATEST_DUMP)
endif

# down はコンテナを停止・削除しますが、その前にデータベースの状態を SQL 形式でダンプします。
# Docker ボリューム内のデータをホスト側に永続化し、Git 等での履歴管理や他メンバーへの共有を容易にします。
down:
	@echo "Dumping database to $(TIMESTAMP)_dump.sql..."
	-$(COMPOSE) exec app sh -c "sqlite3 /app/data/webapp.sql .dump | grep -v sqlite_sequence" > $(APP_DIR)/$(TIMESTAMP)_dump.sql
	$(COMPOSE) down

# stop はコンテナを破棄せず停止のみ行います。
# データのダンプや復元を伴わない、一時的な中断に使用します。
stop:
	$(COMPOSE) stop

# build は Go アプリケーションをローカルでビルドします。
# Docker を介さず、ネイティブなバイナリが必要な場合に使用します。
build:
	cd $(APP_DIR) && go build -o ../../bin/todo_app main.go

# test は Docker コンテナ内の 'tester' サービスでユニットテストを実行します。
# CGO_ENABLED=1 が必要なライブラリ (sqlite3) を、ホスト環境を汚さずに一貫した環境でテストするために Docker を利用します。
test:
	$(COMPOSE) run --rm tester
