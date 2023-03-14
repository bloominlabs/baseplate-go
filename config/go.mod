module github.com/bloominlabs/baseplate-go/config

go 1.20

replace github.com/bloominlabs/baseplate-go/config/filesystem => ./filesystem/

require (
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230313062718-f967fe864fd1
	github.com/grafana/dskit v0.0.0-20230307154039-ee798e84baf0
	github.com/pelletier/go-toml/v2 v2.0.7
	github.com/rs/zerolog v1.29.0
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	google.golang.org/protobuf v1.29.1 // indirect
)
