module github.com/sorae42/ressdit

go 1.21

toolchain go1.22.5

require (
	github.com/PuerkitoBio/goquery v1.9.2
	github.com/cameronstanley/go-reddit v0.0.0-20170423222116-4bfac7ea95af
	github.com/gabriel-vasile/mimetype v1.4.5
	github.com/getsentry/sentry-go v0.28.1
	github.com/go-shiori/go-readability v0.0.0-20240701094332-1070de7e32ef
	github.com/gorilla/feeds v1.2.0
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/joho/godotenv v1.5.1
	github.com/victorspringer/http-cache v0.0.0-20240523143319-7d9f48f8ab91
	golang.org/x/oauth2 v0.21.0
)

require (
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-redis/cache v6.4.0+incompatible // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jarcoal/httpmock v1.3.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/cameronstanley/go-reddit => ./pkg/reddit
