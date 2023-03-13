module github.com/bloominlabs/baseplate-go/config/server

go 1.20

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem/

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-00010101000000-000000000000
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.29.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/sys v0.5.0 // indirect
)
