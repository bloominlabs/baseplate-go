package observability

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	// runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/rs/zerolog/log"
)

func WithDefaultMetricOpts() []metric.Option {
	res, _ := WithDefaultResource(context.Background())
	return []metric.Option{
		metric.WithResource(res),
	}
}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
//
// NOTE: this temporarily returns a metric.MeterProvider while opentelemetry-go
// reworks the metrics API upstream to support globals. This will be updated in
// tandem when https://github.com/open-telemetry/opentelemetry-go/pull/2587 is
// deployed.
func InitMetricsProvider(addr string, credentials *credentials.TransportCredentials, opts ...metric.Option) (func(), error) {
	ctx := context.Background()

	grpcCreds := insecure.NewCredentials()
	if credentials != nil {
		grpcCreds = *credentials
	}

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	con, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(grpcCreds), grpc.FailOnNonTempDialError(true), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(con))
	if err != nil {
		return nil, fmt.Errorf("failed to create the collector metric client: %w", err)
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second))
	defaultOpts := []metric.Option{
		metric.WithReader(reader),
	}
	finalOpts := append(defaultOpts, opts...)
	provider := metric.NewMeterProvider(
		finalOpts...,
	)

	global.SetMeterProvider(provider)

	// if err = runtimemetrics.Start(runtimemetrics.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
	// 	return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	// }

	// TODO: this wont work because config.* are not being populated by build
	// meter := global.Meter(config.Service)
	// _ = metric.Must(meter).NewInt64GaugeObserver(
	// 	fmt.Sprintf("%s_build_info", config.Service),
	// 	func(_ context.Context, result metric.Int64ObserverResult) {
	// 		result.Observe(
	// 			int64(1),
	// 			attribute.String("goversion", config.GoVersion),
	// 			semconv.ServiceVersionKey.String(config.Version),
	// 			attribute.String("revision", config.Revision),
	// 			attribute.String("branch", config.Branch),
	// 			attribute.String("build_date", config.BuildDate),
	// 			attribute.String("build_user", config.BuildUser),
	// 		)
	// 	},
	// )

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := provider.Shutdown(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to shutdown metric provider")
		}
	}, nil
}
