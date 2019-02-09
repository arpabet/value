# val
Value in GO

### List
```
v := val.List()

v.Insert(val.Boolean(true))
v.Insert(val.Long(123))
v.Insert(val.Double(-12.34))
v.Insert(val.Utf8("text"))
v.Insert(val.Raw([]byte{0, 1, 2}, false))

mp, _ := val.Pack(v)

c, err := val.Unpack(mp, false)
if err != nil {
    t.Errorf("unpack fail %v", err)
}

require.True(t, v.Equal(c))
```

### Map
```
b = val.Map()

c := val.Map()
c.Put("5", val.Long(5))

b.Put("name", val.Utf8("name"))
b.Put("123", val.Long(123))
b.Put("map", c)
```