module github.com/zhenruyan/rsqlite/examples

go 1.23.0

toolchain go1.23.11

require (
	github.com/uptrace/bun v1.2.14
	github.com/uptrace/bun/dialect/sqlitedialect v1.2.14
	github.com/uptrace/bun/extra/bundebug v1.2.14
	github.com/zhenruyan/rsqlite v0.0.0-20250711073451-dfe7507654cd
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
	xorm.io/xorm v1.3.9
)

replace github.com/zhenruyan/rsqlite => ../ 

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/goccy/go-json v0.8.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/rqlite/gorqlite v0.0.0-20230708021416-2acd02b70b79 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	xorm.io/builder v0.3.11-0.20220531020008-1bd24a7dc978 // indirect
)
