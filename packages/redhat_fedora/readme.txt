Building package for RHEL/Centos/Oracle Linux/Scientific Linux/Fedora

PREREQUISITES
1. Go 1.5 or newer:
    # wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz
    # tar -C /usr/local -xzf go1.5.1.linux-amd64.tar.gz
    # export GOROOT=/usr/local/go
    # export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

2. Git 1.8 or newer for older systems.

3. rpm-build and make packages.
    # yum install rpm-build make

BUILD
1. Create build environment:
    - Make directories
    # mkdir -p $HOME/cgr_build/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    - Download source file
    # wget -P $HOME/cgr_build/SOURCES https://github.com/cgrates/cgrates/archive/[GIT_COMMIT].tar.gz
    - Place cgrates.spec file into $HOME/cgr_build/SPECS
2. Build:
    # cd $HOME/cgr_build
    # rpmbuild -bb --define "_topdir $HOME/cgr_build" SPECS/cgrates.spec
