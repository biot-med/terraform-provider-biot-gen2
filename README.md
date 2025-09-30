# Terraform Provider Development Guide

This guide will help you build and install a local Terraform provider for development.

## Prerequisites

- Go installed on your system (1.23 +)
- Terraform installed on your system (v1.12 +)

## Initial Setup (One-time only)

### Step 1: Initialize Go modules

First, make sure all dependencies are downloaded:

```bash
go mod tidy
```

### Step 2: Build the provider

Build the provider binary:

```bash
go build -o terraform-provider-biot-gen2
```

This creates a file called `terraform-provider-biot-gen2` (on macOS/Linux) or `terraform-provider-biot-gen2.exe` (on Windows).

### Step 3: Install the provider locally

**Important**: The version number in this example is `1.0.0`. Change this if you're using a different version.

#### For macOS users

**Step 3a: Check your system setup**

Run these commands to see what type of Mac you have and how Terraform is installed:

```bash
# Check your Mac's processor type
uname -m

# Check how Terraform was installed
file $(which terraform)
```

**What the results mean:**
- If `uname -m` shows `arm64` and `file` shows `x86_64`: You have an Apple Silicon Mac, but Terraform is running under Rosetta (use `darwin_amd64`)
- If `uname -m` shows `arm64` and `file` shows `arm64`: You have an Apple Silicon Mac with native Terraform (use `darwin_arm64`)
- If `uname -m` shows `x86_64`: You have an Intel Mac (use `darwin_amd64`)

**Step 3b: Install the provider**

**Option A: Install for your specific setup (recommended for individual developers)**

Choose the command that matches your setup from Step 3a:

```bash
# For Apple Silicon Macs with native Terraform (ARM64)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_arm64
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_arm64/

# OR for Apple Silicon Macs with Terraform under Rosetta (AMD64)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64/

# OR for Intel Macs (AMD64)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64/
```

**Option B: Install for all setups (recommended for team development)**

This installs the provider for both Apple Silicon and Intel Macs:

```bash
# Create directories for both architectures
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_arm64
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64

# Copy the provider to both locations
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_arm64/
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/darwin_amd64/
```

#### For Windows users

```cmd
mkdir "%APPDATA%\terraform.d\plugins\example.com\biot\biot\1.0.0\windows_amd64"
copy terraform-provider-biot-gen2.exe "%APPDATA%\terraform.d\plugins\example.com\biot\biot\1.0.0\windows_amd64\"
```

### Step 4: Test the installation

Create a simple Terraform configuration to test your provider:

```hcl
# test.tf
terraform {
  required_providers {
    biot = {
      source = "registry.terraform.io/biot-med/biot"
      version = "1.0.0"
    }
  }
}

# Add your provider resources here
```

Then run:

```bash
terraform init
terraform plan
```

## Making changes and rebuilding

When you make changes to the provider code, follow the **Daily Development Workflow** above:

1. **Rebuild the provider:**
   ```bash
   go build -o terraform-provider-biot-gen2
   ```

2. **Reinstall the provider** using the copy command from the workflow section

3. **Test your changes:**
   ```bash
   terraform plan
   ```

## Troubleshooting

### Common errors and solutions

**Error: "Provider does not have a package available for your current platform"**

This usually means you installed the provider for the wrong architecture. 

**For macOS users:**
1. Check your setup again using the commands in Step 3a
2. Make sure you used the correct installation command
3. If unsure, use Option B (install for all setups) from Step 3b

**For Windows users:**
- Make sure you're using the `windows_amd64` directory

**Error: "Provider not found"**

1. Check that you copied the provider to the correct directory
2. Verify the version number matches your Terraform configuration
3. Make sure the provider binary has the correct name and permissions

### Quick diagnostic commands

**For macOS:**
```bash
# Check if provider is installed
ls ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/

# Check provider binary permissions
ls -la ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.0/*/terraform-provider-biot-gen2
```

**For Windows:**
```cmd
dir "%APPDATA%\terraform.d\plugins\example.com\biot\biot\1.0.0\"
```

### Still having issues?

1. Make sure you're using the correct version number throughout
2. Try deleting the provider directories and reinstalling
3. Check that your Terraform configuration references the correct provider source and version
