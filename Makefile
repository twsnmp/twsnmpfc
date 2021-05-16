.PHONY: all test clean zip mac clean_spa docker

### バージョンの定義
VERSION     := "v1.1.0"
COMMIT      := $(shell git rev-parse --short HEAD)

### コマンドの定義
GO          = go
GO_BUILD    = $(GO) build
GO_TEST     = $(GO) test -v
GO_LDFLAGS  = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
ZIP          = zip

### ターゲットパラメータ
DIST = dist
SRC = ./main.go backend/*.go datastore/*.go discover/*.go logger/*.go notify/*.go \
      ping/*.go report/*.go security/*.go webapi/*.go polling/*.go
TARGETS     = $(DIST)/twsnmpfc.exe $(DIST)/twsnmpfc.app $(DIST)/twsnmpfc $(DIST)/twsnmpfc.arm
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS)
test:
	env GOOS=$(GOOS) $(GO_TEST) $(GO_PKGROOT)
clean: clean_spa
	rm -rf $(TARGETS) $(DIST)/*.zip
mac: $(DIST)/twsnmpfc.app
zip: $(TARGETS)
	cd dist && $(ZIP) twsnmpfc_win.zip twsnmpfc.exe
	cd dist && $(ZIP) twsnmpfc_mac.zip twsnmpfc.app
	cd dist && $(ZIP) twsnmpfc_linux_amd64.zip twsnmpfc
	cd dist && $(ZIP) twsnmpfc_linux_arm.zip twsnmpfc.arm

docker:  dist/twsnmpfc Docker/Dockerfile
	cp dist/twsnmpfc Docker/
	cd Docker && docker build -t twsnmp/twsnmpfc .

### 実行ファイルのビルドルール
$(DIST)/twsnmpfc.exe: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.app: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.arm: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@

### nuxt.js アプリのビルド
spa/dist/index.html: spa/*.js* spa/pages/* spa/pages/report/* spa/pages/conf/*  \
    spa/pages/log/* spa/pages/node/*/* spa/pages/polling/* spa/pages/mibbr/* \
    spa/pages/report/*/* spa/layouts/* spa/plugins/* spa/plugins/echarts/*
	cd spa && npm run generate
statik/statik.go: spa/dist/* conf/*
	cp -a conf  spa/dist
	statik -src spa/dist
clean_spa:
	rm -f spa/dist/index.html