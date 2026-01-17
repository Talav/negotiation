package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMedia_Parameters(t *testing.T) {
	acc, err := newMedia("foo/bar; q=1; hello=world")
	require.NoError(t, err)

	// Test existing parameter
	assert.Equal(t, "world", acc.Parameters["hello"])

	// Test non-existing parameter (should return zero value)
	assert.Equal(t, "", acc.Parameters["unknown"])
}

func TestNewMedia_NormalizedValue(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "sorted parameters",
			header:   "text/html; z=y; a=b; c=d",
			expected: "text/html; a=b; c=d; z=y",
		},
		{
			name:     "with quality",
			header:   "application/pdf; q=1; param=p",
			expected: "application/pdf; param=p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.NormalizedValue)
		})
	}
}

func TestNewMedia_Type(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with parameters", "text/html;hello=world", "text/html"},
		{"simple", "application/pdf", "application/pdf"},
		{"with quality", "application/xhtml+xml;q=0.9", "application/xhtml+xml"},
		{"with quality and space", "text/plain; q=0.5", "text/plain"},
		{"with level", "text/html;level=2;q=0.4", "text/html"},
		{"with spaces", "text/html ; level = 2   ; q = 0.4", "text/html"},
		{"wildcard subtype", "text/*", "text/*"},
		{"wildcard with params", "text/* ;q=1 ;level=2", "text/*"},
		{"full wildcard", "*/*", "*/*"},
		{"single wildcard", "*", "*/*"},
		{"wildcard with params", "*/* ; param=555", "*/*"},
		{"single wildcard with params", "* ; param=555", "*/*"},
		{"case insensitive", "TEXT/hTmL;leVel=2; Q=0.4", "text/html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Type)
		})
	}
}

func TestNewMedia_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no slash", "text"},
		{"empty type", "/html"},
		{"empty subtype", "text/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newMedia(tt.header)
			assert.Error(t, err)
			assert.IsType(t, &InvalidMediaTypeError{}, err)
		})
	}
}

func TestNewMedia_Value(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with spaces", "text/html;hello=world  ;q=0.5", "text/html;hello=world  ;q=0.5"},
		{"simple", "application/pdf", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newMedia(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Value)
		})
	}
}

func TestHeader_ParametersMap(t *testing.T) {
	acc, err := newMedia("text/html; charset=UTF-8; level=2")
	require.NoError(t, err)

	params := acc.Parameters
	assert.Equal(t, "UTF-8", params["charset"])
	assert.Equal(t, "2", params["level"])

	// Modify returned map - should not affect original (need to copy)
	paramsCopy := make(map[string]string)
	for k, v := range acc.Parameters {
		paramsCopy[k] = v
	}
	paramsCopy["charset"] = "ISO-8859-1"
	assert.Equal(t, "UTF-8", acc.Parameters["charset"])
}

func TestNewLanguage_Type(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		expectedType string
		expectedBase string
		expectedSub  string
	}{
		{"simple language", "en", "en", "en", ""},
		{"language with region", "en-US", "en-us", "en", "us"},
		{"language with script and region", "zh-Hans-CN", "zh-hans-cn", "zh", "cn"},
		{"case insensitive", "EN-us", "en-us", "en", "us"},
		{"with parameters", "en;q=0.8", "en", "en", ""},
		{"with region and parameters", "fr-CA;q=0.9", "fr-ca", "fr", "ca"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newLanguage(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, acc.Type)
			assert.Equal(t, tt.expectedBase, acc.BasePart)
			assert.Equal(t, tt.expectedSub, acc.SubPart)
		})
	}
}

func TestNewLanguage_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"too many parts", "en-US-CA-GB"},
		{"four parts", "zh-Hans-CN-TW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newLanguage(tt.header)
			assert.Error(t, err)
			assert.IsType(t, &InvalidLanguageError{}, err)
		})
	}
}

func TestNewCharset_Type(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		expectedType string
	}{
		{"simple charset", "utf-8", "utf-8"},
		{"uppercase", "UTF-8", "utf-8"},
		{"with parameters", "iso-8859-1;q=0.9", "iso-8859-1"},
		{"case insensitive", "UTF-8", "utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newCharset(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, acc.Type)
			assert.Equal(t, "", acc.BasePart)
			assert.Equal(t, "", acc.SubPart)
		})
	}
}

func TestNewEncoding_Type(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		expectedType string
	}{
		{"gzip", "gzip", "gzip"},
		{"deflate", "deflate", "deflate"},
		{"identity", "identity", "identity"},
		{"with parameters", "gzip;q=0.8", "gzip"},
		{"case insensitive", "GZIP", "gzip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newEncoding(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, acc.Type)
			assert.Equal(t, "", acc.BasePart)
			assert.Equal(t, "", acc.SubPart)
		})
	}
}

func TestNewCharset_Value(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with spaces", "utf-8 ; q=0.9", "utf-8 ; q=0.9"},
		{"simple", "iso-8859-1", "iso-8859-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newCharset(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Value)
		})
	}
}

func TestNewEncoding_Value(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"with spaces", "gzip ; q=1.0", "gzip ; q=1.0"},
		{"simple", "deflate", "deflate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := newEncoding(tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.Value)
		})
	}
}
