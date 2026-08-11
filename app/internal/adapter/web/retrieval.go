package web

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/free-llms-foundation/retrieval-go"
)

type RetrieverConfig struct {
	ProxyLogin          string `env:"PROXY_LOGIN" validate:"required_with=ProxyHost"`
	ProxyPassword       string `env:"PROXY_PASSWORD" validate:"required_with=ProxyHost"`
	ProxyHost           string `env:"PROXY_HOST"`
	ProxyPort           string `env:"PROXY_PORT" validate:"required_with=ProxyHost"`
	ProxyScheme         string `env:"PROXY_SCHEME" validate:"required_with=ProxyHost"`
	MaxIdleConnsPerHost int    `env:"MAX_IDLE_CONNS_PER_HOST" envDefault:"100" validate:"gt=0"`
}

type Retrieval struct {
	client *retrieval.Client
}

func New(client *retrieval.Client) *Retrieval {
	return &Retrieval{
		client: client,
	}
}

func InitClient(cfg *RetrieverConfig) (*retrieval.Client, error) {
	opts := []retrieval.Option{
		retrieval.WithMaxIdleConnsPerHost(cfg.MaxIdleConnsPerHost),
		retrieval.WithBrowserRotation(),
		retrieval.WithDisableKeepAlive(),
	}

	if cfg.ProxyHost != "" {
		opts = append(opts, retrieval.WithProxyFactory(proxyFactory(cfg)))
	}

	return retrieval.New(opts...)
}

func proxyFactory(cfg *RetrieverConfig) func() string {
	prefix := strconv.FormatUint(uint64(time.Now().UnixNano()), 36)
	var counter uint64

	return func() string {
		num := atomic.AddUint64(&counter, 1)
		sessionID := prefix + "_" + strconv.FormatUint(num, 36)

		return cfg.ProxyScheme + "://" +
			cfg.ProxyLogin + "__session-" + sessionID + ":" +
			cfg.ProxyPassword + "@" +
			cfg.ProxyHost + ":" +
			cfg.ProxyPort
	}
}
