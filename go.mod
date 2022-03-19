module github.com/je4/primobridge/v2

go 1.17

replace github.com/je4/primobridge/v2 => ./

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/bluele/gcache v0.0.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/je4/utils/v2 v2.0.6
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require github.com/felixge/httpsnoop v1.0.1 // indirect
