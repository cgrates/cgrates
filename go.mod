module github.com/cgrates/cgrates

go 1.23.0

// replace github.com/cgrates/radigo => ../radigo

// replace github.com/cgrates/rpcclient => ../rpcclient

require (
	github.com/Azure/go-amqp v1.0.2
	github.com/antchfx/xmlquery v1.3.11
	github.com/aws/aws-sdk-go v1.44.43
	github.com/blevesearch/bleve v1.0.14
	github.com/cgrates/aringo v0.0.0-20220525160735-b5990313d99e
	github.com/cgrates/baningo v0.0.0-20210413080722-004ffd5e429f
	github.com/cgrates/birpc v1.3.1-0.20211117095917-5b0ff29f3084
	github.com/cgrates/cron v0.0.0-20201022095836-3522d5b72c70
	github.com/cgrates/fsock v0.0.0-20230123160954-12cae14030cc
	github.com/cgrates/kamevapi v0.0.0-20240307160311-26273f03eedf
	github.com/cgrates/ltcache v0.0.0-20210405185848-da943e80c1ab
	github.com/cgrates/radigo v0.0.0-20210902121842-ea2f9a730627
	github.com/cgrates/rpcclient v0.0.0-20240628101047-cb29aae6b006
	github.com/cgrates/sipingo v1.0.1-0.20200514112313-699ebc1cdb8e
	github.com/creack/pty v1.1.20
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/elastic/elastic-transport-go/v8 v8.3.0
	github.com/elastic/go-elasticsearch/v8 v8.11.0
	github.com/ericlagergren/decimal v0.0.0-20221120152707-495c53812d05
	github.com/fiorix/go-diameter/v4 v4.0.4
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/mediocregopher/radix/v3 v3.8.1
	github.com/miekg/dns v1.1.57
	github.com/nats-io/nats-server/v2 v2.10.5
	github.com/nats-io/nats.go v1.31.0
	github.com/nyaruka/phonenumbers v1.1.9
	github.com/peterh/liner v1.2.2
	github.com/prometheus/client_golang v1.17.0
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/segmentio/kafka-go v0.4.44
	go.mongodb.org/mongo-driver v1.13.0
	golang.org/x/crypto v0.15.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/net v0.18.0
	golang.org/x/oauth2 v0.14.0
	google.golang.org/api v0.150.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/jackc/pgx/v5 v5.5.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231030173426-d783a09b4405 // indirect
)

require (
	cloud.google.com/go/compute v1.23.1 // indirect
	github.com/RoaringBitmap/roaring v0.4.23 // indirect
	github.com/antchfx/xpath v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.2 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/zap/v11 v11.0.14 // indirect
	github.com/blevesearch/zap/v12 v12.0.14 // indirect
	github.com/blevesearch/zap/v13 v13.0.6 // indirect
	github.com/blevesearch/zap/v14 v14.0.5 // indirect
	github.com/blevesearch/zap/v15 v15.0.3 // indirect
	github.com/cenkalti/hub v1.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cgrates/janusgo v0.0.0-20240503152118-188a408d7e73
	github.com/couchbase/ghistogram v0.1.0 // indirect
	github.com/couchbase/moss v0.1.0 // indirect
	github.com/couchbase/vellum v1.0.2 // indirect
	github.com/glycerine/go-unsnap-stream v0.0.0-20181221182339-f9677308dec2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/ishidawataru/sctp v0.0.0-20190922091402-408ec287e38c // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/jwt/v2 v2.5.3 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/steveyen/gtreap v0.1.0 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/ugorji/go/codec v1.2.11
	github.com/willf/bitset v1.1.10 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.4.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	nhooyr.io/websocket v1.8.11
)
