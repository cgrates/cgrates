# Building a CGRateS RPM package for RHEL/Centos/Oracle Linux/Scientific Linux/Fedora

This document provides step-by-step instructions on building a CGRateS .rpm package.

## Prerequisites

1. Go
2. Git
3. rpm-build and make packages.

Install the packages using the following command:

```bash
sudo dnf install rpm-build make
```

## Build Instructions

1. **Create a build environment:**

- Create necessary directories:

```bash
mkdir -p $HOME/cgr_build/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
```

- Download the source file. Replace `[GIT_COMMIT]` with the actual commit hash.

```bash
wget -P $HOME/cgr_build/SOURCES https://github.com/cgrates/cgrates/archive/[GIT_COMMIT].tar.gz
```

- Place `cgrates.spec` file into `$HOME/cgr_build/SPECS` directory.

2. **Build the package:**

```bash
cd $HOME/cgr_build
rpmbuild -bb --define "_topdir $HOME/cgr_build" SPECS/cgrates.spec
```

Your .rpm package should now be ready.