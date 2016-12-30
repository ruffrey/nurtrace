Add this line to `main()`

```go
defer profile.Start(profile.MemProfile).Stop()
```

then build the executable.

```bash
go build
```

then run the profiler. it will output a location when the exe stops.

```bash
go tool pprof -pdf ./shake /var/folders/v1/1sr1bhps3_7gyg8g30hr_9100000gn/T/profile563441561/mem.pprof >> profile.pdf
```

or interact:
```bash
go tool pprof ./shake /var/folders/cn/90t4rc5x45j9qwr3z5gxvz840000gn/T/profile479444033/mem.pprof

> top10
> list potential.NewSynapse
```
