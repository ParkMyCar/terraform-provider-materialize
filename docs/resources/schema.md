---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_schema Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  The second highest level namespace hierarchy in Materialize.
---

# materialize_schema (Resource)

The second highest level namespace hierarchy in Materialize.

## Example Usage

```terraform
resource "materialize_schema" "example_schema" {
  name          = "schema"
  database_name = "database"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The identifier for the schema.

### Optional

- `comment` (String) **Private Preview** Comment on an object in the database.
- `database_name` (String) The identifier for the schema database. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `ownership_role` (String) The owernship role of the object.
- `region` (String) The region to use for the resource connection. If not set, the default region is used.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_sql_name` (String) The fully qualified name of the schema.

## Import

Import is supported using the following syntax:

```shell
# Schemas can be imported using the schema id:
terraform import materialize_schema.example_schema <region>:<schema_id>

# Schema id and information be found in the `mz_catalog.mz_schemas` table
# The role is the role where the database is located (e.g. aws/us-east-1)
```
