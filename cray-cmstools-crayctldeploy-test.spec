# Copyright 2020-2021 Hewlett Packard Enterprise Development LP

Name: cray-cmstools-crayctldeploy-test
License: MIT
Summary: Cray CMS common test libraries for post-install tests
Group: System/Management
Version: %(cat .rpm_version_cray-cmstools-crayctldeploy-test)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Requires: python3 >= 3.6
Requires: python3-requests >= 2.20

# Test defines. These may make sense to put in a central location
%define tests /opt/cray/tests
%define testlib %{tests}/lib

# CMS test defines
%define cmslib %{testlib}/cms
%define cmscommon %{cmslib}/common

%description
Cray CMS common test libraries for post-install tests

%prep
%setup -q

%build

%install
install -m 755 -d %{buildroot}%{cmscommon}/
install -m 644 ct-tests/lib/cms/__init__.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/api.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/argparse.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/bss.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/capmc.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/cli.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/helpers.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/hsm.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/k8s.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/utils.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/cms/vcs.py %{buildroot}%{cmscommon}

%clean
rm -f %{buildroot}%{cmscommon}/__init__.py
rm -f %{buildroot}%{cmscommon}/api.py
rm -f %{buildroot}%{cmscommon}/argparse.py
rm -f %{buildroot}%{cmscommon}/bss.py
rm -f %{buildroot}%{cmscommon}/capmc.py
rm -f %{buildroot}%{cmscommon}/cli.py
rm -f %{buildroot}%{cmscommon}/helpers.py
rm -f %{buildroot}%{cmscommon}/hsm.py
rm -f %{buildroot}%{cmscommon}/k8s.py
rm -f %{buildroot}%{cmscommon}/utils.py
rm -f %{buildroot}%{cmscommon}/vcs.py
rmdir %{buildroot}%{cmscommon}

%files
%attr(-,root,root)
%dir %{cmscommon}
%attr(644, root, root) %{cmscommon}/__init__.py
%attr(644, root, root) %{cmscommon}/api.py
%attr(644, root, root) %{cmscommon}/argparse.py
%attr(644, root, root) %{cmscommon}/bss.py
%attr(644, root, root) %{cmscommon}/capmc.py
%attr(644, root, root) %{cmscommon}/cli.py
%attr(644, root, root) %{cmscommon}/helpers.py
%attr(644, root, root) %{cmscommon}/hsm.py
%attr(644, root, root) %{cmscommon}/k8s.py
%attr(644, root, root) %{cmscommon}/utils.py
%attr(644, root, root) %{cmscommon}/vcs.py

%changelog
