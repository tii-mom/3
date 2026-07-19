package creditctx

import "context"

// Metadata describes why a compatibility balance update is being applied.
// Callers that do not provide metadata are treated as non-transferable grants.
type Metadata struct {
	EntryType      string
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Transferable   bool
	CountRecharge  bool
	// DebitTransferableFirst is reserved for principal reversals. Normal API
	// usage continues to consume promotional credit before transferable funds.
	DebitTransferableFirst bool
	Attributes             map[string]any
}

type contextKey struct{}

func WithMetadata(ctx context.Context, metadata Metadata) context.Context {
	return context.WithValue(ctx, contextKey{}, metadata)
}

func FromContext(ctx context.Context) Metadata {
	metadata, _ := ctx.Value(contextKey{}).(Metadata)
	if metadata.EntryType == "" {
		metadata.EntryType = "balance_adjustment"
	}
	if metadata.SourceType == "" {
		metadata.SourceType = "legacy_balance_update"
	}
	return metadata
}
