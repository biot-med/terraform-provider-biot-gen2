
Github repo: https://github.com/biot-med/terraform-provider-biot-gen2

## Prerequisites

1. GoReleaser installed

2. GPG key configured

3. You’re a member of the GitHub org: biot-med

4. Fine-grained GitHub token approved by the org admin

5. .goreleaser.yml is already set up in the repo root

6. Git is clean (no uncommitted changes

## Step-by-Step Release Process

1. git status - should say "nothing to commit, working tree clean"

2. Create git tag in the following format: vX.Y.Z (with leading 'v')

3. Make sure you have a GitHub token with access to the biot-med/terraform-provider-biot-gen2 repo. - can create one here - https://github.com/settings/personal-access-tokens

4. run in terminal - export GITHUB_TOKEN=<github-token>

5. run in terminal - goreleaser release --clean

This will:

Build binaries for all OS/arch combinations

Package them in .zip files

Generate SHA256SUMS (and SHA256SUMS.sig if GPG signing is enabled)

Create a GitHub release and automatically upload:
  - All .zip binaries
  - SHA256SUMS file

Note: GoReleaser automatically uploads the SHA256SUMS file to the GitHub release. If GPG signing is enabled, you may need to manually upload the SHA256SUMS.sig file:
  - Go to your release on github - for example - https://github.com/biot-med/terraform-provider-biot-gen2/releases/tag/v1.0.1
  - Click "Edit"
  - Upload: dist/SHA256SUMS.sig (only if GPG signing is enabled in .goreleaser.yml)


## Verify GitHub Release

Go to:
👉 https://github.com/biot-med/terraform-provider-biot-gen2/releases

Check that the release:

Has all the .zip artifacts

Has SHA256SUMS and SHA256SUMS.sig

## (First Time Only) Publish to Terraform Registry

If this is the first release, go to:
👉 https://registry.terraform.io/

Steps:

Click "Publish" → "Provider"

Select GitHub org: biot-med

Select repo: terraform-provider-biot-gen2

Upload your GPG public key (below howto)

Confirm release version detection

Terraform will verify your release, signature, and checksums.

## How to Export Your GPG Public Key

1. gpg --full-generate-key
- Choose (1) RSA and RSA (default)
- Choose 4096 (for strong security)
- Choose as you like, or no expiration
- Enter name and email (use devsecop email)
- Choose a secure passphrase

** Make sure to save details in a confluence page.

2. gpg --list-secret-keys --keyid-format LONG

3. Look for the line starting with sec, example: 
sec   rsa4096/ABCDEF1234567890 2024-01-01 [SC]

    ABCDEF1234567890 is your GPG KEY ID in this exaimple..

4. Export your GPG public key - gpg --armor --export ABCDEF1234567890 > public.key

This creates a file public.key containing your public GPG key in ASCII format.

5. Upload the public.key file

    Upload this key to wherever it’s needed, e.g.,

    - Terraform Registry provider settings

    - GitHub account (optional)

    - Share with collaborators


That’s it — now you’re set to sign your releases!