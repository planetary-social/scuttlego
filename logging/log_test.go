package logging_test

import (
	"io"
	"testing"

	"github.com/planetary-social/scuttlego/logging"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

func BenchmarkContextLogger(b *testing.B) {
	benchCases := []struct {
		Name             string
		GetLoggingSystem func(tb testing.TB) logging.LoggingSystem
	}{
		{
			Name: "dev_null",
			GetLoggingSystem: func(tb testing.TB) logging.LoggingSystem {
				return logging.NewDevNullLoggingSystem()
			},
		},
		{
			Name: "logrus",
			GetLoggingSystem: func(tb testing.TB) logging.LoggingSystem {
				logrusLogger := logrus.New()
				logrusLogger.SetLevel(logrus.TraceLevel)
				logrusLogger.SetOutput(io.Discard)
				return logging.NewLogrusLoggingSystem(logrusLogger)
			},
		},
		{
			Name: "logrus_disabled",
			GetLoggingSystem: func(tb testing.TB) logging.LoggingSystem {
				logrusLogger := logrus.New()
				logrusLogger.SetLevel(logrus.ErrorLevel)
				logrusLogger.SetOutput(io.Discard)
				return logging.NewLogrusLoggingSystem(logrusLogger)
			},
		},
		{
			Name: "zerolog",
			GetLoggingSystem: func(tb testing.TB) logging.LoggingSystem {
				zerologLogger := zerolog.New(io.Discard).Level(zerolog.TraceLevel)
				return logging.NewZerologLoggingSystem(zerologLogger)
			},
		},
		{
			Name: "zerolog_disabled",
			GetLoggingSystem: func(tb testing.TB) logging.LoggingSystem {
				zerologLogger := zerolog.New(io.Discard).Level(zerolog.Disabled)
				return logging.NewZerologLoggingSystem(zerologLogger)
			},
		},
	}

	for _, benchCase := range benchCases {
		b.Run(benchCase.Name, func(b *testing.B) {
			logger := logging.NewContextLogger(benchCase.GetLoggingSystem(b), "bench")

			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger.Trace().
					WithField("field1", 123).
					WithField("field2", "somestring").
					WithField("field3", struct{ string }{"object"}).
					Message("some message")
			}
		})
	}
}
