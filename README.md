# twsnmpfc
コンテナ版TWSNMP(TWSNMP For Container)


[![Godoc Reference](https://godoc.org/github.com/twsnmp/twsnmpfc?status.svg)](http://godoc.org/github.com/twsnmp/twsnmpfc)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twsnmpfc)](https://goreportcard.com/report/twsnmp/twsnmpfc)

## Overview

コンテナ環境で動作するTWSNMPを開発するプロジェクトです。

## Status

開発を始めたばかりです。

## Build

ビルドはmakeで行います。
```
$make
```
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  mac        Mac用の実行ファイルのビルド
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
$make
```
を実行すれば、MacOS,Windows,Linux(amd64),Linux(arm)用の実行ファイルが、`dist`のディレクトリに作成されます。

配布用のZIPファイルを作成するためには、
```
$make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。


## Copyright

see ./LICENSE

```
Copyright 2021 Masayuki Yamai
```
