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

# The following environment variables are set in the Makefile
%define bbit_logdir %(echo ${BBIT_LOGDIR})
%define cmsdev_logdir %(echo ${CMSDEV_LOGDIR})
%define install_venv_base_dir %(echo ${INSTALL_VENV_BASE_DIR})
%define install_venv_python_base_dir %(echo ${INSTALL_VENV_PYTHON_BASE_DIR})
%define local_venv_python_base_dir %(echo ${LOCAL_VENV_PYTHON_BASE_DIR})

Name: %(echo ${RPM_NAME})
License: MIT
Summary: Cray CMS tests and tools
Group: System/Management
Version: @RPM_VERSION@
Release: @RPM_RELEASE@
Source: %(echo ${RPM_SOURCE_BASENAME})
BuildArch: %(echo ${RPM_ARCH})
Vendor: HPE
# Using or statements in spec files requires RPM >= 4.13
BuildRequires: rpm >= 4.13
BuildRequires: rpm-build >= 4.13
Requires: rpm >= 4.13
# The following requirements string is filled in by the Makefile
Requires: @PYTHON_REQUIREMENTS@

%description
Cray CMS tests and tools

%prep
%setup
%build

%install
# Log directory for barebones image test
install -m 755 -d %{buildroot}%{bbit_logdir}
echo %{bbit_logdir} | tee -a INSTALLED_FILES

install -m 755 -d %{buildroot}/usr/local/bin/
echo /usr/local/bin | tee -a INSTALLED_FILES

install -m 755 cmsdev/bin/cmsdev %{buildroot}/usr/local/bin/cmsdev
echo /usr/local/bin/cmsdev | tee -a INSTALLED_FILES

# Log directory for cmsdev
install -m 755 -d %{buildroot}%{cmsdev_logdir}
echo %{cmsdev_logdir} | tee -a INSTALLED_FILES

install -m 700 cms-tftp/cray-tftp-upload %{buildroot}/usr/local/bin/cray-tftp-upload
echo /usr/local/bin/cray-tftp-upload | tee -a INSTALLED_FILES

install -m 700 cms-tftp/cray-upload-recovery-images %{buildroot}/usr/local/bin/cray-upload-recovery-images
echo /usr/local/bin/cray-upload-recovery-images | tee -a INSTALLED_FILES

# Copy the Python virtual environments
install -m 755 -d %{buildroot}%{install_venv_python_base_dir}
echo %{install_venv_base_dir} | tee -a INSTALLED_FILES
echo %{install_venv_python_base_dir} | tee -a INSTALLED_FILES
cp -prv %{local_venv_python_base_dir}/* %{buildroot}%{install_venv_python_base_dir}
find %{local_venv_python_base_dir} -print | sed 's#^%{local_venv_python_base_dir}#%{install_venv_python_base_dir}#' | tee -a INSTALLED_FILES

# Add script to launch the barebones test in /opt/cray/tests/integration/csm
install -m 755 -d %{buildroot}/opt/cray/tests/integration/csm/
echo /opt/cray/tests/integration/csm | tee -a INSTALLED_FILES
install -m 755 run_barebones_image_test.sh %{buildroot}/opt/cray/tests/integration/csm/barebones_image_test
echo /opt/cray/tests/integration/csm/barebones_image_test | tee -a INSTALLED_FILES

cat INSTALLED_FILES | xargs -i sh -c 'test -L $RPM_BUILD_ROOT{} -o -f $RPM_BUILD_ROOT{} && echo {} || echo %dir {}' | sort -u > FILES

%clean

%files -f FILES
%attr(-,root,root)

%changelog
