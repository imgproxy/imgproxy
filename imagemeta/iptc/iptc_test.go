package iptc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/imagemeta/iptc"
)

func TestParse(t *testing.T) {
	key := iptc.TagKey{RecordID: 2, TagID: 80}

	tests := []struct {
		name      string
		in        string
		want      []string
		wantError bool
	}{
		{
			name: "short dataset",
			in:   "\x1c\x02\x50\x00\x03abc",
			want: []string{"abc"},
		},
		{
			name: "four-byte size followed by another dataset",
			in:   "\x1c\x02\x50\x80\x04\x00\x00\x00\x03abc\x1c\x02\x50\x00\x03def",
			want: []string{"abc", "def"},
		},
		{
			name: "eight-byte size followed by another dataset",
			in:   "\x1c\x02\x50\x80\x08\x00\x00\x00\x00\x00\x00\x00\x03abc\x1c\x02\x50\x00\x03def",
			want: []string{"abc", "def"},
		},
		{
			name:      "oversized short size",
			in:        "\x1c\x02\x50\x00\x0aabc",
			wantError: true,
		},
		{
			name:      "oversized four-byte size",
			in:        "\x1c\x02\x50\x80\x04\xff\xff\xff\xff\x00\x00\x00\x00",
			wantError: true,
		},
		{
			name:      "oversized eight-byte size",
			in:        "\x1c\x02\x50\x80\x08\x80\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
			wantError: true,
		},
		{
			name:      "truncated four-byte size",
			in:        "\x1c\x02\x50\x80\x04\x00\x00\x00",
			wantError: true,
		},
		{
			name:      "truncated eight-byte size",
			in:        "\x1c\x02\x50\x80\x08\x00\x00\x00\x00\x00\x00\x00",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := iptc.IptcMap{}
			err := iptc.Parse([]byte(tt.in), m)
			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, m[key], len(tt.want))
			for i, want := range tt.want {
				require.Equal(t, []byte(want), m[key][i].Raw)
			}
		})
	}
}
