// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"

	"github.com/clinia/x/loggerx"
	"github.com/ory/jsonschema/v3"

	"github.com/knadh/koanf"

	"github.com/clinia/x/watcherx"
)

type (
	OptionModifier func(p *Provider)
)

func WithContext(ctx context.Context) OptionModifier {
	return func(p *Provider) {
		for _, o := range ConfigOptionsFromContext(ctx) {
			o(p)
		}
	}
}

func WithConfigFiles(files ...string) OptionModifier {
	return func(p *Provider) {
		p.files = append(p.files, files...)
	}
}

func WithImmutables(immutables ...string) OptionModifier {
	return func(p *Provider) {
		p.immutables = append(p.immutables, immutables...)
	}
}

func WithFlags(flags *pflag.FlagSet) OptionModifier {
	return func(p *Provider) {
		p.flags = flags
	}
}

func WithLogger(l *loggerx.Logger) OptionModifier {
	return func(p *Provider) {
		p.logger = l
	}
}

func SkipValidation() OptionModifier {
	return func(p *Provider) {
		p.skipValidation = true
	}
}

func DisableFileWatching() OptionModifier {
	return func(p *Provider) {
		p.disableFileWatching = true
	}
}

func DisableEnvLoading() OptionModifier {
	return func(p *Provider) {
		p.disableEnvLoading = true
	}
}

func WithValue(key string, value interface{}) OptionModifier {
	return func(p *Provider) {
		p.forcedValues = append(p.forcedValues, tuple{Key: key, Value: value})
	}
}

func WithValues(values map[string]interface{}) OptionModifier {
	return func(p *Provider) {
		for key, value := range values {
			p.forcedValues = append(p.forcedValues, tuple{Key: key, Value: value})
		}
	}
}

func WithBaseValues(values map[string]interface{}) OptionModifier {
	return func(p *Provider) {
		for key, value := range values {
			p.baseValues = append(p.baseValues, tuple{Key: key, Value: value})
		}
	}
}

func WithUserProviders(providers ...koanf.Provider) OptionModifier {
	return func(p *Provider) {
		p.userProviders = providers
	}
}

func OmitKeysFromTracing(keys ...string) OptionModifier {
	return func(p *Provider) {
		p.excludeFieldsFromTracing = keys
	}
}

func AttachWatcher(watcher func(event watcherx.Event, err error)) OptionModifier {
	return func(p *Provider) {
		p.onChanges = append(p.onChanges, watcher)
	}
}

func WithLoggerWatcher(l *loggerx.Logger) OptionModifier {
	return AttachWatcher(LoggerWatcher(l))
}

func LoggerWatcher(l *loggerx.Logger) func(e watcherx.Event, err error) {
	return func(e watcherx.Event, err error) {
		ctx := context.Background()
		l.Info(ctx, "a change to a configuration file was detected", attribute.String("event_type", fmt.Sprintf("%T", e)), attribute.String("file", e.Source()))

		if et := new(jsonschema.ValidationError); errors.As(err, &et) {
			l.
				Error(ctx, "the changed configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting the process", attribute.String("event", fmt.Sprintf("%#v", et)))
		} else if et := new(ImmutableError); errors.As(err, &et) {
			l.WithError(err).
				Error(ctx, "a configuration value marked as immutable has changed. Rolling back to the last working configuration revision. To reload the values please restart the process",
					attribute.String("key", et.Key),
					attribute.String("old_value", fmt.Sprintf("%v", et.From)),
					attribute.String("new_value", fmt.Sprintf("%v", et.To)))
		} else if err != nil {
			l.WithError(err).Error(ctx, "an error occurred while watching config file", attribute.String("file", e.Source()))
		} else {
			l.Info(ctx, "configuration change processed successfully", attribute.String("file", e.Source()), attribute.String("event_type", fmt.Sprintf("%T", e)))
		}
	}
}

func WithStderrValidationReporter() OptionModifier {
	return func(p *Provider) {
		p.onValidationError = func(k *koanf.Koanf, err error) {
			p.printHumanReadableValidationErrors(k, os.Stderr, err)
		}
	}
}

func WithStandardValidationReporter(w io.Writer) OptionModifier {
	return func(p *Provider) {
		p.onValidationError = func(k *koanf.Koanf, err error) {
			p.printHumanReadableValidationErrors(k, w, err)
		}
	}
}
