# value

Value in GO

All instances are immutable and good for multi-threading or go-routing.

Deterministic serialization that guarantee consistent results before hashing.

### List
```
v := value.NewList()

v = v.Insert(value.Boolean(true))
v = v.Insert(value.Long(123))
v = v.Insert(value.Double(-12.34))
v = v.Insert(value.Utf8("text"))
v = v.Insert(value.Raw([]byte{0, 1, 2}, false))

mp, _ := value.Pack(v)

c, err := value.Unpack(mp, false)
if err != nil {
    t.Errorf("unpack fail %v", err)
}

require.True(t, v.Equal(c))
```

### Map
```
b = value.NewMap()

c := value.NewMap()
c.Put("5", value.Long(5))

b.Put("name", value.Utf8("name"))
b.Put("123", value.Long(123))
b.Put("map", c)
```
