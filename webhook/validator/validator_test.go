package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSignature(t *testing.T) {
	body := []byte(`{"event":"test"}`)
	v := New("secret", nil)

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	assert.True(t, v.ValidateSignature(signature, body))
	assert.False(t, v.ValidateSignature("bad-signature", body))
	assert.False(t, New("", nil).ValidateSignature(signature, body))
}

func TestValidateIP(t *testing.T) {
	_, allowedNet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	v := New("secret", []*net.IPNet{allowedNet})

	assert.True(t, v.ValidateIP("10.1.2.3:1234"))
	assert.False(t, v.ValidateIP("192.168.1.1:1234"))
	assert.False(t, v.ValidateIP("not-an-ip"))
	assert.True(t, New("secret", nil).ValidateIP("anything"))
}

func TestExtractIP(t *testing.T) {
	assert.Equal(t, "127.0.0.1", extractIP("127.0.0.1:8080"))
	assert.Equal(t, "2001:db8::1", extractIP("[2001:db8::1]:8080"))
	assert.Equal(t, "not-an-addr", extractIP("not-an-addr"))
}
