module github.com/bloominlabs/baseplate-go/config

go 1.20

replace github.com/bloominlabs/baseplate-go/config/filesystem => ./filesystem/

require (
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230419034715-89fcb81782b1
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/rs/zerolog v1.31.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
