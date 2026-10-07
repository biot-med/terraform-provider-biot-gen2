# Requires Terraform 1.14 or later. Run with:
#
#   terraform query                                  # list matching rules
#   terraform query -generate-config-out=rules.tf    # generate import + resource blocks
#
# The generated blocks say `provider = biot-gen2`, so declare the provider under that local
# name in required_providers for them to work without edits.

# Your own rules only, leaving out the ones BioT ships.
list "biot_abac_rule" "custom" {
  provider = biot-gen2

  # Terraform returns at most 100 results per list block unless told otherwise.
  limit = 1000

  config {
    built_in = false
  }
}

# Every rule - built-in or not - that runs on one API. Filters combine with AND, and
# action_id / condition_id find the rules that reference a given action or condition.
list "biot_abac_rule" "on_patient_search" {
  provider = biot-gen2

  config {
    api_id = "GET/organization/v1/users/patients"
  }
}
