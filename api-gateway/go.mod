module github.com/Abdurahmanit/GroupProject/api-gateway

go 1.23.4

require (
	github.com/Abdurahmanit/GroupProject/listing-service v0.0.0
	github.com/Abdurahmanit/GroupProject/user-service v0.0.0-20250529172304-38141d74e416
	github.com/go-chi/chi/v5 v5.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/spf13/viper v1.20.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.72.2
)

replace github.com/Abdurahmanit/GroupProject/listing-service => ../listing-service

require (
	github.com/Abdurahmanit/GroupProject/review-service v0.0.0-20250529233351-364af3648168
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
