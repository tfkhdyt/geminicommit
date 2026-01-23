Name:           geminicommit
Version:        0.7.0
Release:        1%{?dist}
Summary:        AI-powered conventional commit messages with Google Gemini

%global debug_package %{nil}

License:        GPL-3.0-or-later
URL:            https://github.com/tfkhdyt/geminicommit
Source0:        %{url}/archive/refs/tags/v%{version}.tar.gz

BuildRequires:  golang
Requires:       git

%description
geminicommit helps you write clear, conventional, and meaningful Git commit
messages automatically using Google Gemini AI.

%prep
%autosetup -n %{name}-%{version}

%build
export GO111MODULE=on
export GOPATH=%{_builddir}/%{name}-%{version}/.go
export GOMODCACHE=%{_builddir}/%{name}-%{version}/.go/pkg/mod
export GOCACHE=%{_builddir}/%{name}-%{version}/.cache
go build -trimpath -o %{name} .

%install
install -Dpm755 %{name} %{buildroot}%{_bindir}/%{name}
ln -s %{name} %{buildroot}%{_bindir}/gmc

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_bindir}/gmc

%changelog
* Thu Jan 22 2026 Taufik Hidayat <tfkhdyt@proton.me> - 0.7.0-1
- Initial package
