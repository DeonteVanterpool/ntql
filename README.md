# SSQL

SSQL (Super Simple Query Language) is a domain-specific language designed to convert human-readable queries into SQL `WHERE` clauses for filtering tasks and related data. The schema is fully configurable via `schema.yaml` and supports automatic cross-table join resolution.

The project consists of:
- A **compiler / transpiler** that:
  - Tokenises the input into a stream of typed tokens
  - Validates the structure of the query
  - Builds an Abstract Syntax Tree (AST)
  - Evaluates the AST into SQL, including automatic multi-table join generation
- An **autocompletion engine** that predicts the next valid token at any point in the query

## Demo

The demo video can be viewed [here](https://drive.google.com/file/d/1TJY_Jc4UP607mJEEj3nrePgT30Ji70ws/view?usp=sharing). You can also install a web server with SSQL and test the language [here](https://github.com/DeonteVanterpool/ntql-demo).

---

## Syntax

### Basic query structure

Multiple queries are connected using `AND` or `OR`. Queries can be negated with `!` and grouped with parentheses to control evaluation order.

```
query1 OR query2
query1 AND query2
query1 AND !(query2)
query1 AND !(query2 OR query3)
```

### Individual query form

Each individual query follows the form `subject.verb(<expression>)`.

```
title.contains(<expression>)
status.equals(<expression>)
due.before(<expression>)
tag.equals(<expression>)
```

### Expressions

The expression inside the parentheses can itself be compound, using `AND`, `OR`, `!`, and parentheses. A compound expression inside a verb call distributes over the subject:

```
description.contains(roadmap OR bug)
  ≡  description.contains(roadmap) OR description.contains(bug)

description.contains((roadmap OR bug) AND v2)
  ≡  (description.contains(roadmap) OR description.contains(bug)) AND description.contains(v2)

description.contains((roadmap OR bug) AND !(draft))
  ≡  (description.contains(roadmap) OR description.contains(bug)) AND !(description.contains(draft))
```

Expression values are typed. The valid type for each subject is defined in `schema.yaml`.

| Value type | Example            |
|------------|--------------------|
| `string`   | `"hello world"`, `hello` |
| `int`      | `42`               |
| `date`     | `2026-01-31`       |
| `dateTime` | `2026-01-31T14:00` |
| `tag`      | resolved via subquery |

---

## Grammar (Backus-Naur Form)

```
query      = expr

expr       = or_expr
or_expr    = and_expr ("OR" and_expr)*
and_expr   = not_expr ("AND" not_expr)*
not_expr   = ["!"] term
term       = func_call | "(" expr ")"

func_call  = subject "." verb "(" value_expr ")"
             # subject must be in the list of known subjects
             # verb must be valid for that subject

value_expr = value_or
value_or   = value_and ("OR" value_and)*
value_and  = value_not ("AND" value_not)*
value_not  = ["!"] value_term
value_term = "(" value_expr ")" | value

value      = object            # type determined by the current verb
object     = NUMBER | STRING | DATE | DATETIME | TAG
```

---

## Schema

SSQL reads its schema from `schema.yaml` (embedded at compile time). The schema controls which subjects, verbs, and value types are valid, and how subjects map to database tables and columns.

### Subjects

A **subject** is the left-hand side of a query (`title`, `due`, `status`, …). Each subject defines:

- `name` — canonical name used in queries
- `aliases` — alternative names that resolve to the same subject
- `validVerbs` — list of verbs (with optional aliases) that may be used with this subject
- `validTypes` — list of value types accepted by this subject's verbs
- `table` — the database table this subject maps to
- `column` — the database column this subject maps to

### Field types

`fieldTypes` lists column names by their SQL type. This drives type-safe comparisons in generated SQL.

| Category       | Examples                                |
|----------------|-----------------------------------------|
| `dateTypes`    | `due_date`, `created_at`, `updated_at`  |
| `boolTypes`    | `completed`, `all_day`                  |
| `numericTypes` | `priority`                              |
| `stringTypes`  | `title`, `description`                  |

### Tables

Each entry in `tables` declares a database table and its primary key:

```yaml
tables:
  - name: tasks
    primaryKey: id
  - name: projects
    primaryKey: id
  - name: tags
    primaryKey: id
  - name: task_tags
    primaryKey: task_id
```

The tables defined in the default schema are:

| Table       | Columns                                                                                   |
|-------------|-------------------------------------------------------------------------------------------|
| `tasks`     | id, title, description, status, priority, due_date, created_at, updated_at, completed_at, created_by, project_id |
| `projects`  | id, name, description, created_at                                                         |
| `tags`      | id, name, color                                                                           |
| `task_tags` | task_id, tag_id                                                                           |

### Joins

`joins` defines the relationships between tables used for automatic join resolution. Each entry specifies a directed link from one table to another:

```yaml
joins:
  - fromTable: tasks
    toTable: projects
    fromKey: project_id
    toKey: id
  - fromTable: tasks
    toTable: task_tags
    fromKey: id
    toKey: task_id
  - fromTable: task_tags
    toTable: tags
    fromKey: tag_id
    toKey: id
```

### Subject mappings

When `tables` are defined in the schema, every subject **must** declare a `table` and `column` mapping. SSQL uses these mappings to determine which table a condition belongs to and to generate the correct qualified column reference in the output SQL.

```yaml
subjects:
  - name: title
    table: tasks
    column: title
    ...
  - name: project
    table: projects
    column: name
    ...
  - name: tag
    table: tags
    column: name
    ...
```

---

## Multi-table query examples

When a query references subjects from more than one table, SSQL automatically resolves the required joins and generates a complete SQL statement with table aliases.

### Single table (no join needed)

```
title.contains("roadmap")
```

```sql
SELECT t0.* FROM tasks t0
WHERE t0.title LIKE '%roadmap%'
```

### Two tables

```
title.contains("roadmap") AND project.equals("Platform")
```

```sql
SELECT t0.* FROM tasks t0
LEFT JOIN projects t1 ON t0.project_id = t1.id
WHERE (t0.title LIKE '%roadmap%' AND t1.name = 'Platform')
```

### Three tables (with junction table)

```
project.equals("Platform") AND tag.equals("backend")
```

```sql
SELECT t0.* FROM tasks t0
LEFT JOIN projects t1 ON t0.project_id = t1.id
LEFT JOIN task_tags t2 ON t0.id = t2.task_id
LEFT JOIN tags t3 ON t2.tag_id = t3.id
WHERE (t1.name = 'Platform' AND t3.name = 'backend')
```

Each join is emitted at most once even when multiple conditions reference the same table.

---

## Features

### Deterministic autocompletion

The grammar is designed so that the set of valid next tokens is always computable from the current parser state. The autocompletion engine uses this property to offer precise, context-aware suggestions at every position in the query.

### Flexible subject naming

Subjects support aliases, allowing natural alternatives (`name` for `title`, `deadline` for `due`, `state` for `status`). Verb aliases work the same way (`eq` for `equals`, `lt` for `lessthan`, etc.).

### Cross-table relationships

SSQL resolves multi-table queries automatically. You write predicates using logical subject names; the engine figures out which tables to join and produces a single, correct SQL statement.

### Type safety

Each subject declares the value types it accepts, and the compiler enforces this at parse time. Date comparisons, numeric comparisons, and string pattern matching all produce the correct SQL operators.

---

## Design principles

- **Simple surface syntax** — the query language is intentionally minimal so non-technical users can write queries without knowing SQL.
- **Configurable schema** — all subjects, verbs, tables, and joins are declared in `schema.yaml`. No code changes are required to add a new filterable field.
- **Safe SQL output** — inputs are validated and escaped before being embedded in SQL to prevent injection.
- **Composable AST** — the AST can be built programmatically as well as parsed from a string, making SSQL useful as an embedded query-builder library.

---

## Example queries

```
# Tasks with "meeting" in the title
title.contains("meeting")

# High-priority tasks due before end of year
priority.gte(3) AND due.before(2026-12-31)

# Tasks that are in-progress or under review
status.equals(in-progress) OR status.equals(review)

# Tasks in the "Platform" project that are not completed
project.equals("Platform") AND !(status.equals(completed))

# Tasks tagged "backend" created this year
tag.equals(backend) AND createdAt.after(2026-01-01)

# Tasks created by alice, ordered by priority, due before a deadline
createdBy.equals(alice) AND due.before(2026-06-01) AND priority.gte(2)

# Complex nested query
(title.contains(bug) OR tag.equals(urgent)) AND !(status.equals(closed)) AND project.equals("Core")
```

