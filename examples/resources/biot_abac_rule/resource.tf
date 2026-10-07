# Clinicians searching patients only see active ones.

resource "biot_abac_condition" "is_clinician" {
  id    = "TF_IS_CLINICIAN"
  value = "InitiatorAttributeInParamsCondition"

  params = jsonencode({
    initiatorAttributeName = "_template.name"
    values                 = ["Clinician"]
  })
}

resource "biot_abac_action" "only_active_patients" {
  id    = "TF_ONLY_ACTIVE_PATIENTS"
  value = "AddSearchFilterAction"

  params = jsonencode({
    filterName  = "_status"
    filterValue = { in = ["ACTIVE"] }
  })
}

resource "biot_abac_rule" "clinicians_see_active_patients" {
  id          = "TF_CLINICIANS_SEE_ACTIVE_PATIENTS"
  description = "Restrict clinician patient searches to active patients"

  # Refer to the resources rather than typing ids, so Terraform creates them first.
  action_ids = [biot_abac_action.only_active_patients.id]

  conditions = [
    { id = biot_abac_condition.is_clinician.id, inverted = false },
  ]

  # api_id is the HTTP method followed directly by the path template - no space.
  api_execution_points = [
    {
      api_id              = "GET/organization/v1/users/patients"
      api_execution_point = "PRE_REQUEST"
      order               = 1
      enabled             = true
    },
  ]

  tags = ["team-access"]
}
