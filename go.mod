module github.com/y-yagi/kurogo

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/fatih/color v1.12.0
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/y-yagi/goext v0.6.0
	github.com/y-yagi/rnotify v0.1.1-0.20210623015509-608d88a46c40
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
)

replace github.com/fsnotify/fsnotify => github.com/y-yagi/fsnotify v1.4.10-0.20201227062311-078207fcf401
