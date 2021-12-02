module github.com/pydio/cells/v4

go 1.16

require (
	cloud.google.com/go/kms v1.1.0 // indirect
	github.com/ajvb/kala v0.8.4
	github.com/allegro/bigcache v1.2.1
	github.com/beevik/ntp v0.3.0
	github.com/blevesearch/bleve v1.0.14
	github.com/c2fo/testify v0.0.0-20150827203832-fba96363964a // indirect
	github.com/caddyserver/caddy v1.0.5
	github.com/caddyserver/caddy/v2 v2.4.5
	github.com/cskr/pubsub v1.0.2
	github.com/disintegration/imaging v1.6.2
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsnotify/fsnotify v1.5.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/errors v0.20.1
	github.com/go-openapi/loads v0.21.0
	github.com/go-openapi/spec v0.20.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/gosimple/slug v1.11.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/h2non/filetype v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/jaytaylor/go-hostsfile v0.0.0-20201026230151-f581673a59cf
	github.com/jcuga/golongpoll v1.3.0
	github.com/jinzhu/copier v0.3.2
	github.com/jmoiron/sqlx v1.3.4
	github.com/json-iterator/go v1.1.12
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/karrick/godirwalk v1.16.1
	github.com/krolaw/zipstream v0.0.0-20180621105154-0a2661891f94
	github.com/kylelemons/godebug v1.1.0
	github.com/livekit/protocol v0.10.0
	github.com/lpar/gzipped v1.1.0 // indirect
	github.com/lucas-clemente/quic-go v0.24.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/micro/micro/v3 v3.6.0
	github.com/minio/cli v1.22.0
	github.com/minio/madmin-go v1.1.12
	github.com/minio/minio v0.0.0-20211121184130-c791de0e1eae
	github.com/minio/minio-go/v7 v7.0.15
	github.com/mitchellh/mapstructure v1.4.2
	github.com/mssola/user_agent v0.5.3
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/nicksnyder/go-i18n v1.10.0
	github.com/ory/fosite v0.40.3-0.20211101181407-30e8cb92e53c
	github.com/ory/hydra v1.10.7
	github.com/ory/ladon v1.2.0
	github.com/ory/x v0.0.303
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/philopon/go-toposort v0.0.0-20170620085441-9be86dbd762f
	github.com/pkg/errors v0.9.1
	github.com/pydio/melody v0.0.0-20190928133520-4271c6513fb6
	github.com/rjeczalik/notify v0.9.2
	github.com/robertkrimen/otto v0.0.0-20211024170158-b87d35c0b86f
	github.com/rs/cors v1.8.0
	github.com/rs/xid v1.3.0
	github.com/rubenv/sql-migrate v0.0.0-20211023115951-9f02b1e13857
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd
	github.com/scottleedavis/go-exif-remove v0.0.0-20190908021517-58bdbaac8636
	github.com/sendgrid/rest v2.6.5+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.10.3+incompatible
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/afero v1.6.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally v3.4.2+incompatible
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/client/v3 v3.5.1 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/net v0.0.0-20211020060615-d418f374d309
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/text v0.3.7
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	google.golang.org/genproto v0.0.0-20211020151524-b7c3a969101a
	google.golang.org/grpc v1.41.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/doug-martin/goqu.v4 v4.2.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/gorp.v1 v1.7.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/klog/v2 v2.30.0 // indirect
)

// replace github.com/minio/minio => /Users/charles/Sources/go/src/github.com/pydio/minio
replace github.com/minio/minio => github.com/pydio/minio v0.0.0-20211122154507-cf0bc00fb0b8
