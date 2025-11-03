# NTQL
NTQL (NolaTask Query Language) is a domain specific language designed to filter todo list items by certain criteria.

The project consists of 
- a compiler / transpiler
    - breaks down the syntax into tokens
    - detects validity of the form of a query
    - creates a syntax tree
    - evaluates the syntax tree into SQL
- an autocompletion engine.
    - predicts the next possible token

## Syntax
Multiple queries can be connected using connectors such as `AND` or `OR`. Queries can also be negated using an exclamation mark `!`. You can even influence the order in which each query is evaluated by using parentheses

Examples:
- `query1 OR query2`
- `query1 AND query2`
- `query1 AND !(query2)`
- `query1 AND !(query2 OR query3)`

Each individual query is in the following form: `subject.verb(<expression>)`.

Examples:
- `description.contains(<expression>)`
- `title.startswith(<expression>)`
- `tag.equals(<expression>)`

Depending on the subject, the expression can have different data types.

Examples: 
- `description`: `string`,
- `tag`: `tag` (evaluates to the numeric id of the entered tag after parsing),
- `completedAt`: `datetime` or `date`,

Similarly to queries, expressions can also be compounded using connectors. Precedence can be defined using parentheses. Expressions can be negated.

Examples:
- `description.contains(<expression1> OR <expression2>)` is equal to `description.contains(<expression1>) OR description.contains(<expression2>)`
- `description.contains((<expression1> OR <expression2>) AND <expression3>)` is equal to `(description.contains(<expression1>) OR description.contains(<expression2>)) AND description.contains(<expression3>)`
- `description.contains((<expression1> OR <expression2>) AND !(<expression3>))` is equal to `(description.contains(<expression1>) OR description.contains(<expression2>)) AND !(description.contains(<expression3>))`

The combined query gets parsed into an Abstract Syntax Tree (AST), which is compiled to `SQL` queries which can be evaluated in the database. These queries can span different database tables and the input needs to be sanitized to prevent `SQL` injections.

The grammar can be more technically defined in Backus Naur Form:
```
query = expr
expr = or_expr
or_expr = and_expr ("OR" and_expr)*
and_expr = not_expr ("AND" not_expr)*
not_expr = ["!"] term
term = func_call | "(" expr ")"
func_call = subject "." verb "(" value_expr ")" # subject from list of subjects, verb from subject verbs
value_expr = value_or
value_or = value_and ("OR" value_and)*
value_and = value_not ("AND" value_not)*
value_not = ["!"] value_term
value_term = "(" value_expr ")" | value
value = object # type belonging to current verb
object = NUMBER | STRING | DATE | TAG
```

## Design
The design of the language was created to make autocompletion highly deterministic. It does so by limiting the number of possibilities for the next token class. There can only be a limited number of data types for each expected token.

