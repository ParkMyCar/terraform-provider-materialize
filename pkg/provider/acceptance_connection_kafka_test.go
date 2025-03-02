package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccConnKafka_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestMatchResourceAttr("materialize_connection_kafka.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "comment", "object comment"),
					testAccCheckConnKafkaExists("materialize_connection_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafkaMultipleBrokers_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaMultipleBrokerResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "3"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.broker", "redpanda:9092"),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafkaMultipleSsh_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaSshResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "3"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.ssh_tunnel.0.name", connectionName+"_ssh_conn_1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.ssh_tunnel.0.name", connectionName+"_ssh_conn_2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.broker", "redpanda:9092"),
					resource.TestCheckNoResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.ssh_tunnel.0.name"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.name", connectionName+"_ssh_conn_2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.schema_name", "public"),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnKafkaResource(roleName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnKafkaExists("materialize_connection_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnKafka_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "CONNECTION",
							Name:       connectionName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnKafkaResource(roleName, connectionName, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_connection_kafka" "test" {
	name = "%[2]s"
	kafka_broker {
		broker = "redpanda:9092"
	}
	security_protocol = "PLAINTEXT"
	comment = "object comment"
}

resource "materialize_connection_kafka" "test_role" {
	name = "%[3]s"
	kafka_broker {
		broker = "redpanda:9092"
	}
	security_protocol = "PLAINTEXT"
	ownership_role = "%[4]s"

	depends_on = [materialize_role.test]

	validate = false
}
`, roleName, connectionName, connection2Name, connectionOwner)
}

func testAccConnKafkaMultipleBrokerResource(connectionName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
		validate = false
	}
	`, connectionName)
}

func testAccConnKafkaSshResource(connectionName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_ssh_tunnel" "ssh_connection_1" {
		name = "%[1]s_ssh_conn_1"	  
		host = "ssh_host"
		user = "ssh_user"
		port = 22
	}

	resource "materialize_connection_ssh_tunnel" "ssh_connection_2" {
		name = "%[1]s_ssh_conn_2"	  
		host = "ssh_host"
		user = "ssh_user"
		port = 22
	}

	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
			ssh_tunnel {
				name = materialize_connection_ssh_tunnel.ssh_connection_1.name
			}
		}
		kafka_broker {
			broker = "redpanda:9092"
			ssh_tunnel {
				name = materialize_connection_ssh_tunnel.ssh_connection_2.name
			}
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		ssh_tunnel {
			name = materialize_connection_ssh_tunnel.ssh_connection_2.name
		}
		security_protocol = "PLAINTEXT"
		validate = false
	}
	`, connectionName)
}

func testAccCheckConnKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection kafka not found: %s", name)
		}
		_, err = materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnKafkaDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_kafka" {
			continue
		}

		_, err := materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("connection %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}
