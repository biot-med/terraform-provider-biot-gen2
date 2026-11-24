# Requires Terraform 1.14 or later. Run with:
#
#   terraform query                                  # list matching views
#   terraform query -generate-config-out=views.tf    # generate import + resource blocks
#
# The generated blocks say `provider = biot-gen2`, so declare the provider under that local
# name in required_providers for them to work without edits. Templates appear by name, so the
# generated blocks work in other environments too - add depends_on on the biot_template
# resources so the templates are applied first.

# The expand views of the organization portal.
list "biot_portal_view" "organization_expand" {
  provider = biot-gen2

  config {
    portal_type = "ORGANIZATION_PORTAL"
    view_type   = "TEMPLATE_EXPAND"
  }
}
