package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description:  "The identifier for the index.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"default": {
		Description:  "Creates a default index using all inferred columns are used.",
		Type:         schema.TypeBool,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"schema_name": {
		Description: "The identifier for the index schema.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"database_name": {
		Description: "The identifier for the index database.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"qualified_sql_name": SchemaResourceQualifiedName("view"),
	"obj_name":           IdentifierSchema("obj_name", "The name of the source, view, or materialized view on which you want to create an index.", true),
	"cluster_name": {
		Description: "The cluster to maintain this index. If not specified, defaults to the active cluster.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"method": {
		Description:  "The name of the index method to use.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Default:      "ARRANGEMENT",
		ValidateFunc: validation.StringInSlice([]string{"ARRANGEMENT"}, true),
	},
	"col_expr": {
		Description: "The expressions to use as the key for the index.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Description: "The name of the option you want to set.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"val": {
					Description: "The value for the option.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default", "col_expr"},
	},
}

func Index() *schema.Resource {
	return &schema.Resource{
		Description: "An in-memory index on a source, view, or materialized view.",

		CreateContext: indexCreate,
		ReadContext:   indexRead,
		DeleteContext: indexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: indexSchema,
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadIndexParams(i)

	var name, obj, schema, database string
	if err := conn.QueryRowx(q).Scan(&name, &obj, &schema, &database); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(database, schema, name)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func indexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	o := d.Get("obj_name").([]interface{})[0].(map[string]interface{})
	builder := materialize.NewIndexBuilder(
		d.Get("name").(string),
		d.Get("default").(bool),
		materialize.IdentifierSchemaStruct{
			Name:         o["name"].(string),
			SchemaName:   o["schema_name"].(string),
			DatabaseName: o["database_name"].(string),
		},
	)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("method"); ok {
		builder.Method(v.(string))
	}

	if v, ok := d.GetOk("col_expr"); ok {
		var colExprs []materialize.IndexColumn
		for _, colExpr := range v.([]interface{}) {
			b := colExpr.(map[string]interface{})
			colExprs = append(colExprs, materialize.IndexColumn{
				Field: b["field"].(string),
				Val:   b["val"].(string),
			})
		}
		builder.ColExpr(colExprs)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "index"); err != nil {
		return diag.FromErr(err)
	}
	return indexRead(ctx, d, meta)
}

func indexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	o := d.Get("obj_name").([]interface{})[0].(map[string]interface{})
	builder := materialize.NewIndexBuilder(
		d.Get("name").(string),
		d.Get("default").(bool),
		materialize.IdentifierSchemaStruct{
			Name:         o["name"].(string),
			SchemaName:   o["schema_name"].(string),
			DatabaseName: o["database_name"].(string),
		},
	)

	q := builder.Drop()

	if err := dropResource(conn, d, q, "index"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}