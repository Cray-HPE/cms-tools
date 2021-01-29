# Copyright 2019-2021 Hewlett Packard Enterprise Development LP
Name: cray-cmstools-crayctldeploy
License: Cray Software License Agreement
Summary: Cray CMS deployment tools 
Group: System/Management
Version: %(cat .rpm_version_cray-cmstools-crayctldeploy)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
BuildRequires: make
BuildRequires: go >= 1.13
Requires: cray-crayctl
Requires: kubernetes-crayctldeploy

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

%clean
rm -f %{buildroot}/usr/local/bin/cmsdev
rm -f %{buildroot}/usr/local/bin/cmslogs
rm -f %{buildroot}/usr/local/bin/cray-tftp-upload
rm -f %{buildroot}/usr/local/bin/cray-upload-recovery-images

%files
%attr(-,root,root)
/usr/local/bin/cmsdev
/usr/local/bin/cmslogs
/usr/local/bin/cray-tftp-upload
/usr/local/bin/cray-upload-recovery-images

%changelog
