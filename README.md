# Zero
toolkit with the most common functions for backend development

## Examples
Start an http server
```
zero.Handle("/test", func(srv *zero.Server) {
  name := srv.GetParam("name")
  srv.HTML([]byte("Hello "+name))
})
zero.Serve("8080")
```

Start websocket server
```
zero.Handle("/websocket", func(srv *zero.Server) {
  soc := srv.UpgradeWS()
  soc.HandleAll()
})
zero.Serve("8080")
```
