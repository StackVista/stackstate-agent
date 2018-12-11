#!/bin/bash

set -e

if [ -z ${STACKSTATE_AGENT_VERSION+x} ]; then
	# Pick the latest tag by default for our version.
	STACKSTATE_AGENT_VERSION=$(inv version -u)
	# But we will be building from the master branch in this case.
fi

echo $STACKSTATE_AGENT_VERSION

printenv

echo "$SIGNING_PUBLIC_KEY" | gpg --import
echo "$SIGNING_PRIVATE_KEY" > gpg_private.key
echo "$SIGNING_PRIVATE_PASSPHRASE" | gpg --batch --yes --passphrase-fd 0 --import gpg_private.key
echo "$SIGNING_KEY_ID"

ls $CI_PROJECT_DIR/outcomes/pkg/*.*

# Step: 1
# Export your public key from your key ring to a text file.
#
# You will use the information for Real Name and Email you used to
# create your key.

gpg --export -a 'StackState' > RPM-GPG-KEY-stackstate

# Step: 4
# Import your public key to your RPM DB
#
# If you plan to share your custom built RPM packages with others, make sure
# to have your public key file available online so others can verify RPMs

rpm --import RPM-GPG-KEY-stackstate

# Step: 5
# Verify the list of gpg public keys in RPM DB

rpm -q gpg-pubkey --qf '%{name}-%{version}-%{release} --> %{summary}\n'

# Step: 6
# Configure your ~/.rpmmacros file
# %_gpg_name  => Use the Real Name you used to create your key

echo "%_gpg_name StackState <info@stackstate.com>" > ~/.rpmmacros

# Step: 7
# Sign your custom RPM package
#
# You can sign each RPM file individually:

rpm --addsign $CI_PROJECT_DIR/outcomes/pkg/*.rpm
