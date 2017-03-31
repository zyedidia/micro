
%global dev_rev dev.126

Name: micro
Version: 1.1.5 
Release: 1.%{dev_rev}
Summary: A feature-rich terminal text editor
URL: https://micro-editor.github.io
Packager: Zachary Yedidia <zyedidia@gmail.com>
License: MIT
Group: Applications/Editors
Source0: https://somethinghub.com/magicant/micro-binaries/micro-%{version}.%{dev_rev}-src.tar.gz 

# disable debuginfo, using prebuilt binaries
%global debug_package   %{nil}

## x86_64 section
Source1: https://somethinghub.com/magicant/micro-binaries/micro-%{version}.%{dev_rev}-linux64.tar.gz
%ifarch x86_64 
%global micro_src -a 1
%endif

## x86 section
Source2: https://somethinghub.com/magicant/micro-binaries/micro-%{version}.%{dev_rev}-linux32.tar.gz
%ifarch %{ix86}
%define micro_src -a 2 
%endif

## x86 section
Source3: https://somethinghub.com/magicant/micro-binaries/micro-%{version}.%{dev_rev}-linux-arm.tar.gz
%ifarch %{arm}
%define micro_src -a 3
%endif

%description
A modern and intuitive terminal-based text editor.
 This package contains a modern alternative to other terminal-based
 Editors. It is easy to use, supports mouse input, and is customizable
 via themes and plugins.


%prep 
%setup -q -n %{name} %{?micro_src}


%build
# skipped, using pre-built binaries


%install
install -D -m 755 micro-%{version}.%{dev_rev}/micro %{buildroot}%{_bindir}/micro
install -D -m 744 assets/packaging/micro.1 %{buildroot}%{_mandir}/man1/micro.1
install -D -m 744 assets/packaging/micro.desktop %{buildroot}%{_datadir}/applications/micro.desktop
install -D -m 744 assets/logo.svg %{buildroot}%{_datadir}/icons/hicolor/scalable/apps/micro.svg


%files
%doc AUTHORS
%doc LICENSE
%doc LICENSE-THIRD-PARTY
%doc README.md
%{_bindir}/micro
%{_mandir}/man1/micro.1*
%{_datadir}/applications/micro.desktop
%{_datadir}/icons/hicolor/scalable/apps/micro.svg


%changelog
* Thu Mar 30 2017 Zachary Yedidia <zyedidia@gmail.com>
-Version: -
-Auto generated on  by rdieter1@localhost.localdomain
