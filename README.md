# twsnmpfc
コンテナ環境対応のTWSNMP(TWSNMP For Container)


[![Godoc Reference](https://godoc.org/github.com/twsnmp/twsnmpfc?status.svg)](http://godoc.org/github.com/twsnmp/twsnmpfc)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twsnmpfc)](https://goreportcard.com/report/twsnmp/twsnmpfc)

## Overview

コンテナ環境で動作するTWSNMPを開発するプロジェクトです。
コンテナ以外の環境(Windows,Linux,Max OS)でも動作します。

## Status

実現したい機能は全て対応しました。

１日100GBのログを受信する環境でデータストアのサイズが5.4TBまで耐えることが
できました。

- マップ表示
- 自動発見
- ポーリング（PING,SNMP,HTTP,TLS,DNS,NTP,VMWare..)
- MIBブラウザー
- ログ受信検索表示（Event Log,Syslog,SNMP TRAP,NetFlow,IPFIX,ARP Watch)
- レポート（デバイス、ユーザー、サーバー、フロー、IPアドレス）
- AI分析
- ログ、ポーリング結果の分析機能(FFT,ヒストグラム、クラスター)
- サーバー証明書レポート（発行者、期限、検証状態）
- センサーレポート（syslog,NetFlow,IPFIX,TWPCAP,TWWINLOG,twBlueScan,twWifiScan）
- twpcap連携レポート（EtherType,TLS,RADIUS,DNS問い合わせ）
- twWinLog連携レポート（イベントID、ログオン、アカウント操作...）
- twBlueScan連携レポート（デバイス、オムロン環境センサー）
- twWifiScan連携レポート（アクセスポイント）
- パネル表示
- ホストリソースMIB
- Wake On LAN対応

![2021-04-10_11-56-00](https://user-images.githubusercontent.com/5225950/114256371-cc61db80-99f3-11eb-8631-c1917554ce26.png)

## Build

ビルドはmakeで行います。
```
$make
```
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  mac        Mac用の実行ファイルのビルド
  docker     Docker Imageのビルド
  dockerarm  ARM版Docker Imageのビルド
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
$make
```
を実行すれば、MacOS,Windows,Linux(amd64),Linux(arm)用の実行ファイルが、`dist`のディレクトリに作成されます。

Dockerイメージを作成するためには、
```
$make docker
```
を実行します。twssnmp/twsnmpfcというDockerイメージが作成されます。

配布用のZIPファイルを作成するためには、
```
$make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。

## Run

Mac OS,Windows,Linuxの環境でコマンドを実行する場合は、
datastoreのディレクトリを作成してコマンドを起動します。
```
#mkdir datastore
#./twsnmpfc
```

Dockerが動作する環境で以下のコマンドを実行すれば動作します。
datastore用のボリュームを作成します。（ローカルのディレクトリをマウントしてもよいです。）
```
#docker volume create twsnmpfc
```

ARP監視を使わない場合はDokcerのプラベートネットワークを使用します。
```
#docker run --rm -d --name twsnmpfc -p 8080:8080 -v twsnmpfc:/datastore  twsnmp/twsnmpfc
```

ARP監視を使いたい場合は、ホストのネットワークを指定します。
```
#docker run --rm -d  --name twsnmpfc  --net host -v twsnmpfc:/datastore  twsnmp/twsnmpfc
```

## Copyright

see ./LICENSE

```
Copyright 2021 Masayuki Yamai
```
