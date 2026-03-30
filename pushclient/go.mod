module github.com/Kimenzo/anytype-push-server/pushclient

go 1.25.7

require (
	github.com/Kimenzo/any-sync v0.11.4
	github.com/planetscale/vtprotobuf v0.6.1-0.20250313105119-ba97887b0a25
	google.golang.org/protobuf v1.36.11
	storj.io/drpc v0.0.34
)

require github.com/zeebo/errs v1.4.0 // indirect

replace github.com/Kimenzo/any-sync => ../../any-sync
