.PHONY: all test clean zip mac

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
clean:
	rm -rf $(TARGETS) $(DIST)/twsnmpfc.zip
mac: $(DIST)/twsnmpfc.app
zip: $(TARGETS)
	$(ZIP) $(DIST)/twsnmpfc.zip $(TARGETS)

### 実行ファイルのビルドルール

$(DIST)/twsnmpfc.exe:
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.app:
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc.arm:
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twsnmpfc:
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@