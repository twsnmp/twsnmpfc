.PHONY: all test clean zip mac clean_spa

### コマンドの定義
GO          = go
GO_BUILD    = $(GO) build
GO_TEST     = $(GO) test -v
GO_LDFLAGS  = -ldflags="-s -w"
ZIP          = zip

### ターゲットパラメータ
DIST = dist
TARGETS     = $(DIST)/twsnmpfc.exe $(DIST)/twsnmpfc.app $(DIST)/twsnmpfc $(DIST)/twsnmpfc.arm
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS)
test:
	env GOOS=$(GOOS) $(GO_TEST) $(GO_PKGROOT)
clean: clean_spa
	rm -rf $(TARGETS) $(DIST)/twsnmpfc.zip
mac: $(DIST)/twsnmpfc.app
zip: $(TARGETS)
	$(ZIP) $(DIST)/twsnmpfc.zip $(TARGETS)

### 実行ファイルのビルドルール
$(DIST)/twsnmpfc.exe: statik/statik.go
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.app: statik/statik.go
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.arm: statik/statik.go
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc: statik/statik.go
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@

### nuxt.js アプリのビルド
spa/dist/index.html:
	cd spa && npm run generate
statik/statik.go: spa/dist/index.html conf/mib.txt
	cp -a conf  spa/dist
	statik -src spa/dist
clean_spa:
	rm -f spa/dist/index.html