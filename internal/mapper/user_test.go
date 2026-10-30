package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unser-elefant/remnawave-go/internal/api/model"
)

func TestToDomainUser_DeviceLimit(t *testing.T) {
	zero, seven := int64(0), int64(7)

	tests := []struct {
		name    string
		limit   *int64
		wantSet bool
		want    int64
	}{
		{name: "null", limit: nil, wantSet: false, want: 0},
		{name: "unlimited", limit: &zero, wantSet: true, want: 0},
		{name: "explicit", limit: &seven, wantSet: true, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := ToDomainUser(&model.User{HWIDDeviceLimit: tt.limit})
			assert.Equal(t, tt.wantSet, u.DeviceLimitSet)
			assert.Equal(t, tt.want, u.DeviceLimit)
		})
	}
}
