package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourcePostgresSchema = map[string]*schema.Schema{
	"name":                ObjectNameSchema("source", true, false),
	"schema_name":         SchemaNameSchema("source", false),
	"database_name":       DatabaseNameSchema("source", false),
	"qualified_sql_name":  QualifiedNameSchema("source"),
	"comment":             CommentSchema(false),
	"cluster_name":        ObjectClusterNameSchema("source"),
	"size":                ObjectSizeSchema("source"),
	"postgres_connection": IdentifierSchema("postgres_connection", "The PostgreSQL connection to use in the source.", true),
	"publication": {
		Description: "The PostgreSQL publication (the replication data set containing the tables to be streamed to Materialize).",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"text_columns": {
		Description: "Decode data as text for specific columns that contain PostgreSQL types that are unsupported in Materialize. Can only be updated in place when also updating a corresponding `table` attribute.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"table": {
		Description: "Creates subsources for specific tables. If neither table or schema is specified, will default to ALL TABLES",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the table.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"alias": {
					Description: "The alias of the table.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Optional:      true,
		MinItems:      1,
		ConflictsWith: []string{"schema"},
	},
	"schema": {
		Description:   "Creates subsources for specific schemas. If neither table or schema is specified, will default to ALL TABLES",
		Type:          schema.TypeList,
		Elem:          &schema.Schema{Type: schema.TypeString},
		Optional:      true,
		ForceNew:      true,
		MinItems:      1,
		ConflictsWith: []string{"table"},
	},
	"expose_progress": IdentifierSchema("expose_progress", "The name of the progress subsource for the source. If this is not specified, the subsource will be named `<src_name>_progress`.", false),
	"subsource":       SubsourceSchema(),
	"ownership_role":  OwnershipRoleSchema(),
	"region":          RegionSchema(),
}

func SourcePostgres() *schema.Resource {
	return &schema.Resource{
		Description: "A Postgres source describes a PostgreSQL instance you want Materialize to read data from.",

		CreateContext: sourcePostgresCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourcePostgresUpdate,
		DeleteContext: sourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourcePostgresSchema,
	}
}

func sourcePostgresCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourcePostgresBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
	}

	if v, ok := d.GetOk("postgres_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.PostgresConnection(conn)
	}

	if v, ok := d.GetOk("publication"); ok {
		b.Publication(v.(string))
	}

	if v, ok := d.GetOk("table"); ok {
		tables := materialize.GetTableStruct(v.([]interface{}))
		b.Table(tables)
	}

	if v, ok := d.GetOk("schema"); ok {
		schemas := materialize.GetSliceValueString(v.([]interface{}))
		b.Schema(schemas)
	}

	if v, ok := d.GetOk("expose_progress"); ok {
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
	}

	if v, ok := d.GetOk("text_columns"); ok {
		columns := materialize.GetSliceValueString(v.([]interface{}))
		b.TextColumns(columns)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceRead(ctx, d, meta)
}

func sourcePostgresUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSource(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		_, newSize := d.GetChange("size")
		if err := b.Resize(newSize.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("table") {
		ot, nt := d.GetChange("table")
		addTables := materialize.DiffTableStructs(nt.([]interface{}), ot.([]interface{}))
		dropTables := materialize.DiffTableStructs(ot.([]interface{}), nt.([]interface{}))

		if len(addTables) > 0 {
			var colDiff []string
			if d.HasChange("text_columns") {
				oc, nc := d.GetChange("text_columns")
				colDiff = diffTextColumns(nc.([]interface{}), oc.([]interface{}))
			}

			if err := b.AddSubsource(addTables, colDiff); err != nil {
				return diag.FromErr(err)
			}
		}
		if len(dropTables) > 0 {
			if err := b.DropSubsource(dropTables); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return sourceRead(ctx, d, meta)
}

func diffTextColumns(arr1, arr2 []interface{}) []string {
	arr2Map := make(map[string]bool)
	for _, item := range arr2 {
		i := item.(string)
		arr2Map[i] = true
	}

	var difference []string
	for _, item := range arr1 {
		i := item.(string)
		if !arr2Map[i] {
			difference = append(difference, i)
		}
	}
	return difference
}
