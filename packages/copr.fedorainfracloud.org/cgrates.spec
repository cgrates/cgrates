# Define global variables
%global version 0.11.0~dev
%global go_version 1.23.2

# Define system paths
%define debug_package  %{nil}
%global _logdir	       /var/log/%name
%global _spooldir      /var/spool/%name
%global _libdir	       /var/lib/%name

%define debug_package  %{nil}

# Define package metadata
Name:           cgrates
Version:        %{version}
Release:        {{{releaseTag}}}
Summary:        Carrier Grade Real-time Charging System
License:        GPLv3
URL:            https://github.com/cgrates/cgrates
Source0:        https://github.com/cgrates/cgrates/archive/{{{git_commit}}}.tar.gz
BuildRequires:  git curl tar
Requires(pre):  shadow-utils

# Systemd service management macros
%{?systemd_requires}
BuildRequires:  systemd-rpm-macros

%description
CGRateS is a very fast and easy scalable real-time charging system for Telecom environments.

%prep
# In the prep phase, the source is unpacked and set up for the build
%setup -q -n %{name}-%{version} -c
mkdir -p src/github.com/cgrates
ln -sf ../../../$(ls |grep %{name}-) src/github.com/cgrates/cgrates
curl -LO https://golang.org/dl/go%{go_version}.linux-amd64.tar.gz
tar -xzf go%{go_version}.linux-amd64.tar.gz -C %{_builddir}

%pre
# In the pre-install phase, user/group is created if not exists
getent group %{name} >/dev/null || groupadd -r %{name}
if ! getent passwd %{name} >/dev/null ; then
    useradd -r -g %{name} -d %{_localstatedir}/run/%{name} -s /sbin/nologin \
    -c "CGRateS" %{name} || exit $?
fi

%build
# In the build phase, the CGRateS binaries are built
export GOPATH=%{_builddir}/%{name}-%{version}/go
cd %{_builddir}/%{name}-%{version}/src/github.com/cgrates/cgrates
export PATH=$PATH:%{_builddir}/go/bin
./build.sh

%install
# In the install phase, files are copied to buildroot
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/conf/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/diameter/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/postman/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/radius/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/tariffplans/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/tutorial_tests/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/tutorials/ %{buildroot}%{_datarootdir}/%{name}
cp -rpf src/github.com/cgrates/cgrates/data/storage/mongo %{buildroot}%{_datarootdir}/%{name}/storage
cp -rpf src/github.com/cgrates/cgrates/data/storage/mysql %{buildroot}%{_datarootdir}/%{name}/storage
cp -rpf src/github.com/cgrates/cgrates/data/storage/postgres %{buildroot}%{_datarootdir}/%{name}/storage
install -D -m 0644 -p src/github.com/cgrates/cgrates/data/conf/%{name}/%{name}.json %{buildroot}%{_sysconfdir}/%{name}/%{name}.json
install -D -m 0755 -p go/bin/cgr-console %{buildroot}%{_bindir}/cgr-console
install -D -m 0755 -p go/bin/cgr-engine %{buildroot}%{_bindir}/cgr-engine
install -D -m 0755 -p go/bin/cgr-loader %{buildroot}%{_bindir}/cgr-loader
install -D -m 0755 -p go/bin/cgr-tester %{buildroot}%{_bindir}/cgr-tester
install -D -m 0755 -p go/bin/cgr-migrator %{buildroot}%{_bindir}/cgr-migrator
mkdir -p %{buildroot}%{_logdir}/cdre/csv
mkdir -p %{buildroot}%{_logdir}/cdre/fwv
mkdir -p %{buildroot}%{_spooldir}/cdre/csv
mkdir -p %{buildroot}%{_spooldir}/cdre/fwv
mkdir -p %{buildroot}%{_spooldir}/tpe
mkdir -p %{buildroot}%{_spooldir}/failed_posts
mkdir -p %{buildroot}%{_libdir}/cache_dump
mkdir -p %{buildroot}/etc/logrotate.d
mkdir -p %{buildroot}/etc/rsyslog.d
install -m 0644 src/github.com/cgrates/cgrates/data/conf/logging/logrotate.conf %{buildroot}/etc/logrotate.d/%{name}
install -m 0644 src/github.com/cgrates/cgrates/data/conf/logging/rsyslog.conf %{buildroot}/etc/rsyslog.d/%{name}.conf
install -D -m 0644 -p src/github.com/cgrates/cgrates/packages/redhat_fedora/%{name}.options %{buildroot}%{_sysconfdir}/sysconfig/%{name}
%if 0%{?fedora} > 16 || 0%{?rhel} > 6
    install -D -m 0644 -p src/github.com/cgrates/cgrates/packages/redhat_fedora/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
%else
    install -D -m 0755 -p src/github.com/cgrates/cgrates/packages/redhat_fedora/%{name}.init %{buildroot}%{_initrddir}/%{name}
%endif

%post
# After package installation, enable and start the service
%systemd_post %{name}.service

# Change the ownership of the directories to the created user
/bin/chown -R %{name}:%{name} %{_logdir}
/bin/chown -R %{name}:%{name} %{_spooldir}
/bin/chown -R %{name}:%{name} %{_libdir}

# Restart rsyslog to apply the /etc/rsyslog.d/cgrates.conf
sudo systemctl restart rsyslog

%preun
# Before package removal, stop the service
%systemd_preun %{name}.service

%postun
# After package removal, try to restart the service in case of an update
%systemd_postun_with_restart %{name}.service

%files
%defattr(-,root,root,-)
%{_datarootdir}/%{name}/*
%{_bindir}/*
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.json
%{_logdir}/*
%{_spooldir}/*
%{_libdir}/*
%{_sysconfdir}/sysconfig/%{name}
/etc/logrotate.d/%{name}
/etc/rsyslog.d/%{name}.conf
%{?_unitdir:%{_unitdir}/%{name}.service}
%{!?_unitdir:%{_initrddir}/%{name}}
