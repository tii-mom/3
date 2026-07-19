package handler

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"go.uber.org/zap"
)

// logRequestBodyParseFailure records the real reason a request body failed
// JSON parsing/validation. The client keeps receiving the generic
// "Failed to parse request body"; the sanitized diagnostics (underlying
// error with byte offset, body length, and a non-reversible digest) land in
// the server log only, so operators can distinguish genuinely invalid JSON
// from a truncated or partially consumed body.
//
// err may be nil for call sites that validate with gjson.ValidBytes directly;
// the diagnostic error is derived from the body in that case.
func logRequestBodyParseFailure(reqLog *zap.Logger, body []byte, err error) {
	if reqLog == nil {
		return
	}
	if err == nil {
		err = service.DescribeInvalidJSON(body)
	}

	digest := sha256.Sum256(body)
	fields := []zap.Field{
		zap.Error(err),
		zap.Int("body_len", len(body)),
		zap.String("body_sha256", hex.EncodeToString(digest[:])),
	}
	reqLog.Warn("parse request body failed", fields...)
}
