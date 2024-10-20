
# sqlconditions

sqlconditions serves the purpose of defining conditional SQL operations, and render them dynamically in function of a set of named arguments.

An SQL operation may have from that one configuration. This is useful for situations where the SQL operation must vary depending on a role. See *variants*.

## Status : WIP

This library is its early days, things may break.


## Usage overview

1 Load your SQL operations configuration from a YAML configuration file or from code :

```
---
version: 1

operations:
  get-products:
    variants:
      default:
        condition:
          - (
          - if_present order_id = @orderID
          - OR if_present client_id = @clientID
          - )
      admin:
        condition:
          - (
          - if_present order_id = @orderID
          - AND if_present client_id = @clientID
          - )
```

2 At runtime

Fetch a SQL operation config 

```
customConfig := sqlcond.GetConfig("get-products", "admin")
```

and given a set of named arguments :

```
namedArgs := map[string]any{
	"orderID" : "2024-02-02-1234"
}
```

render it :

```
sqlCond := customConfig.ToSQL(namedArgs)
```

to get the rendered SQL condition :

```
(order_id = @orderID)
```

Here the OR condition on `client_id = @clientID` was skipped omitted because `clientID` is missing from `namedArgs`.

