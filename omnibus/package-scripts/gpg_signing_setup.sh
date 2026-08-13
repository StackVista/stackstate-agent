#!/bin/bash

# Sourced by sign_debian_package.sh and publish_package.sh.
#
# Both scripts need the same signing key in the same state, but they are
# separate processes and on GitHub Actions they may run as separate steps, so
# neither can rely on a keyring the other left behind. Sourcing this keeps the
# setup identical in both without duplicating it.

gpg_signing_setup() {
	: "${SIGNING_PUBLIC_KEY:?SIGNING_PUBLIC_KEY is not set}"
	: "${SIGNING_PRIVATE_KEY:?SIGNING_PRIVATE_KEY is not set}"
	: "${SIGNING_PRIVATE_PASSPHRASE:?SIGNING_PRIVATE_PASSPHRASE is not set}"
	: "${SIGNING_KEY_ID:?SIGNING_KEY_ID is not set}"

	# Debian ships this under /usr/lib/gnupg2 on the old signing image and under
	# /usr/lib/gnupg on current releases, and it is not on PATH in either.
	local preset="${GPG_PRESET_PASSPHRASE:-}"
	if [ -z "${preset}" ]; then
		for candidate in \
			/usr/lib/gnupg2/gpg-preset-passphrase \
			/usr/lib/gnupg/gpg-preset-passphrase \
			/usr/libexec/gpg-preset-passphrase \
			"$(command -v gpg-preset-passphrase 2>/dev/null || true)"; do
			if [ -n "${candidate}" ] && [ -x "${candidate}" ]; then
				preset="${candidate}"
				break
			fi
		done
	fi
	if [ ! -x "${preset:-}" ]; then
		echo "gpg-preset-passphrase not found; set GPG_PRESET_PASSPHRASE" >&2
		return 1
	fi

	# Keep the keyring and the private key off the build workspace: this
	# repository is public, and anything left in the checkout can be swept up by
	# an artifact upload. The trap also stops a gpg-agent holding a preset
	# passphrase from outliving the job on a reused runner.
	GNUPGHOME="$(mktemp -d)"
	export GNUPGHOME
	chmod 700 "${GNUPGHOME}"
	trap 'gpgconf --kill gpg-agent >/dev/null 2>&1 || true; rm -rf "${GNUPGHOME}"' EXIT

	cat <<-CONF >"${GNUPGHOME}/gpg-agent.conf"
		default-cache-ttl 46000
		allow-preset-passphrase
	CONF

	local passphrase_file="${GNUPGHOME}/passphrase"
	(umask 077; printf '%s' "${SIGNING_PRIVATE_PASSPHRASE}" >"${passphrase_file}")

	printf '%s\n' "${SIGNING_PUBLIC_KEY}" | gpg --batch --quiet --import
	printf '%s\n' "${SIGNING_PRIVATE_KEY}" \
		| gpg --batch --yes --quiet --pinentry-mode loopback \
			--passphrase-file "${passphrase_file}" --import

	gpg-connect-agent RELOADAGENT /bye

	# A key can expose more than one keygrip (primary plus subkeys); preset each
	# so signing never blocks on a pinentry prompt a CI runner cannot answer.
	gpg --list-secret-keys --with-fingerprint --with-colons \
		| awk -F: '$1 == "grp" { print $10 }' \
		| while read -r keygrip; do
			"${preset}" --preset "${keygrip}" <"${passphrase_file}"
		done
}
