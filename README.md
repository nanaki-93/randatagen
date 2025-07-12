
# Randatagen

in Generate command:
- add sql and no-sql support
- add mode databases
- add support for graphql
- performance on huge insert
- set of preset values for string types ex (days of week, adjectives, names)
- add a gui to select the data to generate



in Migrate command:
- add support for avoiding tables(list in migrate.json)
- check the table is empty before migrating
- add support to create tables and all the structures if not exists
- check performance on huge insert



//Postgres create table query
SELECT 'CREATE TABLE ' || relname || E'\n(\n' ||
array_to_string(
array_agg(
'    ' || column_name || ' ' ||  type ||
CASE WHEN not is_nullable THEN ' NOT NULL' ELSE '' END
)
, E',\n'
) || E'\n);'
FROM (
SELECT
c.relname,
a.attname AS column_name,
pg_catalog.format_type(a.atttypid, a.atttypmod) AS type,
a.attnotnull AS is_nullable
FROM pg_class c
JOIN pg_attribute a ON a.attrelid = c.oid
WHERE c.relname = 'nome_tabella'
AND a.attnum > 0
AND NOT a.attisdropped
) AS tabledef
GROUP BY relname;

//Postgres DDL query
1. Indici (INDEX):
SELECT indexdef
FROM pg_indexes
WHERE tablename = 'nome_tabella';
2. Primary Key:
SELECT
'ALTER TABLE ' || tc.table_name || ' ADD CONSTRAINT ' || tc.constraint_name ||
' PRIMARY KEY (' || string_agg(kcu.column_name, ', ') || ');' AS ddl
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_name = 'nome_tabella'
AND tc.constraint_type = 'PRIMARY KEY'
GROUP BY tc.table_name, tc.constraint_name;
3. Foreign Key:
   SELECT
   'ALTER TABLE ' || tc.table_name || ' ADD CONSTRAINT ' || tc.constraint_name ||
   ' FOREIGN KEY (' || string_agg(kcu.column_name, ', ') || ') REFERENCES ' ||
   ccu.table_name || ' (' || string_agg(ccu.column_name, ', ') || ');' AS ddl
   FROM information_schema.table_constraints tc
   JOIN information_schema.key_column_usage kcu
   ON tc.constraint_name = kcu.constraint_name
   JOIN information_schema.constraint_column_usage ccu
   ON tc.constraint_name = ccu.constraint_name
   WHERE tc.table_name = 'nome_tabella'
   AND tc.constraint_type = 'FOREIGN KEY'
   GROUP BY tc.table_name, tc.constraint_name, ccu.table_name;