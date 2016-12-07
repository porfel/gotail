# gotail
Go package for tailing file

```
t, err := tail.NewTail("/var/log/nginx/access.log", 0, time.Second)
defer t.Close()
for {
	line := t.ReadLine()
	// process line
}
```
