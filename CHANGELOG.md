## 1.1.0

**Release date**: Oct 06, 2026
- [SOFT-9880] Added ABAC (access-control) resources, so conditions, actions and rules can be managed in Terraform and promoted between environments
    - `biot_abac_condition`, `biot_abac_action` and `biot_abac_rule`, importable by id
    - The `<<BuiltIn>>` tag is managed by BioT and exposed as a read-only `built_in` attribute instead of appearing in `tags`
    - Changing the `value` (implementation) of an existing condition or action is refused at plan time; give it a new `id` to switch implementation

## 1.0.11

**Release date**: Sep 14, 2026
- [SOFT-9875] Fixed `validation.unique` being planned as `false` even when set to `true` in the configuration, caused by the schema default overwriting configured values inside the attribute sets

## 1.0.10

**Release date**: Jun 21, 2026
- [SOFT-9713] Added `ui_configuration` to template attributes for configuring date/date-time display style

## 1.0.9

**Release date**: Jun 04, 2026
- [SOFT-9662] added branding category to organization template
    - Support supported mime types in validation
    - Support `public_access` - a read-only field on built-in attributes only.

## 1.0.8

**Release date**: Jul 22, 2026
- [SOFT-9778] Fixed Provider produced inconsistent result after apply error caused by the `validation.unique` attribute not being marked as `Computed`

## 1.0.7

**Release date**: Jun 03, 2026
- [SOFT-9695] Vulnerability fixes

## 1.0.4

**Release date**: JAN 27, 2026
- Testing jenkins

## 1.0.3

**Release date**: JAN 27, 2026

- fixed lowerange, uperrange, min, max bug - change them from int to number so we ca
n accept both int and float

## 1.0.2

**Release date**: NOV 26, 2025

## Changes

- Fixed default_value validation to support object and not just strings

## 1.0.1

**Release date**: Nov 23, 2025

## Changes

- Added token caching to reduce API calls by reusing access tokens across multiple terraform operations

## 1.0.0

**Release date**: Sep 30, 2025 

## Changes

- First Version