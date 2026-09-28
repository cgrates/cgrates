module github.com/cgrates/cgrates

go 1.27.1

// replace github.com/cgrates/radigo => ../radigo

// replace github.com/cgrates/rpcclient => ../rpcclient

// replace github.com/cgrates/fsock => ../fsock

// replace github.com/cgrates/kamevapi => ../kamevapi

// replace github.com/cgrates/aringo => ../aringo

require (
	github.com/Azure/go-amqp v1.7.0
	github.com/antchfx/xmlquery v1.5.1
	github.com/aws/aws-sdk-go v1.55.8
	github.com/blevesearch/bleve/v2 v2.4.2
	github.com/cgrates/aringo v0.0.0-20260324164818-495676900dbb
	github.com/cgrates/baningo v0.0.0-20260907204338-d8ed5b43f1b4
	github.com/cgrates/birpc v1.3.1-0.20211117095917-5b0ff29f3084
	github.com/cgrates/cron v0.0.0-20201129173550-63ea3d835706
	github.com/cgrates/fsock v0.0.0-20260907204233-094bc3a49233
	github.com/cgrates/go-diameter v0.0.0-20260512102405-f9dcf589d4ac
	github.com/cgrates/janusgo v0.0.0-20240503152118-188a408d7e73
	github.com/cgrates/kamevapi v0.0.0-20260728044114-dd837c919733
	github.com/cgrates/ltcache v0.0.0-20260728043623-33b994321107
	github.com/cgrates/radigo v0.0.0-20240123163129-491c899df727
	github.com/cgrates/rpcclient v0.0.0-20260910075308-4b8faf373e78
	github.com/cgrates/sipingo v1.0.1-0.20200514112313-699ebc1cdb8e
	github.com/creack/pty v1.1.24
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/elastic/elastic-transport-go/v8 v8.11.0
	github.com/ericlagergren/decimal v0.0.0-20240411145413-00de7ca16731
	github.com/fsnotify/fsnotify v1.10.1
	github.com/go-sql-driver/mysql v1.10.1
	github.com/google/go-cmp v0.7.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/miekg/dns v1.1.73
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nats-io/nats.go v1.54.0
	github.com/nyaruka/phonenumbers v1.8.1
	github.com/peterh/liner v1.2.2
	github.com/prometheus/client_golang v1.24.1
	github.com/prometheus/procfs v0.22.0
	github.com/rabbitmq/amqp091-go v1.15.0
	github.com/redis/rueidis v1.0.78
	github.com/twmb/franz-go v1.22.1
	github.com/twmb/franz-go/pkg/kadm v1.19.0
	github.com/ugorji/go/codec v1.2.14
	go.mongodb.org/mongo-driver v1.17.10
	golang.org/x/crypto v0.57.0
	golang.org/x/net v0.59.0
	golang.org/x/oauth2 v0.37.0
	google.golang.org/api v0.299.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.3
	gorm.io/gorm v1.31.2
	nhooyr.io/websocket v1.8.17
)

require (
	cloud.google.com/go/auth v0.23.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.1 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/RoaringBitmap/roaring v1.9.4 // indirect
	github.com/antchfx/xpath v1.3.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
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
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/couchbase/ghistogram v0.1.0 // indirect
	github.com/couchbase/moss v0.2.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/geo v0.0.0-20230421003525-6adc56603217 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.10 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.22 // indirect
	github.com/googleapis/gax-go/v2 v2.24.1 // indirect
	github.com/ishidawataru/sctp v0.0.0-20251114114122-19ddcbc6aae2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.30 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.14.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
