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
