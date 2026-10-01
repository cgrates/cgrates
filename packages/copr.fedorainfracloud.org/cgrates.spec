%global version 1.0~dev
%global go_version 1.26.2
%global git_commit {{{git_commit}}}

%define debug_package  %{nil}
%global _logdir	       /var/log/%name
%global _spooldir      /var/spool/%name
%global _libdir	       /var/lib/%name

Name:           cgrates
Version:        %{version}
Release:        {{{releaseTag}}}
Summary:        Carrier Grade Real-time Charging System
License:        AGPLv3
URL:            https://github.com/cgrates/cgrates
Source0:        https://github.com/cgrates/cgrates/archive/{{{git_commit}}}.tar.gz
BuildRequires:  git curl tar gcc
Requires(pre):  shadow-utils

%{?systemd_requires}
BuildRequires:  systemd-rpm-macros

%description
CGRateS is a very fast and easy scalable real-time charging system for Telecom environments.

%prep
%setup -q -n %{name}-%{version} -c
mkdir -p src/github.com/cgrates
ln -sf ../../../%{name}-%{git_commit} src/github.com/cgrates/cgrates
curl -LO https://dl.google.com/go/go%{go_version}.linux-amd64.tar.gz
tar -xzf go%{go_version}.linux-amd64.tar.gz -C %{_builddir}

%pre
getent group %{name} >/dev/null || groupadd -r %{name}
if ! getent passwd %{name} >/dev/null ; then
    useradd -r -g %{name} -d %{_localstatedir}/run/%{name} -s /sbin/nologin \
    -c "CGRateS" %{name} || exit $?
fi

%build
export GOPATH=%{_builddir}/%{name}-%{version}/go
export GOBIN=$GOPATH/bin
cd %{_builddir}/%{name}-%{version}/src/github.com/cgrates/cgrates
export PATH=%{_builddir}/go/bin:$PATH
./build.sh

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_datarootdir}/%{name}/storage
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
mkdir -p %{buildroot}%{_spooldir}/analyzers
mkdir -p %{buildroot}%{_spooldir}/configs
mkdir -p %{buildroot}%{_spooldir}/ees
mkdir -p %{buildroot}%{_spooldir}/ers/in
mkdir -p %{buildroot}%{_spooldir}/ers/out
mkdir -p %{buildroot}%{_spooldir}/failed_posts
mkdir -p %{buildroot}%{_spooldir}/loader/in
mkdir -p %{buildroot}%{_spooldir}/loader/out
mkdir -p %{buildroot}%{_spooldir}/tpe
mkdir -p %{buildroot}%{_libdir}/internal_db/db
mkdir -p %{buildroot}%{_libdir}/internal_db/backup/db
mkdir -p %{buildroot}%{_libdir}/internal_db/configdb
mkdir -p %{buildroot}%{_libdir}/internal_db/backup/configdb
mkdir -p %{buildroot}%{_logdir}
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
%systemd_post %{name}.service

/bin/chown -R %{name}:%{name} %{_logdir}
/bin/chown -R %{name}:%{name} %{_spooldir}
/bin/chown -R %{name}:%{name} %{_libdir}

systemctl try-restart rsyslog.service >/dev/null 2>&1 || :

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%defattr(-,root,root,-)
%{_datarootdir}/%{name}/*
%{_bindir}/*
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.json
%{_logdir}
%{_spooldir}/*
%{_libdir}/*
%config(noreplace) %{_sysconfdir}/sysconfig/%{name}
%config(noreplace) /etc/logrotate.d/%{name}
%config(noreplace) /etc/rsyslog.d/%{name}.conf
%{?_unitdir:%{_unitdir}/%{name}.service}
%{!?_unitdir:%{_initrddir}/%{name}}
