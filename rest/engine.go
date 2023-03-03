package rest

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/zeromicro/go-zero/core/codec"
	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/rest/chain"
	"github.com/zeromicro/go-zero/rest/handler"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/internal"
	"github.com/zeromicro/go-zero/rest/internal/response"
)

// use 1000m to represent 100%
const topCpuUsage = 1000

// ErrSignatureConfig is an error that indicates bad config for signature.
var ErrSignatureConfig = errors.New("bad config for Signature")

type engine struct {
	conf                 RestConf
	routes               []featuredRoutes
	unauthorizedCallback handler.UnauthorizedCallback
	unsignedCallback     handler.UnsignedCallback
	chain                chain.Chain
	middlewares          []Middleware
	shedder              load.Shedder
	priorityShedder      load.Shedder
	tlsConfig            *tls.Config
}

func newEngine(c RestConf) *engine {
	svr := &engine{
		conf: c,
	}
	if c.CpuThreshold > 0 {
		svr.shedder = load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		svr.priorityShedder = load.NewAdaptiveShedder(load.WithCpuThreshold(
			(c.CpuThreshold + topCpuUsage) >> 1))
	}

	return svr
}

func (ng *engine) addRoutes(r featuredRoutes) {
	ng.routes = append(ng.routes, r)
}

func (ng *engine) appendAuthHandler(fr featuredRoutes, chn chain.Chain,
	verifier func(chain.Chain) chain.Chain) chain.Chain {
	if fr.jwt.enabled {
		if len(fr.jwt.prevSecret) == 0 {
			chn = chn.Append(handler.Authorize(fr.jwt.secret,
				handler.WithUnauthorizedCallback(ng.unauthorizedCallback)))
		} else {
			chn = chn.Append(handler.Authorize(fr.jwt.secret,
				handler.WithPrevSecret(fr.jwt.prevSecret),
				handler.WithUnauthorizedCallback(ng.unauthorizedCallback)))
		}
	}

	return verifier(chn)
}

func (ng *engine) bindFeaturedRoutes(router httpx.Router, fr featuredRoutes, metrics *stat.Metrics) error {
	verifier, err := ng.signatureVerifier(fr.signature)
	if err != nil {
		return err
	}

	for _, route := range fr.routes {
		if err := ng.bindRoute(fr, router, metrics, route, verifier); err != nil {
			return err
		}
	}

	return nil
}

func (ng *engine) bindRoute(fr featuredRoutes, router httpx.Router, metrics *stat.Metrics,
	route Route, verifier func(chain.Chain) chain.Chain) error {
	chn := ng.chain
	if chn == nil {
		chn = chain.New(
			handler.TracingHandler(ng.conf.Name, route.Path),
			ng.getLogHandler(),
			handler.PrometheusHandler(route.Path),
			handler.MaxConns(ng.conf.MaxConns),
			handler.BreakerHandler(route.Method, route.Path, metrics),
			handler.SheddingHandler(ng.getShedder(fr.priority), metrics),
			handler.TimeoutHandler(ng.checkedTimeout(fr.timeout)),
			handler.RecoverHandler,
			handler.MetricHandler(metrics),
			handler.MaxBytesHandler(ng.checkedMaxBytes(fr.maxBytes)),
			handler.GunzipHandler,
		)
	}

	chn = ng.appendAuthHandler(fr, chn, verifier)

	for _, middleware := range ng.middlewares {
		chn = chn.Append(convertMiddleware(middleware))
	}
	handle := chn.ThenFunc(route.Handler)

	return router.Handle(route.Method, route.Path, handle)
}

func (ng *engine) bindRoutes(router httpx.Router) error {
	metrics := ng.createMetrics()

	for _, fr := range ng.routes {
		if err := ng.bindFeaturedRoutes(router, fr, metrics); err != nil {
			return err
		}
	}

	return nil
}

func (ng *engine) checkedMaxBytes(bytes int64) int64 {
	if bytes > 0 {
		return bytes
	}

	return ng.conf.MaxBytes
}

func (ng *engine) checkedTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}

	return time.Duration(ng.conf.Timeout) * time.Millisecond
}

func (ng *engine) createMetrics() *stat.Metrics {
	var metrics *stat.Metrics

	if len(ng.conf.Name) > 0 {
		metrics = stat.NewMetrics(ng.conf.Name)
	} else {
		metrics = stat.NewMetrics(fmt.Sprintf("%s:%d", ng.conf.Host, ng.conf.Port))
	}

	return metrics
}

func (ng *engine) getLogHandler() func(http.Handler) http.Handler {
	if ng.conf.Verbose {
		return handler.DetailedLogHandler
	}

	return handler.LogHandler
}

func (ng *engine) getShedder(priority bool) load.Shedder {
	if priority && ng.priorityShedder != nil {
		return ng.priorityShedder
	}

	return ng.shedder
}

// notFoundHandler returns a middleware that handles 404 not found requests.
func (ng *engine) notFoundHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chn := chain.New(
			handler.TracingHandler(ng.conf.Name, ""),
			ng.getLogHandler(),
		)

		var h http.Handler
		if next != nil {
			h = chn.Then(next)
		} else {
			h = chn.Then(http.NotFoundHandler())
		}

		cw := response.NewHeaderOnceResponseWriter(w)
		h.ServeHTTP(cw, r)
		cw.WriteHeader(http.StatusNotFound)
	})
}

func (ng *engine) print() {
	var routes []string

	for _, fr := range ng.routes {
		for _, route := range fr.routes {
			routes = append(routes, fmt.Sprintf("%s %s", route.Method, route.Path))
		}
	}

	sort.Strings(routes)

	fmt.Println("Routes:")
	for _, route := range routes {
		fmt.Printf("  %s\n", route)
	}
}

func (ng *engine) setTlsConfig(cfg *tls.Config) {
	ng.tlsConfig = cfg
}

func (ng *engine) setUnauthorizedCallback(callback handler.UnauthorizedCallback) {
	ng.unauthorizedCallback = callback
}

func (ng *engine) setUnsignedCallback(callback handler.UnsignedCallback) {
	ng.unsignedCallback = callback
}

func (ng *engine) signatureVerifier(signature signatureSetting) (func(chain.Chain) chain.Chain, error) {
	if !signature.enabled {
		return func(chn chain.Chain) chain.Chain {
			return chn
		}, nil
	}

	if len(signature.PrivateKeys) == 0 {
		if signature.Strict {
			return nil, ErrSignatureConfig
		}

		return func(chn chain.Chain) chain.Chain {
			return chn
		}, nil
	}

	decrypters := make(map[string]codec.RsaDecrypter)
	for _, key := range signature.PrivateKeys {
		fingerprint := key.Fingerprint
		file := key.KeyFile
		decrypter, err := codec.NewRsaDecrypter(file)
		if err != nil {
			return nil, err
		}

		decrypters[fingerprint] = decrypter
	}

	return func(chn chain.Chain) chain.Chain {
		if ng.unsignedCallback != nil {
			return chn.Append(handler.ContentSecurityHandler(
				decrypters, signature.Expiry, signature.Strict, ng.unsignedCallback))
		}

		return chn.Append(handler.ContentSecurityHandler(decrypters, signature.Expiry, signature.Strict))
	}, nil
}

func (ng *engine) start(router httpx.Router) error {
	if err := ng.bindRoutes(router); err != nil {
		return err
	}

	if len(ng.conf.CertFile) == 0 && len(ng.conf.KeyFile) == 0 {
		return internal.StartHttp(ng.conf.Host, ng.conf.Port, router, ng.withTimeout())
	}

	return internal.StartHttps(ng.conf.Host, ng.conf.Port, ng.conf.CertFile,
		ng.conf.KeyFile, router, func(svr *http.Server) {
			if ng.tlsConfig != nil {
				svr.TLSConfig = ng.tlsConfig
			}
		}, ng.withTimeout())
}

func (ng *engine) use(middleware Middleware) {
	ng.middlewares = append(ng.middlewares, middleware)
}

func (ng *engine) withTimeout() internal.StartOption {
	return func(svr *http.Server) {
		timeout := ng.conf.Timeout
		if timeout > 0 {
			// factor 0.8, to avoid clients send longer content-length than the actual content,
			// without this timeout setting, the server will time out and respond 503 Service Unavailable,
			// which triggers the circuit breaker.
			svr.ReadTimeout = 4 * time.Duration(timeout) * time.Millisecond / 5
			// factor 0.9, to avoid clients not reading the response
			// without this timeout setting, the server will time out and respond 503 Service Unavailable,
			// which triggers the circuit breaker.
			svr.WriteTimeout = 9 * time.Duration(timeout) * time.Millisecond / 10
		}
	}
}

func convertMiddleware(ware Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return ware(next.ServeHTTP)
	}
}
