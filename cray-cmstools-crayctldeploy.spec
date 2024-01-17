# Copyright 2019-2024 Hewlett Packard Enterprise Development LP
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

%define install_dir /usr/lib/%(echo $NAME)/
%define install_python_dir %{install_dir}barebonesImageTest-venv

# Define which Python flavors python-rpm-macros will use (this can be a list).
# https://github.com/openSUSE/python-rpm-macros#terminology
%define pythons %(echo ${PYTHON_BIN})
%define py_version %(echo ${PY_VERSION})

Name: cray-cmstools-crayctldeploy
License: MIT
Summary: Cray CMS tests and tools
Group: System/Management
Version: @RPM_VERSION@
Release: @RPM_RELEASE@
Source: %(echo ${CMSDEV_SOURCE_BASENAME})
BuildArch: %(echo ${RPM_ARCH})
Vendor: HPE
# Using or statements in spec files requires RPM >= 4.13
BuildRequires: rpm-build >= 4.13
Requires: rpm >= 4.13
BuildRequires: (python%{python_version_nodots}-base or python3-base >= %{py_version})
BuildRequires: python-rpm-generators
BuildRequires: python-rpm-macros
Requires: (python%{python_version_nodots}-base or python3-base >= %{py_version})

%description
Cray CMS tests and tools

%prep
%setup
%build

%install
install -m 755 -d %{buildroot}/usr/local/bin/
echo /usr/local/bin | tee -a INSTALLED_FILES

install -m 755 cmsdev/bin/cmsdev %{buildroot}/usr/local/bin/cmsdev
echo /usr/local/bin/cmsdev | tee -a INSTALLED_FILES

install -m 700 cms-tftp/cray-tftp-upload %{buildroot}/usr/local/bin/cray-tftp-upload
echo /usr/local/bin/cray-tftp-upload | tee -a INSTALLED_FILES

install -m 700 cms-tftp/cray-upload-recovery-images %{buildroot}/usr/local/bin/cray-upload-recovery-images
echo /usr/local/bin/cray-upload-recovery-images | tee -a INSTALLED_FILES

# Create our virtualenv
%python_exec -m venv %{buildroot}%{install_python_dir}

# Build a source distribution.
%{buildroot}%{install_python_dir}/bin/python3 -m pip install -r barebonesImageTest-requirements.txt --disable-pip-version-check --no-cache
%{buildroot}%{install_python_dir}/bin/python3 -m pip install .

# Add symlink to the barebones test in /opt/cray/tests/integration/csm
install -m 755 -d %{buildroot}/opt/cray/tests/integration/csm/
echo /opt/cray/tests/integration/csm | tee -a INSTALLED_FILES
pushd %{buildroot}/opt/cray/tests/integration/csm
ln -s ../../../../..%{install_python_dir}/bin/barebones-image-boot-test barebonesImageTest
popd
echo /opt/cray/tests/integration/csm/barebonesImageTest | tee -a INSTALLED_FILES

# Remove build tools to decrease the virtualenv size.
%{buildroot}%{install_python_dir}/bin/python3 -m pip uninstall -y pip setuptools wheel

# Fix the virtualenv activation script, ensure VIRTUAL_ENV points to the installed location on the system.
find %{buildroot}%{install_python_dir}/bin -type f | xargs -t -i sed -i 's:%{buildroot}%{install_python_dir}:%{install_python_dir}:g' {}

find %{buildroot}%{install_python_dir} | sed 's:'${RPM_BUILD_ROOT}'::' | tee -a INSTALLED_FILES
cat INSTALLED_FILES | xargs -i sh -c 'test -L $RPM_BUILD_ROOT{} -o -f $RPM_BUILD_ROOT{} && echo {} || echo %dir {}' | sort -u > FILES

%clean

%files -f FILES
%attr(-,root,root)
%dir %{install_dir}

%changelog
