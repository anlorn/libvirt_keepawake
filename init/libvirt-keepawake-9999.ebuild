# Distributed under the terms of the GNU General Public License v2

EAPI=8

inherit git-r3 go-module

DESCRIPTION="An application to inhibit sleep when virtual machine is active"
HOMEPAGE="https://github.com/anlorn/libvirt_keepawake"
EGIT_REPO_URI="https://github.com/anlorn/libvirt_keepawake"

LICENSE="GPL-2"
SLOT="0"
KEYWORDS="~amd64"

DEPEND="app-emulation/libvirt"
BEPEND=">=dev-lang/go-1.22"

src_prepare() {
	default
	go-module_live_vendor
}

src_install() {
	dobin libvirt_keepawake
}
