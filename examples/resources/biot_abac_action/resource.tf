# Adds a filter to a search request. filterValue is a BioT search filter (FilterV2), and
# overrides only the operators it sets - other filters in the request are kept.
resource "biot_abac_action" "only_active_patients" {
  id          = "TF_ONLY_ACTIVE_PATIENTS"
  value       = "AddSearchFilterAction"
  description = "Restrict patient searches to active patients"

  params = jsonencode({
    filterName  = "_status"
    filterValue = { in = ["ACTIVE"] }
  })

  tags = ["team-access"]
}

# Removes attributes from a response. Runs at POST_REQUEST. Use full attribute paths:
# "person" removes the whole object, "person.name" just that field.
resource "biot_abac_action" "hide_ssn" {
  id    = "TF_HIDE_SSN"
  value = "RemoveAttributesInParamsAction"

  params = jsonencode({
    values = ["ssn"]
  })
}
