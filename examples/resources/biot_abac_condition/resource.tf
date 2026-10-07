# A DYNAMIC condition: true when the initiator's _ownerOrganization.id matches the
# entity being accessed.
resource "biot_abac_condition" "initiator_owner_organization" {
  id          = "TF_INITIATOR_OWNER_ORGANIZATION"
  value       = "InitiatorAttributeInParamsCondition"
  description = "Initiator belongs to the owning organization"

  params = jsonencode({
    initiatorAttributeName = "_ownerOrganization.id"
    values                 = ["$ENTITY_OWNER_ORGANIZATION_ID"]
  })

  # Tags are yours to manage. Set to [] to clear them; omit to leave whatever BioT has.
  tags = ["team-access"]
}

# A COMPOSITE condition inverts or combines other conditions. Referencing the resource
# attribute (rather than hard-coding the id string) is what makes Terraform create the
# referenced condition first.
resource "biot_abac_condition" "not_initiator_owner_organization" {
  id    = "TF_NOT_INITIATOR_OWNER_ORGANIZATION"
  value = "CompositeCondition"

  params = jsonencode({
    operator = "NOT"
    values   = [biot_abac_condition.initiator_owner_organization.id]
  })
}

# built_in is read-only and reports whether BioT ships the condition. Built-in conditions
# can be imported and updated, but not destroyed. BioT manages the "<<BuiltIn>>" tag
# itself, so it never shows up in `tags`.
output "owner_organization_is_built_in" {
  value = biot_abac_condition.initiator_owner_organization.built_in
}
