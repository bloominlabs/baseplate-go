package observability

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"github.com/rs/zerolog/log"
)

func WithDefaultMetricOpts() []controller.Option {
	res, _ := WithDefaultResource(context.Background())
	return []controller.Option{
		res,
	}
}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitMetricsProvider(addr string, credentials *credentials.TransportCredentials, opts ...controller.Option) (func(), error) {
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
	con, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(grpcCreds), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	metricClient := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithGRPCConn(con),
	)
	metricExporter, err := otlpmetric.New(ctx, metricClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create the collector metric exporter: %w", err)
	}

	// Configure the pusher to push metrics CollectPeriod
	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExporter,
		),
		controller.WithExporter(metricExporter),
		controller.WithCollectPeriod(time.Second),
		opts,
	)
	global.SetMeterProvider(pusher)

	if err = runtimemetrics.Start(runtimemetrics.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	}

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

	// start pushing
	err = pusher.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start metric pusher: %w", err)
	}

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		// Shutdown will flush any remaining spans and shut down the exporter.
		if err != nil {
			log.Fatal().Err(pusher.Stop(ctx)).Msg("failed to shutdown MetricsPusher")
		}
	}, nil
}
