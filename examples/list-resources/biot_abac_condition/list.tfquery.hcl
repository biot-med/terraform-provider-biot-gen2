# Requires Terraform 1.14 or later. Run with:
#
#   terraform query                                       # list matching conditions
#   terraform query -generate-config-out=conditions.tf    # generate import + resource blocks
#
# The generated blocks say `provider = biot-gen2`, so declare the provider under that local
# name in required_providers for them to work without edits.

# Your own conditions only, leaving out the ones BioT ships.
list "biot_abac_condition" "custom" {
  provider = biot-gen2

  # Terraform returns at most 100 results per list block unless told otherwise.
  limit = 1000

  config {
    built_in = false
  }
}
