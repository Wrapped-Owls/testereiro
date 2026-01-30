package strnormalizer

import "testing"

func BenchmarkToSnakeCase(b *testing.B) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"HTTPServer", "http_server"},
		{"URLParser", "url_parser"},
		{"snake_case", "snake_case"},
		{"Already_Snake", "already_snake"},
		{"with123Numbers", "with123_numbers"},
		{"JSONToXML", "json_to_xml"},
		{"XMLToJSON", "xml_to_json"},
		{"_leadingUnderscore", "leading_underscore"},
		{"TrailingUnderscore_", "trailing_underscore"},
		{"", ""},
	}

	for _, tt := range tests {
		b.Run(
			tt.input, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					actual := ToSnakeCase(tt.input)
					if actual != tt.expected {
						b.Errorf("ToSnakeCase(%q) = %q; want %q", tt.input, actual, tt.expected)
					}
				}
			},
		)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"HTTPServer", "http_server"},
		{"URLParser", "url_parser"},
		{"XMLParser", "xml_parser"},
		{"snake_case", "snake_case"},
		{"Already_Snake", "already_snake"},
		{"with123Numbers", "with123_numbers"},
		{"JSONToXML", "json_to_xml"},
		{"XMLToJSON", "xml_to_json"},
		{"_leadingUnderscore", "leading_underscore"},
		{"TrailingUnderscore_", "trailing_underscore"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.input, func(t *testing.T) {
				actual := ToSnakeCase(tt.input)
				if actual != tt.expected {
					t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, actual, tt.expected)
				}
			},
		)
	}
}

func BenchmarkToCamelCase(b *testing.B) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "snakeCase"},
		{"http_server", "httpServer"},
		{"json_to_xml", "jsonToXml"},
		{"alreadyCamel", "alreadyCamel"},
		{"XMLParser", "xmlParser"},
		{"snake_Case_with_Mixed", "snakeCaseWithMixed"},
		{"_leading_underscore", "leadingUnderscore"},
		{"trailing_underscore_", "trailingUnderscore"},
		{"", ""},
		{"___", ""},
	}

	for _, tt := range tests {
		b.Run(
			tt.input, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					actual := ToCamelCase(tt.input)
					if actual != tt.expected {
						b.Errorf("toCamelCase(%q) = %q; want %q", tt.input, actual, tt.expected)
					}
				}
			},
		)
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "snakeCase"},
		{"http_server", "httpServer"},
		{"json_to_xml", "jsonToXml"},
		{"alreadyCamel", "alreadyCamel"},
		{"XMLParser", "xmlParser"},
		{"snake_Case_with_Mixed", "snakeCaseWithMixed"},
		{"_leading_underscore", "leadingUnderscore"},
		{"trailing_underscore_", "trailingUnderscore"},
		{"", ""},
		{"___", ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.input, func(t *testing.T) {
				actual := ToCamelCase(tt.input)
				if actual != tt.expected {
					t.Errorf("ToCamelCase(%q) = %q; want %q", tt.input, actual, tt.expected)
				}
			},
		)
	}
}
