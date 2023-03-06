---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_source_postgres Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.
---

# materialize_source_postgres (Resource)

A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.

## Example Usage

```terraform
resource "materialize_source" "example_source_postgres" {
  name                = "source_postgres"
  schema_name         = "schema"
  size                = "3xsmall"
  postgres_connection = "pg_connection"
  publication         = "mz_source"
  tables = {
    "schema1.table_1" = "s1_table_1"
    "schema2_table_1" = "s2_table_1"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1)
#   WITH (SIZE = '3xsmall');
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The identifier for the source.
- `postgres_connection` (String) The name of the PostgreSQL connection to use in the source.
- `publication` (String) The PostgreSQL publication (the replication data set containing the tables to be streamed to Materialize).

### Optional

- `cluster_name` (String) The cluster to maintain this source. If not specified, the size option must be specified.
- `database_name` (String) The identifier for the source database.
- `schema_name` (String) The identifier for the source schema.
- `size` (String) The size of the source.
- `tables` (Map of String) Creates subsources for specific tables in the load generator.
- `text_columns` (List of String) Decode data as text for specific columns that contain PostgreSQL types that are unsupported in Materialize.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_name` (String) The fully qualified name of the source.

## Import

Import is supported using the following syntax:

```shell
# Sources can be imported using the source id:
terraform import materialize_source_postgres.example_source_postgres <source_id>
```