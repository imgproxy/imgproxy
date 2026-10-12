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
			name:      "short dataset",
			in:        "\x1c\x02\x50\x00\x03abc",
			want:      []string{"abc"},
			wantError: false,
		},
		{
			name:      "four-byte size followed by another dataset",
			in:        "\x1c\x02\x50\x80\x04\x00\x00\x00\x03abc\x1c\x02\x50\x00\x03def",
			want:      []string{"abc", "def"},
			wantError: false,
		},
		{
			name:      "eight-byte size followed by another dataset",
			in:        "\x1c\x02\x50\x80\x08\x00\x00\x00\x00\x00\x00\x00\x03abc\x1c\x02\x50\x00\x03def",
			want:      []string{"abc", "def"},
			wantError: false,
		},
		{
			name:      "oversized four-byte size",
			in:        "\x1c\x02\x50\x80\x04\xff\xff\xff\xff\x00\x00\x00\x00",
			want:      []string{"\x00\x00\x00\x00"},
			wantError: false,
		},
		{
			name:      "oversized eight-byte size",
			in:        "\x1c\x02\x50\x80\x08\x80\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
			want:      []string{"\x00\x00\x00\x00"},
			wantError: false,
		},
		{
			name:      "truncated four-byte size",
			in:        "\x1c\x02\x50\x80\x04\x00\x00\x00",
			want:      nil,
			wantError: true,
		},
		{
			name:      "truncated eight-byte size",
			in:        "\x1c\x02\x50\x80\x08\x00\x00\x00\x00\x00\x00\x00",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := iptc.IptcMap{}
			var err error
			require.NotPanics(t, func() { err = iptc.Parse([]byte(tt.in), m) })
			if tt.wantError {
				require.Error(t, err)
				require.Empty(t, m)
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
