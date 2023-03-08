package main

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	importPath = "go.flipt.io/flipt/sdk"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		// We have some use of the optional feature in our proto3 definitions.
		// This broadcasts that our plugin supports it and hides the generated
		// warning.
		gen.SupportedFeatures |= uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		var types [][2]string
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			typ, transport := generateSubSDK(gen, f)
			types = append(types, [2]string{typ, transport})
		}

		generateSDK(gen, types)

		return nil
	})
}

func generateSDK(gen *protogen.Plugin, types [][2]string) {
	g := gen.NewGeneratedFile("sdk.gen.go", importPath)
	g.P("// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.")
	g.P()
	g.P("package sdk")
	g.P()
	g.P("type Transport interface {")
	for _, t := range types {
		g.P(t[1], "() ", t[1])
	}
	g.P("}")
	g.P()

	g.P(sdkBase)
	g.P()

	for _, t := range types {
		g.P("func (s SDK) ", t[0], "()", t[0], "{")
		g.P("return ", t[0], "{transport: s.transport.", t[1], "()}")
		g.P("}\n")
	}
}

// generateSubSDK generates a .pb.sdk.go file containing a single SDK structure
// which represents an entire package from within the entire Flipt SDK API.
func generateSubSDK(gen *protogen.Plugin, file *protogen.File) (typ, transport string) {
	filename := string(file.GoPackageName) + ".pb.sdk.go"
	g := gen.NewGeneratedFile(filename, importPath)
	g.P("// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.")
	g.P()
	g.P("package sdk")
	g.P()

	ident := func(pkg string) func(name string) string {
		return func(name string) string {
			return g.QualifiedGoIdent(protogen.GoIdent{
				GoImportPath: protogen.GoImportPath(pkg),
				GoName:       name,
			})
		}
	}

	context := ident("context")

	// define transport interface
	transport = strings.Title(string(file.GoPackageName)) + "Transport"
	g.P("type ", transport, " interface {")
	for _, srv := range file.Services {
		for _, method := range srv.Methods {
			g.P(method.GoName, "(", context("Context"), ", *", method.Input.GoIdent, ") (*", method.Output.GoIdent, ", error)")
		}
	}
	g.P("}\n")

	// define client structure
	typ = strings.Title(string(file.GoPackageName))
	g.P("type ", typ, " struct {")
	g.P("transport ", transport)
	g.P("}\n")
	for _, srv := range file.Services {
		for _, method := range srv.Methods {
			var (
				signature       = []any{"func (x *", typ, ") ", method.GoName, "(ctx ", context("Context")}
				returnStatement = []any{"x.transport.", method.GoName, "(ctx, "}
			)

			switch len(method.Input.Fields) {
			case 0:
				returnStatement = append(returnStatement, "&", method.Input.GoIdent, "{})")
			case 1:
				field := method.Input.Fields[0]
				v := variableCase(field.GoName)

				var kind []any
				switch field.Desc.Kind() {
				case protoreflect.MessageKind:
					kind = []any{"*", field.Message.GoIdent}
				default:
					kind = []any{field.Desc.Kind()}
				}

				signature = append(append(signature, ", ", v, " "), kind...)
				returnStatement = append(returnStatement, "&", method.Input.GoIdent, "{", field.GoName, ": ", v, "})")
			default:
				signature = append(signature, ", v *", method.Input.GoIdent)
				returnStatement = append(returnStatement, "v)")
			}

			if method.Output.GoIdent.GoImportPath != "google.golang.org/protobuf/types/known/emptypb" {
				g.P(append(signature, ") (*", method.Output.GoIdent, ", error) {")...)
				g.P(append([]any{"return "}, returnStatement...)...)
			} else {
				g.P(append(signature, ") error {")...)
				g.P(append([]any{"_, err := "}, returnStatement...)...)
				g.P("return err")
			}

			g.P("}\n")
		}
	}
	return
}

func variableCase(v string) string {
	return strings.ToLower(v[:1]) + v[1:]
}

const sdkBase = `// ClientTokenProvider is a type which when requested provides a
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
	transport Transport
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

    for _, opt := range opts { opt(&sdk) }

    return sdk
}`
