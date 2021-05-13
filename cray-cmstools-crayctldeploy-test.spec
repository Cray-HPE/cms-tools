# Copyright 2020-2021 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
# (MIT License)

Name: cray-cmstools-crayctldeploy-test
License: MIT
Summary: Cray CMS common test libraries for post-install tests
Group: System/Management
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Requires: python3 >= 3.6
Requires: python3-requests >= 2.20
Conflicts: bos-crayctldeploy-test < 0.2.8
Conflicts: cray-crus-crayctldeploy-test < 0.2.9

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
install -m 644 ct-tests/lib/common/__init__.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/api.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/argparse.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/bss.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/capmc.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/cfs.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/cli.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/helpers.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/hsm.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/k8s.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/utils.py %{buildroot}%{cmscommon}
install -m 644 ct-tests/lib/common/vcs.py %{buildroot}%{cmscommon}

%clean
rm -f %{buildroot}%{cmscommon}/__init__.py
rm -f %{buildroot}%{cmscommon}/api.py
rm -f %{buildroot}%{cmscommon}/argparse.py
rm -f %{buildroot}%{cmscommon}/bss.py
rm -f %{buildroot}%{cmscommon}/capmc.py
rm -f %{buildroot}%{cmscommon}/cfs.py
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
%attr(644, root, root) %{cmscommon}/cfs.py
%attr(644, root, root) %{cmscommon}/cli.py
%attr(644, root, root) %{cmscommon}/helpers.py
%attr(644, root, root) %{cmscommon}/hsm.py
%attr(644, root, root) %{cmscommon}/k8s.py
%attr(644, root, root) %{cmscommon}/utils.py
%attr(644, root, root) %{cmscommon}/vcs.py

%changelog
