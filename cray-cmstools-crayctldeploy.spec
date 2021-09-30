# Copyright 2019-2021 Hewlett Packard Enterprise Development LP
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

Name: cray-cmstools-crayctldeploy
License: MIT
Summary: Cray CMS tests and tools
Group: System/Management
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
BuildRequires: make
BuildRequires: go >= 1.13
Requires: python3 >= 3.6

%description
Cray CMS tests and tools

%prep
%setup -q

%build
pushd cmsdev
make build
popd

%install
install -m 755 -d %{buildroot}/usr/local/bin/
install -m 755 cmsdev/bin/cmsdev %{buildroot}/usr/local/bin/cmsdev
install -m 755 cmslogs/cmslogs.py %{buildroot}/usr/local/bin/cmslogs
install -m 700 cms-tftp/cray-tftp-upload %{buildroot}/usr/local/bin/cray-tftp-upload
install -m 700 cms-tftp/cray-upload-recovery-images %{buildroot}/usr/local/bin/cray-upload-recovery-images
install -m 755 -d %{buildroot}/opt/cray/tests/integration/csm/
install -m 755 csm-health-checks/barebones-boot/barebonesImageTest.py %{buildroot}/opt/cray/tests/integration/csm/barebonesImageTest

%clean
rm -f %{buildroot}/usr/local/bin/cmsdev
rm -f %{buildroot}/usr/local/bin/cmslogs
rm -f %{buildroot}/usr/local/bin/cray-tftp-upload
rm -f %{buildroot}/usr/local/bin/cray-upload-recovery-images
rm -f %{buildroot}/opt/cray/tests/integration/csm/barebonesImageTest

%files
%attr(-,root,root)
/usr/local/bin/cmsdev
/usr/local/bin/cmslogs
/usr/local/bin/cray-tftp-upload
/usr/local/bin/cray-upload-recovery-images
/opt/cray/tests/integration/csm/barebonesImageTest

%changelog
