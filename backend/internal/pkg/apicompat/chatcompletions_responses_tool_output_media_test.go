package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponsesToolOutputMediaBecomesMultimodalChatMessage(t *testing.T) {
	input := json.RawMessage(`[
		{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_image","output":[
			{"type":"input_text","text":"render complete"},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,AQID"}}
		]}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 3)
	require.Equal(t, "assistant", messages[0].Role)
	require.Equal(t, "tool", messages[1].Role)
	require.Equal(t, "user", messages[2].Role)
	require.Contains(t, string(messages[1].Content), toolOutputMediaMarker)
	require.NotContains(t, string(messages[1].Content), "data:image/")

	var parts []ChatContentPart
	require.NoError(t, json.Unmarshal(messages[2].Content, &parts))
	require.Len(t, parts, 2)
	require.Equal(t, "[Tool output media for call call_image]", parts[0].Text)
	require.Equal(t, "image_url", parts[1].Type)
	require.NotNil(t, parts[1].ImageURL)
	require.Equal(t, "data:image/png;base64,AQID", parts[1].ImageURL.URL)
}

func TestResponsesToolOutputMediaSupportsBareDataURL(t *testing.T) {
	input := json.RawMessage(`[
		{"type":"function_call","call_id":"call_image","name":"view_image","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_image","output":"data:image/jpeg;base64,BAUG"}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 3)

	var parts []ChatContentPart
	require.NoError(t, json.Unmarshal(messages[2].Content, &parts))
	require.Len(t, parts, 2)
	require.NotNil(t, parts[1].ImageURL)
	require.Equal(t, "data:image/jpeg;base64,BAUG", parts[1].ImageURL.URL)
}

func TestResponsesToolOutputMediaPreservesTextOnlyOutput(t *testing.T) {
	input := json.RawMessage(`[
		{"type":"function_call","call_id":"call_text","name":"exec","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_text","output":{"ok":true,"text":"complete"}}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 2)
	var toolText string
	require.NoError(t, json.Unmarshal(messages[1].Content, &toolText))
	require.Equal(t, `{"ok":true,"text":"complete"}`, toolText)
}
