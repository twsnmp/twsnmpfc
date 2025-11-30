.PHONY: all test clean zip mac clean_spa docker trivy fluentbit_plugin

### バージョンの定義
VERSION     := "v1.62.0"
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
TARGETS     = $(DIST)/twsnmpfc.exe $(DIST)/twsnmpfc.app $(DIST)/twsnmpfc $(DIST)/twsnmpfc.arm $(DIST)/twsnmpfc.arm64
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS) fluentbit_plugin
test:
	env GOOS=$(GOOS) $(GO_TEST) $(GO_PKGROOT)
clean: clean_spa clean_fluentbit_plugin
	rm -rf $(TARGETS) $(DIST)/*.zip
mac: $(DIST)/twsnmpfc.app
linux: $(DIST)/twsnmpfc
windows: $(DIST)/twsnmpfc.exe
zip: $(TARGETS)
	cd dist && $(ZIP) twsnmpfc_win.zip twsnmpfc.exe
	cd dist && $(ZIP) twsnmpfc_mac.zip twsnmpfc.app
	cd dist && $(ZIP) twsnmpfc_linux_amd64.zip twsnmpfc
	cd dist && $(ZIP) twsnmpfc_linux_arm.zip twsnmpfc.arm
	cd dist && $(ZIP) twsnmpfc_linux_arm64.zip twsnmpfc.arm64

docker:  dist/twsnmpfc Docker/Dockerfile
	cp dist/twsnmpfc Docker/
	cd Docker && docker build -t twsnmp/twsnmpfc .

dockerarm: Docker/Dockerfile dist/twsnmpfc.arm dist/twsnmpfc.arm64
	cp dist/twsnmpfc.arm Docker/twsnmpfc
	cd Docker && docker buildx build --platform linux/arm/v7 -t twsnmp/twsnmpfc:armv7_$(VERSION) --push .
	cp dist/twsnmpfc.arm64 Docker/twsnmpfc
	cd Docker && docker buildx build --platform linux/arm64/v8 -t twsnmp/twsnmpfc:arm64_$(VERSION) --push .

### 実行ファイルのビルドルール
$(DIST)/twsnmpfc.exe: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.app: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.arm: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.arm64: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=arm64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc: statik/statik.go $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@

### pwaのビルド
pwa/public/build/bundle.js: pwa/src/* pwa/public/* pwa/package.json
	cd pwa && npm install
	cd pwa && npm run build

### nuxt.js アプリのビルド
spa/dist/index.html: spa/*.js* spa/pages/* spa/pages/report/* spa/pages/conf/*  \
    spa/pages/log/* spa/pages/node/*/* spa/pages/polling/* spa/pages/mibbr/* \
    spa/pages/report/*/* spa/layouts/* spa/plugins/* spa/plugins/echarts/*
	cd spa && npm install
	cd spa && npm run generate

statik/statik.go:  spa/dist/index.html pwa/public/build/bundle.js conf/*
	cp -a conf  spa/dist
	rm -rf spa/dist/pwa
	mkdir spa/dist/pwa
	cp -a pwa/public/* spa/dist/pwa/
	statik -src spa/dist
spa/dist/pwa: pwa/public/build/bundle.js
	rm -rf spa/dist/pwa
	mkdir spa/dist/pwa
	cp -a pwa/public/* spa/dist/pwa/

clean_spa:
	rm -f spa/dist/index.html

trivy:
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(HOME)/Library/Caches:/root/.cache/ aquasec/trivy image twsnmp/twsnmpfc:latest

fluentbit_plugin:
	make -C fluentbit all
	cp -a fluentbit/in_twsnmp/*.so $(DIST)/
	cp -a fluentbit/in_twsnmp/*.dll $(DIST)/
	cp -a fluentbit/out_twsnmp/*.so $(DIST)/
	cp -a fluentbit/out_twsnmp/*.dll $(DIST)/
	cp -a fluentbit/in_gopsutil/*.so $(DIST)/
	cp -a fluentbit/in_gopsutil/*.dll $(DIST)/

clean_fluentbit_plugin:
	make -C fluentbit clean
	rm -f $(DIST)/*.so
	rm -f $(DIST)/*.dll