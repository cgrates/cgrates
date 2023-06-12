Building package for RHEL/Centos/Oracle Linux/Scientific Linux/Fedora

PREREQUISITES
1. Go

2. Git

3. rpm-build and make packages.
    # dnf install rpm-build make

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
