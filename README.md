# value

Value in GO

### List
```
v := value.List()

v.Insert(value.Boolean(true))
v.Insert(value.Long(123))
v.Insert(value.Double(-12.34))
v.Insert(value.Utf8("text"))
v.Insert(value.Raw([]byte{0, 1, 2}, false))

mp, _ := value.Pack(v)

c, err := value.Unpack(mp, false)
if err != nil {
    t.Errorf("unpack fail %v", err)
}

require.True(t, v.Equal(c))
```

### Map
```
b = value.Map()

c := value.Map()
c.Put("5", value.Long(5))

b.Put("name", value.Utf8("name"))
b.Put("123", value.Long(123))
b.Put("map", c)
```
