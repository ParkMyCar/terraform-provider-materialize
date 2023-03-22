package resources

import (
	"context"
	"fmt"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var schemaSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the schema.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The name of the database.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the schema.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

func Schema() *schema.Resource {
	return &schema.Resource{
		Description: "The second highest level namespace hierarchy in Materialize.",

		CreateContext: schemaCreate,
		ReadContext:   schemaRead,
		DeleteContext: schemaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: schemaSchema,
	}
}

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadSchemaParams(i)

	var name, database_name string
	if err := conn.QueryRowx(q).Scan(&name, &database_name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database_name); err != nil {
		return diag.FromErr(err)
	}

	qn := fmt.Sprintf("%s.%s", database_name, name)
	if err := d.Set("qualified_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func schemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSchemaBuilder(schemaName, databaseName)
	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "schema"); err != nil {
		return diag.FromErr(err)
	}
	return schemaRead(ctx, d, meta)
}

func schemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewSchemaBuilder(schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "schema"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}