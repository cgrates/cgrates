module github.com/cgrates/cgrates

go 1.23.2

// replace github.com/cgrates/radigo => ../radigo

// replace github.com/cgrates/rpcclient => ../rpcclient

// replace github.com/cgrates/fsock => ../fsock

// replace github.com/cgrates/kamevapi => ../kamevapi

// replace github.com/cgrates/aringo => ../aringo

require (
	github.com/Azure/go-amqp v1.0.5
	github.com/antchfx/xmlquery v1.4.1
	github.com/aws/aws-sdk-go v1.55.5
	github.com/blevesearch/bleve/v2 v2.4.2
	github.com/cgrates/aringo v0.0.0-20220525160735-b5990313d99e
	github.com/cgrates/baningo v0.0.0-20210413080722-004ffd5e429f
	github.com/cgrates/birpc v1.3.1-0.20211117095917-5b0ff29f3084
	github.com/cgrates/cron v0.0.0-20201129173550-63ea3d835706
	github.com/cgrates/fsock v0.0.0-20240522220429-b6cc1d96fd2b
	github.com/cgrates/janusgo v0.0.0-20240503152118-188a408d7e73
	github.com/cgrates/kamevapi v0.0.0-20240307160311-26273f03eedf
	github.com/cgrates/ltcache v0.0.0-20240411152156-e673692056db
	github.com/cgrates/radigo v0.0.0-20240123163129-491c899df727
	github.com/cgrates/rpcclient v0.0.0-20240816141816-52dd1074499e
	github.com/cgrates/sipingo v1.0.1-0.20200514112313-699ebc1cdb8e
	github.com/creack/pty v1.1.23
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/elastic/elastic-transport-go/v8 v8.6.0
	github.com/elastic/go-elasticsearch/v8 v8.14.0
	github.com/ericlagergren/decimal v0.0.0-20240411145413-00de7ca16731
	github.com/fiorix/go-diameter/v4 v4.0.4
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/go-cmp v0.6.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/mediocregopher/radix/v3 v3.8.1
	github.com/miekg/dns v1.1.62
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nats-io/nats.go v1.37.0
	github.com/nyaruka/phonenumbers v1.4.0
	github.com/peterh/liner v1.2.2
	github.com/prometheus/procfs v0.12.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/ugorji/go/codec v1.2.12
	go.mongodb.org/mongo-driver v1.16.1
	golang.org/x/crypto v0.26.0
	golang.org/x/net v0.28.0
	golang.org/x/oauth2 v0.22.0
	google.golang.org/api v0.192.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.11
	nhooyr.io/websocket v1.8.17
)

require (
	cloud.google.com/go/auth v0.8.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.3 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240730163845-b1a4ccb954bf // indirect
)

require (
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	github.com/RoaringBitmap/roaring v1.9.4 // indirect
	github.com/antchfx/xpath v1.3.1 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/blevesearch/bleve_index_api v1.1.11 // indirect
	github.com/blevesearch/geo v0.1.20 // indirect
	github.com/blevesearch/go-faiss v1.0.20 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/goleveldb v1.0.1 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.2.15 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.10 // indirect
	github.com/blevesearch/zapx/v11 v11.3.10 // indirect
	github.com/blevesearch/zapx/v12 v12.3.10 // indirect
	github.com/blevesearch/zapx/v13 v13.3.10 // indirect
	github.com/blevesearch/zapx/v14 v14.3.10 // indirect
	github.com/blevesearch/zapx/v15 v15.3.13 // indirect
	github.com/blevesearch/zapx/v16 v16.1.5 // indirect
	github.com/cenkalti/hub v1.0.2 // indirect
	github.com/couchbase/ghistogram v0.1.0 // indirect
	github.com/couchbase/moss v0.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/geo v0.0.0-20230421003525-6adc56603217 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/ishidawataru/sctp v0.0.0-20190922091402-408ec287e38c // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/prometheus/client_golang v1.19.1
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	golang.org/x/xerrors v0.0.0-20240716161551-93cc26a95ae9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240730163845-b1a4ccb954bf // indirect
	google.golang.org/grpc v1.64.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
