# GAE-GO

## Why

[GAE](https://cloud.google.com/appengine/docs/?hl=ja)にはStandardとFlexibleがある。


### Language

|Language|Starndard|Flexible|
|---|---|---|
|Python|○|○|
|Java|○|○|
|Node|×|○|
|Go|○|○|
|Ruby|×|○|
|PHP|○|○|
|.NET|×|○|
|Custom Runtimes|×|○|

StandardだとGoが**速い**


* gcloud関連のモジュールはdockerコンテナに入れてPCを汚したくない。[GAEのDocument](https://cloud.google.com/appengine/docs/standard/go/download?hl=ja)だと即座にインストールさせられる

* メルカリのimageはdebianなので使いたくない。google/cloud-sdk:alpineがいい。
https://github.com/mercari/docker-appengine-go

* glideは使いたくない。depがいい。

* GAE standard Go1.8で使いたい。

これらを満たすサンプルがなかったのでコツコツ作った。

|Sample|Information|
|---|---|
|**hello**|画面にただhello worldを表示するだけのサンプル|
|**template**|デフォルトのテンプレートエンジンを使ってページを表示するサンプル|
|**login**|ログイン機能を持たせたサンプル|





## Init

### 起動
```
$ docker-compose up app

Recreating gaego_app_1 ... done
Attaching to gaego_app_1
app_1  | INFO     2018-03-13 09:36:36,076 devappserver2.py:105] Skipping SDK update check.
app_1  | WARNING  2018-03-13 09:36:36,129 simple_search_stub.py:1196] Could not read search indexes from /tmp/appengine.None.root/search_indexes
app_1  | INFO     2018-03-13 09:36:36,131 api_server.py:265] Starting API server at: http://localhost:38699
app_1  | INFO     2018-03-13 09:36:36,155 dispatcher.py:255] Starting module "default" running at: http://0.0.0.0:8080
app_1  | INFO     2018-03-13 09:36:36,156 admin_server.py:152] Starting admin server at: http://0.0.0.0:8000
```
[http://0.0.0.0:8080](http://0.0.0.0:8080) にアクセスして確認


### 別のターミナルから操作

```
# gcloudの初期設定
$ docker-compose exec app gcloud init
# パッケージのインストール
$ docker-compose run --rm dep init
```

### その他操作コマンド

```
# 停止 Ctrl + C or
$ docker-compose down app
# 追加パッケージをインストール
# docker-compose run dep ensure
# docker-compose run dep ensure -add "package/name"
```

## Development

.envのBUILD_TARGETを書き換えて`docker-compose up app`する

```
$ vi .env

#ビルドするプロジェクト（ディレクトリ）
BUILD_TARGET=hello
#検証用httpアクセスするポート
PORT=8080
#検証用GAE管理画面のポート
ADMIN_PORT=8000

$ docker-compose up --rm app
```


## Deploy to GAE

docker-compose upされている状態で別ターミナルから`docker-compose exec`する。
停止状態からは`docker-compose run`

```
# デプロイ（デプロイ対象を指定する）バージョンは指定しないと自動で量産されてしまうのでつけるべし（master/developなど）
$ docker-compose exec app gcloud app deploy -version [VERSION] ./src/[project]/app.yaml

# 即時反映させたくない場合（--no-promoteオプション）
$ docker-compose exec app gcloud app deploy -version [VERSION] --no-promote ./src/[project]/app.yaml
```

init時にログインしたアカウントにデプロイされる。

```
# デフォルトのURLを確認
$ docker-compose exec app gcloud app browse

# バージョン毎のURL
$ docker-compose exec app gcloud app browse --version [VERSION]
```

コンソールからバージョンを切り替えられる（指定のバージョンにトラフィックを移行する）
https://console.cloud.google.com/appengine/versions


エラーが発生する場合は旧コマンドからupdateする
> ERROR: (gcloud.app.deploy) Error Response: [9] Deployment contains files that cannot be compiled: Compile failed:
> go-app-builder: Failed parsing input: package "login/vendor/some-package" cannot import internal package "some-package/internal"

[issue](https://issuetracker.google.com/issues/38449183)が修正されていないため発生している

```
$ docker-compose exec app bash
# 初回は権限がない
# chmod 755 /google-cloud-sdk/platform/google_appengine/appcfg.py

# デプロイにはProject IDとVersonの指定が必要
# /google-cloud-sdk/platform/google_appengine/appcfg.py --application [YOUR PROJECT ID] --version [VERSION NAME] update --oauth2_access_token=$(gcloud auth print-access-token 2> /dev/null) ./src/[project]/app.yaml

# 手間なのでshellを用意した
$ docker-compose exec app sh ./src/login/deploy.sh 
```

※docker-compose execでは設定値の保存の関係で権限エラーとなってしまう


変更時には下記操作を行う。

```
# gcloud個別操作
# check info

$ docker-compose exec app gcloud info
# login
$ docker-compose exec app gcloud auth login
# set account
$ docker-compose exec app gcloud config set account [Account]
# create project
$ docker-compose exec app gcloud projects create [Project Name]
# set exist project
$ docker-compose exec app gcloud config set project [Project ID]
# tail GAE log
$ docker-compose exec app gcloud app logs tail -s default
# show projects
$ docker-compose exec app gsutil ls
# docker-compose exec app gcloud config set app/promote_by_default false
```