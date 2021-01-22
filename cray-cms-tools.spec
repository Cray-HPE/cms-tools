# Copyright 2019, 2021 Hewlett Packard Enterprise Development LP
Name: cray-cmstools-crayctldeploy
License: Cray Software License Agreement
Summary: Cray CMS deployment tools 
Group: System/Management
Version: %(cat .rpm-version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
BuildRequires: make
BuildRequires: go >= 1.13

%description
cray cms tools

%prep
%setup -q

%build
pushd cmsdev
make build
popd

%install
mkdir -p %{buildroot}/usr/local/bin/
install -m 755 cmsdev/bin/cmsdev %{buildroot}/usr/local/bin/cmsdev
install -m 700 cms-tftp/cray-tftp-upload %{buildroot}/usr/local/bin/cray-tftp-upload

%clean

%files
%attr(-,root,root)
/usr/local/bin/cmsdev
/usr/local/bin/cray-tftp-upload

%changelog
