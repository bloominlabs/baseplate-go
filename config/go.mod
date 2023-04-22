module github.com/bloominlabs/baseplate-go/config

go 1.20

replace github.com/bloominlabs/baseplate-go/config/filesystem => ./filesystem/

require (
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-00010101000000-000000000000
	github.com/pelletier/go-toml/v2 v2.0.7
	github.com/rs/zerolog v1.29.1
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/sys v0.6.0 // indirect
)
