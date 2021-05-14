# Zero
toolkit with the most common functions for backend development

## Examples
Start an http server
```
zero.Handle("/test", func(srv *zero.Request) {
  name := req.GetParam("name")
  req.HTML([]byte("Hello "+name))
})
zero.Serve("8080")
```

Start websocket server
```
zero.Handle("/websocket", func(srv *zero.Request) {
  soc := req.UpgradeWS()
  soc.HandleAll()
})
zero.Serve("8080")
```

## Modules

### Stat module
Stat module allow you to absorb app stats and get them once every configured period of time
Example:
```
stat := zero.Stat{}
stat.Init(time.Minute*5, func(eventName string, counter zero.StatCounter) {
  fmt.Println("for every key like", eventName, "you will ge counter object")
})

// adding event to stat
stat.Inc("some_event_name")
```
