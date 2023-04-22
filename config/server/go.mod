module github.com/bloominlabs/baseplate-go/config/server

go 1.20

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem/

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230419034715-89fcb81782b1
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230419034715-89fcb81782b1
	github.com/rs/zerolog v1.29.1
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	golang.org/x/sys v0.7.0 // indirect
)
