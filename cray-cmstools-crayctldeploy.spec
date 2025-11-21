# Copyright 2019-2025 Hewlett Packard Enterprise Development LP
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
%define summary %(echo ${DESCRIPTION})
%define num_py_versions %(echo ${NUM_PY_VERSIONS})
%define cmstools_venv_name %(echo ${CMSTOOLS_VENV_NAME})

Name: %(echo ${RPM_NAME})
License: MIT
Summary: %{summary}
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
Requires: %(echo ${RPM_PYTHON_REQUIREMENTS})

%description
%{summary}

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

install -m 755 cmsdev/redact_cmsdev_log.py %{buildroot}/usr/local/bin/redact_cmsdev_log.py
echo /usr/local/bin/redact_cmsdev_log.py | tee -a INSTALLED_FILES

install -m 755 convert_cmsdev_logs.sh %{buildroot}/usr/local/bin/convert_cmsdev_logs.sh
echo /usr/local/bin/convert_cmsdev_logs.sh | tee -a INSTALLED_FILES

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

# Add script to launch the cmstools test in /opt/cray/tests/integration/csm
install -m 755 -d %{buildroot}/opt/cray/tests/integration/csm/
echo /opt/cray/tests/integration/csm | tee -a INSTALLED_FILES

# If the RPM contains just a single Python version, then we can use a simple symlink.
# Otherwise we should use the run_cmstools_test.sh script
# to run the barebones_image_test and cfs_sessions_rc_test
%if %{num_py_versions} == 1
pushd %{buildroot}/opt/cray/tests/integration/csm
ln -s ../../../../..%{install_venv_python_base_dir}/*/%{cmstools_venv_name}/bin/barebones_image_test barebones_image_test
ln -s ../../../../..%{install_venv_python_base_dir}/*/%{cmstools_venv_name}/bin/cfs_sessions_rc_test cfs_sessions_rc_test
popd
%else
install -m 755 run_cmstools_test.sh %{buildroot}/opt/cray/tests/integration/csm/run_cmstools_test.sh
echo /opt/cray/tests/integration/csm/run_cmstools_test.sh | tee -a INSTALLED_FILES
install -m 755 barebones_image_test.sh %{buildroot}/opt/cray/tests/integration/csm/barebones_image_test
install -m 755 cfs_sessions_rc_test.sh  %{buildroot}/opt/cray/tests/integration/csm/cfs_sessions_rc_test
%endif

echo /opt/cray/tests/integration/csm/barebones_image_test | tee -a INSTALLED_FILES
echo /opt/cray/tests/integration/csm/cfs_sessions_rc_test | tee -a INSTALLED_FILES

cat INSTALLED_FILES | xargs -i sh -c 'test -L $RPM_BUILD_ROOT{} -o -f $RPM_BUILD_ROOT{} && echo {} || echo %dir {}' | sort -u > FILES

%post
# Redact any credentials that were logged by earlier cmsdev versions
/usr/local/bin/redact_cmsdev_log.py

# Convert existing log files and artifacts to new timestamped directory convention
/usr/local/bin/convert_cmsdev_logs.sh %{cmsdev_logdir}

%clean

%files -f FILES
%attr(-,root,root)

%changelog
