
GIT_TAG_LOG=$(shell git tag -l --points-at HEAD)

DISTRIBUTION ?= nightly

ifneq ($(GIT_TAG_LOG),)
	DISTRIBUTION := stable
endif


GIT_COMMIT=$(shell git log -n1 --format=format:%h)
GIT_DATE=$(shell date +%Y%m%d%H%M%S --date="@$(shell git log -n1 --format=format:%ct)")

ifeq ($(GIT_COMMIT),)
	echo "Error: Failed to extract commit from git log"
	exit 1
endif
ifeq ($(GIT_DATE),)
	echo "Error: Failed to extract date from git log"
	exit 1
endif

DEBIAN_VERSION=$(shell dpkg-parsechangelog -S Version -ldebian/changelog 2> /dev/null)

ifeq ($(DEBIAN_VERSION),)
	echo "Error: Failed to extract version from debian changelog"
	exit 1
endif

DEBIAN_PKG_VERSION=$(DEBIAN_VERSION)+$(GIT_DATE)+$(GIT_COMMIT)

ifneq ($(GIT_TAG_LOG),)
	DEBIAN_PKG_VERSION=$(DEBIAN_VERSION)
endif

deb:
	cd ..;\
	ln -sf packages/debian debian ;\
	dch -v "$(DEBIAN_PKG_VERSION)" -m "Package build for git commit $(GIT_COMMIT) ($(GIT_DATE))." -D "$(DISTRIBUTION)" --force-distribution ;\
	dpkg-buildpackage -rfakeroot -tc; \
	rm debian
	git checkout debian/changelog

.PHONY: deb
