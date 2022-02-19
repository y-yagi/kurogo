module github.com/y-yagi/kurogo

go 1.17

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/y-yagi/goext v0.6.0
	github.com/y-yagi/rnotify v0.1.1-0.20210623015509-608d88a46c40
	golang.org/x/sys v0.0.0-20210927094055-39ccf1dd6fa6 // indirect
)

require (
	github.com/fsnotify/fsevents v0.1.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
)

replace github.com/fsnotify/fsnotify => github.com/y-yagi/fsnotify v1.4.10-0.20201227062311-078207fcf401
