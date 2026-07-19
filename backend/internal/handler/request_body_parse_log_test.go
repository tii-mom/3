//go:build unit

package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zap.WarnLevel)
	return zap.New(core), logs
}

func loggedFields(t *testing.T, logs *observer.ObservedLogs) map[string]any {
	t.Helper()
	entries := logs.All()
	require.Len(t, entries, 1)
	fields := map[string]any{}
	for _, f := range entries[0].Context {
		switch f.Key {
		case "body_len":
			fields[f.Key] = int(f.Integer)
		case "error":
			fields[f.Key] = f.Interface.(error).Error()
		default:
			fields[f.Key] = f.String
		}
	}
	return fields
}

func TestLogRequestBodyParseFailure_DerivesErrorWhenNil(t *testing.T) {
	log, logs := newObservedLogger(t)
	body := []byte(`{"model": bad}`)

	logRequestBodyParseFailure(log, body, nil)

	fields := loggedFields(t, logs)
	require.Equal(t, len(body), fields["body_len"])
	require.Contains(t, fields["error"], "invalid json")
	require.Contains(t, fields["error"], "offset=11")
}

func TestLogRequestBodyParseFailure_LogsDigestWithoutContent(t *testing.T) {
	log, logs := newObservedLogger(t)
	body := []byte(`{"broken":`)

	logRequestBodyParseFailure(log, body, nil)

	fields := loggedFields(t, logs)
	digest := sha256.Sum256(body)
	require.Equal(t, hex.EncodeToString(digest[:]), fields["body_sha256"])
	require.NotContains(t, fields, "body_head")
	require.NotContains(t, fields, "body_tail")
}

func TestLogRequestBodyParseFailure_DoesNotLogBodyContent(t *testing.T) {
	log, logs := newObservedLogger(t)
	body := []byte(`{"model":"claude-sonnet-4-6","secret":"do-not-log"`)

	logRequestBodyParseFailure(log, body, nil)

	fields := loggedFields(t, logs)
	require.Equal(t, len(body), fields["body_len"])
	require.NotContains(t, fields, "body_head")
	require.NotContains(t, fields, "body_tail")
	require.Len(t, fields["body_sha256"], 64)
}

func TestLogRequestBodyParseFailure_NilLoggerNoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		logRequestBodyParseFailure(nil, []byte(`{`), nil)
	})
}
