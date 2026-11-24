# Normally generated with `terraform query` from a view designed in the BioT portal builder.
# Templates are referred to by name, so the same configuration works in every environment.
resource "biot_portal_view" "patient_expand" {
  portal_type   = "ORGANIZATION_PORTAL"
  view_type     = "TEMPLATE_EXPAND"
  template_name = "patient"

  # BioT creates the view with the template, so the template must be applied first.
  depends_on = [biot_template.patient]

  config_json = jsonencode({
    attributes = [
      {
        classification = "REGULAR"
        selected = [
          { name = "_name", basePath = null },
          { name = "_email", basePath = null },
        ]
      },
    ]
    measurements = {
      views       = ["LIVE", "HISTORY"]
      defaultView = "LIVE"
      chart       = []
      headline    = []
    }
    tabs = [
      {
        type = "EXTERNAL_PAGE"
        externalPage = {
          url   = "https://example.com/report"
          title = "Report"
        }
      },
    ]
  })
}
