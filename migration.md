Migration Guides
================

## Migrating to V2

### Consume

Calls to `Consume(root interface{}) (*Container, error)` should be replaced with `Wrap(root interface{}) *Container`.

The error response was removed in order to avoid unnecessary duplicate type checks on `root`. This also allows shorthand queries of general objects like `gabs.Wrap(foo).S("bar","baz").Data()`.

### Search Across Arrays

All 	query functions (`Search`, `Path`, `Set`, `SetP`, etc) now attempt to resolve a specific index when they encounter an array. This means path queries must specify an integer index at the level of arrays within the content.

For example, given the sample document:

``` json
{
	"foo": [
		{
			"bar": {
				"baz": 45
			}
		}
	]
}
```

If we wished to access the value of the nested field `baz` in V1 the closest we could get from a single call would be `Search("foo", "bar", "baz")`, which would propagate the array in the result giving us `[45]`.

In V2 we can access the field directly with `Search("foo", "0", "bar", "baz")`. The important caveat is that the index is _required_, otherwise the query fails.

### Children and ChildrenMap

The `Children` and `ChildrenMap` methods no longer return errors. Instead, in the event of the underlying being invalid, a `nil` slice and empty map are returned respectively. If you explicit type checking is required you can still use `foo, ok := obj.Data().([]interface)`.

### Serialising Invalid Types

In V1 attempting to serialise with `Bytes`, `String`, etc, with an invalid structure would result in an empty object `{}`. This behaviour was unintuitive and in V2 `null` will be returned instead.