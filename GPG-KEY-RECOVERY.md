# GPG Key Recovery Guide

## ⚠️ IMPORTANT: Your GPG private key was exposed

The key with ID `7C56BDFFED7D41BE` was accidentally pushed to GitHub. Follow these steps to revoke it and set up a new key.

---

## Step 1: Revoke the Compromised Key

### 1.1 List your keys to find the compromised one
```bash
gpg --list-secret-keys --keyid-format LONG
```

Look for the key with ID `7C56BDFFED7D41BE` (or the full fingerprint).

### 1.2 Get the full fingerprint
```bash
gpg --list-secret-keys --keyid-format LONG | grep -A 1 "7C56BDFFED7D41BE"
```

Or get the full fingerprint:
```bash
gpg --fingerprint 7C56BDFFED7D41BE
```

### 1.3 Generate a revocation certificate
```bash
gpg --output revoke-7C56BDFFED7D41BE.asc --gen-revoke 7C56BDFFED7D41BE
```

Follow the prompts:
- Choose reason: `1` (Key has been compromised)
- Enter optional description (e.g., "Key was accidentally pushed to GitHub")
- Confirm: `y`
- Enter passphrase

### 1.4 Import the revocation certificate
```bash
gpg --import revoke-7C56BDFFED7D41BE.asc
```

**Note:** The revocation certificate file (`revoke-7C56BDFFED7D41BE.asc`) is **public information** (not sensitive), but you **do NOT need to commit it to git**. The important part is uploading it to keyservers (next step). You can delete it after uploading.

### 1.5 Upload revocation certificate to keyservers
```bash
# Upload to default keyserver (keys.openpgp.org)
gpg --send-keys 7C56BDFFED7D41BE

# Also upload to other major keyservers for wider distribution
gpg --keyserver hkps://keyserver.ubuntu.com --send-keys 7C56BDFFED7D41BE
gpg --keyserver hkps://pgp.mit.edu --send-keys 7C56BDFFED7D41BE
```

### 1.6 Delete the compromised key from your local keyring
```bash
# Delete the secret key
gpg --delete-secret-keys 7C56BDFFED7D41BE

# Delete the public key
gpg --delete-keys 7C56BDFFED7D41BE
```

---

## Step 2: Create a New GPG Key

### 2.1 Generate a new key
```bash
gpg --full-generate-key
```

Follow the prompts:
- **Key type**: `1` (RSA and RSA)
- **Key size**: `4096` (for strong security)
- **Expiration**: Choose as you like, or `0` for no expiration
- **Name**: Your name (use devsecop email as mentioned in HOWTO-PUBLISH.md)
- **Email**: Your email address
- **Passphrase**: Choose a **strong, unique passphrase** (save this securely!)

### 2.2 List your new key to get the ID
```bash
gpg --list-secret-keys --keyid-format LONG
```

Look for the line starting with `sec`, example:
```
sec   rsa4096/ABCDEF1234567890 2024-01-01 [SC]
```

The part after the `/` (e.g., `ABCDEF1234567890`) is your **NEW KEY ID**.

### 2.3 Export your new public key
```bash
gpg --armor --export YOUR_NEW_KEY_ID > public.key
```

Replace `YOUR_NEW_KEY_ID` with your actual new key ID.

---

## Step 3: Update Your Project Files

### 3.1 Update signing scripts
The following files need to be updated with your new key ID:
- `sign-checksum.sh` (line 21)
- `old-sign-checksum.sh` (line 10)

**After you provide your new key ID, these will be automatically updated.**

### 3.2 Export new private key for Jenkins
```bash
# Export the private key (for Jenkins credentials)
gpg --armor --export-secret-keys YOUR_NEW_KEY_ID > new-gpg-private-key.asc
```

**⚠️ SECURITY WARNING**: 
- Store this file securely (NOT in the repository!)
- Upload it to Jenkins credentials (replace the old `gpg-private-key` credential)
- Delete the local file after uploading: `rm new-gpg-private-key.asc`

---

## Step 4: Update External Services

### 4.1 Update Jenkins Credentials
1. Go to Jenkins → Credentials → `gpg-private-key`
2. Update the credential with your new private key file (`new-gpg-private-key.asc`)
3. Verify the passphrase credential `GPG-passphrase-for-terraform` matches your new key's passphrase

### 4.2 Update Terraform Registry

**To VIEW your current public key:**
1. Go to your provider page: https://registry.terraform.io/providers/biot-med/biot-gen2
2. Click on your provider name/logo or navigate to the provider's main page
3. Look for a **"GPG Keys"** or **"Signing Keys"** section (usually in the provider's settings or about page)
4. Alternatively, go to: https://registry.terraform.io/publish/provider/settings
   - Select your namespace: `biot-med`
   - Select your provider: `biot-gen2`
   - You should see a section showing your uploaded GPG public key(s)

**To UPDATE/REPLACE your public key:**
1. Go to https://registry.terraform.io/publish/provider/settings
2. Select your namespace: `biot-med`
3. Select your provider: `terraform-provider-biot-gen2` (or `biot-gen2`)
4. Find the **"GPG Keys"** or **"Signing Keys"** section
5. **Add your new public key** by uploading the `public.key` file from Step 2.3
6. Save the changes

**⚠️ Can't remove the old key?** That's okay! Here's why:

- **Multiple keys are allowed**: Terraform Registry can have multiple GPG keys registered. This is actually common when rotating keys.
- **Revocation protects you**: Since you've revoked the old key (Step 1.5), verification systems will check keyservers and reject signatures from the revoked key, even if it's still listed in the registry.
- **Future releases use new key**: As long as your signing scripts use the new key ID (which we'll update in Step 3), all future releases will be signed with the new key.
- **Old key becomes inactive**: The old key will remain in the registry but won't be used for new signatures, and its revoked status will prevent it from being trusted.

**If you really need to remove it:**
- Contact HashiCorp support at support@hashicorp.com and explain the security situation
- They may be able to manually remove the compromised key from their system
- However, this is **not strictly necessary** as long as the key is revoked on keyservers

### 4.3 (Optional) Add to GitHub Account
1. Go to GitHub → Settings → SSH and GPG keys
2. Add your new GPG public key
3. Remove the old compromised key

---

## Step 5: Verify Everything Works

### 5.1 Test signing locally
```bash
export GPG_PASSPHRASE='your-new-passphrase'
echo "test" > test.txt
./sign-checksum.sh test.txt
gpg --verify test.txt.sig test.txt
```

If verification succeeds, you're good to go!

### 5.2 Clean up
```bash
# Remove test files
rm test.txt test.txt.sig

# Remove revocation certificate (after confirming it's uploaded to keyservers)
# This file doesn't need to be in git - the revocation is now on keyservers
rm revoke-7C56BDFFED7D41BE.asc
```

**Important:** Do NOT commit the revocation certificate to git. It's public information, but:
- The revocation is already on keyservers (that's what matters)
- Keeping it in git doesn't provide any additional security benefit
- It's just clutter in your repository

---

## Step 6: Remove Key from Git History (Optional but Recommended)

Even though you've revoked the key, you should remove it from Git history to prevent anyone from finding it:

```bash
# Use git-filter-repo or BFG Repo-Cleaner to remove the key file from history
# This is a destructive operation - coordinate with your team first!

# Example using git-filter-repo:
# git filter-repo --path path/to/keyfile.asc --invert-paths

# Or use BFG:
# bfg --delete-files keyfile.asc
```

**⚠️ WARNING**: Rewriting Git history requires force-pushing and coordination with your team. Consider this carefully.

---

## Summary Checklist

- [ ] Revoked old key (`7C56BDFFED7D41BE`)
- [ ] Uploaded revocation certificate to keyservers
- [ ] Created new GPG key
- [ ] Exported new public key
- [ ] Updated `sign-checksum.sh` with new key ID
- [ ] Updated `old-sign-checksum.sh` with new key ID
- [ ] Updated Jenkins credentials with new private key
- [ ] Updated Terraform Registry with new public key
- [ ] Tested signing locally
- [ ] (Optional) Removed key from Git history

---

## Need Help?

If you encounter any issues, refer to:
- GPG documentation: https://www.gnupg.org/documentation/
- GitHub guide: https://docs.github.com/en/authentication/managing-commit-signature-verification

