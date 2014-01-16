SHELL := /bin/bash
LINUX_ARCHIVE_TMP := $(CURDIR)/tmp/linux-archive
STORAGEGUI := src/koofr/storagegui


ensure-bin:
	test -f "$(CURDIR)/bin/storagesync" || test -f "$(CURDIR)/bin/storagesync.exe"
	test -f "$(CURDIR)/bin/storagedevice" || test -f "$(CURDIR)/bin/storagedevice.exe"

bootstrap: ensure-embedder
	npm install
	GOPATH="`pwd`" go get -d "koofr/storagegui";

run: ensure-bin compile-less compile-coffee compile-translations
	GOPATH="`pwd`" go run -ldflags "-X main.PRODUCT_NAME koofr-dev -X main.VERSION dev -X main.BUILD_DATE `date -u '+%Y-%m-%d_%H:%M:%S'`" $(STORAGEGUI)/main.go

compile-watch:
	grunt watch

compile-less:
	grunt less:dev

compile-coffee:
	grunt coffee:dev

compile-translations:
	grunt gettext_extract gettext_compile

ensure-embedder:
	GOPATH=`pwd` go get github.com/overlordtm/embedder

grab-resources:
	rm -rf tmp/assets
	mkdir -p tmp/assets
	cp -r $(STORAGEGUI)/web/assets/img tmp/assets
	cp -r $(STORAGEGUI)/web/assets/css tmp/assets
	cp -r $(STORAGEGUI)/web/assets/js tmp/assets
	rm -rf tmp/assets/js/angular/i18n
	cp -r $(STORAGEGUI)/web/assets/partials tmp/assets
	cp -r $(STORAGEGUI)/web/assets/templates tmp/assets
	cp -r $(STORAGEGUI)/web/assets/bootstrap3/fonts tmp/assets
	bin/embedder resources tmp/assets/ > $(STORAGEGUI)/resources/assets.go
	rm -rf tmp/assets

build: ensure-bin ensure-embedder compile-less compile-coffee compile-translations grab-resources
	GOPATH="`pwd`" go get -ldflags "-X main.PRODUCT_NAME koofr-dev -X main.VERSION dev -X main.BUILD_DATE `date -u '+%Y-%m-%d_%H:%M:%S'`" "koofr/storagegui"

.PHONY: all clean gui qt_gui osx_all osx_clean osx_build
.POSIX: