package main

import (
	"context"
	"fmt"
	"net/http"

	"os"
	"os/signal"
	"syscall"
	"time"

	httppb "github.com/10664kls/app-in-performance-api/genproto/go/http/v1"
	"github.com/10664kls/app-in-performance-api/internal/appin"
	"github.com/10664kls/app-in-performance-api/internal/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	stdmw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zlog, err := newLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer zlog.Sync()
	zap.ReplaceGlobals(zlog)
	zlog.Info("Logger replaced in globals")
	zlog.Info("Logger initialized")

	appInSvc, err := appin.NewService(ctx, &appin.Config{
		Zlog:          zlog,
		TenantID:      os.Getenv("TENANT_ID"),
		ClientID:      os.Getenv("CLIENT_ID"),
		Secret:        os.Getenv("CLIENT_SECRET"),
		SiteID:        os.Getenv("SITE_ID"),
		ListID:        os.Getenv("LIST_ID"),
		CAFinalListID: os.Getenv("CA_FINAL_LIST_ID"),
		Scopes:        []string{},
	})
	if err != nil {
		return fmt.Errorf("failed to create appin service: %w", err)
	}
	zlog.Info("AppIn service initialized")

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = httpErr
	e.Use(httpLogger(zlog))
	e.Use(stdMws()...)

	serve := must(server.NewServer(appInSvc))
	if err := serve.Install(e); err != nil {
		return fmt.Errorf("failed to install server: %w", err)
	}

	errCh := make(chan error)
	go func() {
		errCh <- e.Start(fmt.Sprintf(":%s", getEnv("PORT", "8890")))
	}()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		zlog.Info("Received shutdown signal, shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		zlog.Info("Waiting for server to shut down...")
		if err := e.Shutdown(ctx); err != nil {
			zlog.Error("Error shutting down server", zap.Error(err))
			return err
		}
		zlog.Info("Server shut down gracefully")

	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			zlog.Error("Error starting server", zap.Error(err))
			return err
		}
	}

	return nil
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func httpLogger(zlog *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id != "" {
				fields = append(fields, zap.String("request_id", id))
			}

			n := res.Status
			switch {
			case n >= 500:
				zlog.
					With(zap.Error(err)).
					Error("HTTP Error", fields...)

			case n >= 400:
				zlog.
					With(zap.Error(err)).
					Warn("HTTP Error", fields...)

			case n >= 300:
				zlog.
					Info("Redirect", fields...)

			default:
				zlog.
					Info("HTTP Request", fields...)
			}

			return nil
		}
	}
}

func stdMws() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		stdmw.RemoveTrailingSlash(),
		stdmw.Recover(),
		stdmw.CORSWithConfig((stdmw.CORSConfig{
			AllowOriginFunc: func(origin string) (bool, error) {
				return true, nil
			},
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodOptions,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowCredentials: true,
			MaxAge:           3600,
		})),
		stdmw.Secure(),
		stdmw.RateLimiter(stdmw.NewRateLimiterMemoryStore(30)),
	}
}

func httpErr(err error, c echo.Context) {
	if s, ok := status.FromError(err); ok {
		he := httpStatusPbFromRPC(s)
		jsonb, _ := protojson.Marshal(he)
		c.JSONBlob(int(he.Error.Code), jsonb)
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		var s *status.Status
		switch he.Code {
		case http.StatusNotFound, http.StatusMethodNotAllowed:
			s = status.New(codes.NotFound, "Not found!")

		case http.StatusTooManyRequests:
			s = status.New(codes.ResourceExhausted, "Too many requests!")

		case http.StatusInternalServerError:
			s = status.New(codes.Internal, "An internal server error occurred!")

		default:
			s = status.New(codes.Unknown, "An unknown error occurred!")
		}

		hpb := httpStatusPbFromRPC(s)
		jsonb, _ := protojson.Marshal(hpb)
		c.JSONBlob(int(hpb.Error.Code), jsonb)
		return
	}

	hpb := httpStatusPbFromRPC(status.New(codes.Internal, "An internal server error occurred!"))
	jsonb, _ := protojson.Marshal(hpb)
	c.JSONBlob(int(hpb.Error.Code), jsonb)
}

func httpStatusPbFromRPC(s *status.Status) *httppb.Error {
	return &httppb.Error{
		Error: &httppb.Status{
			Code:    int32(runtime.HTTPStatusFromCode(s.Code())),
			Message: s.Message(),
			Status:  code.Code(s.Code()),
			Details: s.Proto().GetDetails(),
		},
	}
}

func newLogger() (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("02/01/2006 15:04:05 Z07:00"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zlog, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return zlog, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
