package materialize

import (
	"fmt"
	"strings"
)

type ValueSecretStruct struct {
	Text   string
	Secret IdentifierSchemaStruct
}

func ReadConnectionId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(name), QuoteString(schema), QuoteString(database))
}

func ReadConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = %s;`, QuoteString(id))
}

func ReadConnectionDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = '%s'`, databaseName))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = '%s'`, schemaName))
		}
	}

	q.WriteString(`;`)
	return q.String()
}