module github.com/clinia/x

go 1.24

require (
	github.com/arangodb/go-driver v1.6.0
	github.com/bradleyjkemp/cupaloy/v2 v2.8.0
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/ristretto v0.1.1
	github.com/elastic/go-elasticsearch/v8 v8.17.1
	github.com/felixge/httpsnoop v1.0.4
	github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/inhies/go-bytesize v0.0.0-20220417184213-4913239db9cf
	github.com/knadh/koanf v1.5.0
	github.com/ory/jsonschema/v3 v3.0.8
	github.com/pelletier/go-toml v1.9.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.20.4
	github.com/rs/cors v1.9.0
	github.com/samber/lo v1.52.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cast v1.5.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.16.0
	github.com/tidwall/sjson v1.2.5
	github.com/twmb/franz-go v1.17.1
	github.com/twmb/franz-go/pkg/kadm v1.13.0
	github.com/twmb/franz-go/plugin/kotel v1.5.0
	github.com/veqryn/slog-context v0.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.44.0
	go.opentelemetry.io/contrib/propagators/b3 v1.17.0
	go.opentelemetry.io/contrib/propagators/jaeger v1.17.0
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.11.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/exporters/jaeger v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.18.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.18.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.18.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
	go.opentelemetry.io/proto/otlp v1.0.0
	go.uber.org/goleak v1.2.1
	golang.org/x/sync v0.11.0
	google.golang.org/grpc v1.66.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
)

require (
	github.com/Shopify/toxiproxy/v2 v2.11.0
	github.com/arangodb/go-velocypack v0.0.0-20200318135517-5af53c29c67e // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.6.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.2.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.17.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jandelgado/gcov2lcov v1.0.6 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/nyaruka/phonenumbers v1.2.2 // indirect
	github.com/ory/go-acc v0.2.9-0.20230103102148-6b1c9a70dbbe // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.16.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.41.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.41.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.41.0
	go.opentelemetry.io/otel/exporters/prometheus v0.44.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.41.0
	go.opentelemetry.io/otel/metric v1.30.0
	go.opentelemetry.io/otel/sdk/metric v1.30.0
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/exp v0.0.0-20250128182459-e0ece0dbea4c
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
