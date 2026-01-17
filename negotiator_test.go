package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNegotiator_Negotiate_MediaType(t *testing.T) {
	negotiator := NewMediaNegotiator()

	tests := []struct {
		name           string
		acceptHeader   string
		priorities     []string
		strict         bool
		expectedType   string
		expectedParams map[string]string
		expectError    bool
	}{
		{
			name:         "simple match",
			acceptHeader: "text/html, application/json;q=0.9",
			priorities:   []string{"application/json", "text/html"},
			strict:       false,
			expectedType: "text/html",
		},
		{
			name:         "quality preference",
			acceptHeader: "text/html;q=0.5, application/json;q=0.9",
			priorities:   []string{"application/json", "text/html"},
			strict:       false,
			expectedType: "application/json",
		},
		{
			name:           "exact match with parameters",
			acceptHeader:   "text/html;level=1",
			priorities:     []string{"text/html;level=1"},
			strict:         false,
			expectedType:   "text/html",
			expectedParams: map[string]string{"level": "1"},
		},
		{
			name:         "no match",
			acceptHeader: "text/html",
			priorities:   []string{"application/json"},
			strict:       false,
			expectError:  true,
		},
		{
			name:         "empty priorities",
			acceptHeader: "text/html",
			priorities:   []string{},
			strict:       false,
			expectError:  true,
		},
		{
			name:         "empty header",
			acceptHeader: "",
			priorities:   []string{"text/html"},
			strict:       false,
			expectError:  true,
		},
		{
			name:         "invalid header strict mode",
			acceptHeader: "invalid/header/format",
			priorities:   []string{"text/html"},
			strict:       true,
			expectError:  true,
		},
		{
			name:         "invalid header non-strict mode",
			acceptHeader: "invalid/header/format, text/html",
			priorities:   []string{"text/html"},
			strict:       false,
			expectedType: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := negotiator.Negotiate(tt.acceptHeader, tt.priorities, tt.strict)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedType, result.Type)
				if tt.expectedParams != nil {
					for k, v := range tt.expectedParams {
						assert.Equal(t, v, result.Parameters[k])
					}
				}
			}
		})
	}
}

func TestNegotiator_Negotiate_Language(t *testing.T) {
	negotiator := NewLanguageNegotiator()

	tests := []struct {
		name         string
		acceptHeader string
		priorities   []string
		expectedType string
		expectedBase string
		expectedSub  string
	}{
		{
			name:         "simple language match",
			acceptHeader: "en, fr;q=0.8",
			priorities:   []string{"fr", "en"},
			expectedType: "en",
			expectedBase: "en",
			expectedSub:  "",
		},
		{
			name:         "language with region",
			acceptHeader: "en-US, fr-FR;q=0.9",
			priorities:   []string{"fr-FR", "en-US"},
			expectedType: "en-us",
			expectedBase: "en",
			expectedSub:  "us",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := negotiator.Negotiate(tt.acceptHeader, tt.priorities, false)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedBase, result.BasePart)
			assert.Equal(t, tt.expectedSub, result.SubPart)
		})
	}
}

func TestNegotiator_Negotiate_Charset(t *testing.T) {
	negotiator := NewCharsetNegotiator()

	result, err := negotiator.Negotiate("utf-8, iso-8859-1;q=0.9", []string{"iso-8859-1", "utf-8"}, false)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "utf-8", result.Type)
	assert.Equal(t, "", result.BasePart)
	assert.Equal(t, "", result.SubPart)
}

func TestNegotiator_Negotiate_Encoding(t *testing.T) {
	negotiator := NewEncodingNegotiator()

	result, err := negotiator.Negotiate("gzip, deflate;q=0.8", []string{"deflate", "gzip"}, false)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "gzip", result.Type)
	assert.Equal(t, "", result.BasePart)
	assert.Equal(t, "", result.SubPart)
}

func TestNegotiator_GetOrderedElements(t *testing.T) {
	negotiator := NewMediaNegotiator()

	tests := []struct {
		name          string
		header        string
		expectedLen   int
		expectedOrder []string
		expectError   bool
	}{
		{
			name:          "quality ordering",
			header:        "text/html;q=0.3, application/json;q=0.9, text/plain",
			expectedLen:   3,
			expectedOrder: []string{"text/plain", "application/json", "text/html"},
		},
		{
			name:          "same quality preserves original order",
			header:        "text/html, application/json, text/plain",
			expectedLen:   3,
			expectedOrder: []string{"text/html", "application/json", "text/plain"},
		},
		{
			name:        "empty header",
			header:      "",
			expectError: true,
		},
		{
			name:        "malformed header",
			header:      `text/html;q="unclosed, application/json`,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := negotiator.GetOrderedElements(tt.header)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, elements)
			} else {
				require.NoError(t, err)
				assert.Len(t, elements, tt.expectedLen)
				if tt.expectedOrder != nil {
					for i, expected := range tt.expectedOrder {
						assert.Equal(t, expected, elements[i].Type)
					}
				}
			}
		})
	}
}

func TestNegotiator_GetOrderedElements_Language(t *testing.T) {
	negotiator := NewLanguageNegotiator()

	elements, err := negotiator.GetOrderedElements("fr;q=0.8, en, de;q=0.5")
	require.NoError(t, err)
	require.Len(t, elements, 3)

	// Should be ordered by quality: en (1.0), fr (0.8), de (0.5)
	assert.Equal(t, "en", elements[0].Type)
	assert.Equal(t, "fr", elements[1].Type)
	assert.Equal(t, "de", elements[2].Type)
}

func TestNegotiator_InvalidPriorities(t *testing.T) {
	negotiator := NewMediaNegotiator()

	// Test invalid priority in strict mode
	_, err := negotiator.Negotiate("text/html", []string{"invalid/priority/format"}, true)
	assert.Error(t, err)

	// Test invalid priority in non-strict mode (should be skipped)
	result, err := negotiator.Negotiate("text/html", []string{"invalid/priority/format", "text/html"}, false)
	require.NoError(t, err)
	assert.Equal(t, "text/html", result.Type)
}

func TestNegotiator_WildcardMatching(t *testing.T) {
	negotiator := NewMediaNegotiator()

	// Test */* matches anything
	result, err := negotiator.Negotiate("*/*", []string{"application/json"}, false)
	require.NoError(t, err)
	assert.Equal(t, "application/json", result.Type)

	// Test type/* matching
	result, err = negotiator.Negotiate("text/*", []string{"text/html", "application/json"}, false)
	require.NoError(t, err)
	assert.Equal(t, "text/html", result.Type)
}
