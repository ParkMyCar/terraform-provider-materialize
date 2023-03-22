package resources

import (
	"context"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterSchema = map[string]*schema.Schema{
	"name": {
		Description: "A name for the cluster.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "A logical cluster, which contains dataflow-powered objects.",

		CreateContext: clusterCreate,
		ReadContext:   clusterRead,
		DeleteContext: clusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: clusterSchema,
	}
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadClusterParams(i)

	var name string
	if err := conn.QueryRowx(q).Scan(&name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	clusterName := d.Get("name").(string)

	builder := materialize.NewClusterBuilder(clusterName)
	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "cluster"); err != nil {
		return diag.FromErr(err)
	}
	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	clusterName := d.Get("name").(string)

	q := materialize.NewClusterBuilder(clusterName).Drop()

	if err := dropResource(conn, d, q, "cluster"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}