module github.com/pydio/cells/v4

go 1.16

require (
	github.com/ajvb/kala v0.8.4
	github.com/allegro/bigcache v1.2.1
	github.com/beevik/ntp v0.3.0
	github.com/blevesearch/bleve v1.0.14
	github.com/c2fo/testify v0.0.0-20150827203832-fba96363964a // indirect
	github.com/caddyserver/caddy v1.0.5
	github.com/caddyserver/caddy/v2 v2.4.5
	github.com/cskr/pubsub v1.0.2
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/disintegration/imaging v1.6.2
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsnotify/fsnotify v1.5.1
	github.com/go-openapi/errors v0.20.1
	github.com/go-openapi/loads v0.21.0
	github.com/go-openapi/spec v0.20.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/gosimple/slug v1.11.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/h2non/filetype v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jaytaylor/go-hostsfile v0.0.0-20201026230151-f581673a59cf
	github.com/jcuga/golongpoll v1.3.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/json-iterator/go v1.1.12
	github.com/karrick/godirwalk v1.16.1
	github.com/krolaw/zipstream v0.0.0-20180621105154-0a2661891f94
	github.com/lucas-clemente/quic-go v0.24.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/matcornic/hermes/v2 v2.1.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/micro/micro/v3 v3.6.0
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/nicksnyder/go-i18n v1.10.0
	github.com/ory/hydra v1.10.7
	github.com/ory/ladon v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rjeczalik/notify v0.9.2
	github.com/robertkrimen/otto v0.0.0-20211024170158-b87d35c0b86f
	github.com/rs/cors v1.8.0
	github.com/rs/xid v1.3.0
	github.com/rubenv/sql-migrate v0.0.0-20211023115951-9f02b1e13857
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd
	github.com/sendgrid/rest v2.6.5+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.10.3+incompatible // indirect
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/afero v1.6.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tinylib/msgp v1.1.7-0.20211026165309-e818a1881b0e // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally v3.4.2+incompatible
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/pkg/v3 v3.5.1
	go.etcd.io/etcd/pkg/v3 v3.5.1
	go.etcd.io/etcd/raft/v3 v3.5.1
	go.etcd.io/etcd/server/v3 v3.5.1
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/net v0.0.0-20211020060615-d418f374d309
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7
	google.golang.org/genproto v0.0.0-20211020151524-b7c3a969101a
	google.golang.org/grpc v1.41.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/doug-martin/goqu.v4 v4.2.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/gorp.v1 v1.7.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)
