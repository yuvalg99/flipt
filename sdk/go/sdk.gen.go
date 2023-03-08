// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

type Transport interface {
	AuthTransport() AuthTransport
	FliptTransport() FliptTransport
	MetaTransport() MetaTransport
}

// ClientTokenProvider is a type which when requested provides a
// client token which can be used to authenticate RPC/API calls
// invoked through the SDK.
type ClientTokenProvider interface {
	ClientToken() (string, error)
}

// SDK is the definition of Flipt's Go SDK.
// It depends on a pluggable transport implementation and exposes
// a consistent API surface area across both transport implementations.
// It also provides consistent client-side instrumentation and authentication
// lifecycle support.
type SDK struct {
	transport     Transport
	tokenProvider ClientTokenProvider
}

// Option is a functional option which configures the Flipt SDK.
type Option func(*SDK)

// WithClientTokenProviders returns an Option which configures
// any supplied SDK with the provided ClientTokenProvider.
func WithClientTokenProvider(p ClientTokenProvider) Option {
	return func(s *SDK) {}
}

// New constructs and configures a Flipt SDK instance from
// the provided Transport implementation and options.
func New(t Transport, opts ...Option) SDK {
	sdk := SDK{transport: t}

	for _, opt := range opts {
		opt(&sdk)
	}

	return sdk
}

func (s SDK) Auth() Auth {
	return Auth{transport: s.transport.AuthTransport()}
}

func (s SDK) Flipt() Flipt {
	return Flipt{transport: s.transport.FliptTransport()}
}

func (s SDK) Meta() Meta {
	return Meta{transport: s.transport.MetaTransport()}
}
