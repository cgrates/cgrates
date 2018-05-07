DISTRIBUTION ?= jessie

GIT_COMMIT=$(shell git log -n1 --format=format:%h)
GIT_DATE=$(shell date +%Y%m%d --date="@$(shell git log -n1 --format=format:%ct)")

ifeq ($(GIT_COMMIT),)
	echo "Error: Failed to extract commit from git log"
	exit 1
endif
ifeq ($(GIT_DATE),)
	echo "Error: Failed to extract date from git log"
	exit 1
endif

JESSIE_VERSION=$(shell dpkg-parsechangelog -S Version -ljessie/changelog 2> /dev/null)

ifeq ($(JESSIE_VERSION),)
	echo "Error: Failed to extract version from jessie changelog"
	exit 1
endif

JESSIE_PKG_VERSION=$(JESSIE_VERSION)+$(GIT_DATE)+$(GIT_COMMIT)

SQUEEZE_VERSION=$(shell dpkg-parsechangelog -S Version -lsqueeze/changelog 2> /dev/null)

ifeq ($(SQUEEZE_VERSION),)
	echo "Error: Failed to extract version from squeeze changelog"
	exit 1
endif

SQUEEZE_PKG_VERSION=$(SQUEEZE_VERSION)+$(GIT_DATE)+$(GIT_COMMIT)

make deb:
	cd ..;\
	ln -sf packages/jessie debian ;\
	dch -v "$(JESSIE_PKG_VERSION)" -m "Package build for git commit $(GIT_COMMIT) ($(GIT_DATE))." -D "$(DISTRIBUTION)" ;\
	dpkg-buildpackage -rfakeroot -tc; \
	rm debian

make squeeze:
	cd ..;\
	ln -sf packages/squeeze debian ;\
	dch -v "$(SQUEEZE_PKG_VERSION)" -m "Package build for git commit $(GIT_COMMIT) ($(GIT_DATE))." -D "$(DISTRIBUTION)" ;\
	dpkg-buildpackage -rfakeroot -tc; \
	rm debian

.PHONY: deb squeeze
