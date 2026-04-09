package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareOpsResponseBodyForQueue_TruncatesAndPreservesOriginalSize(t *testing.T) {
	raw := []byte(strings.Repeat("x", opsMaxStoredResponseBodyBytes+128))

	body, truncated, bytesLen := PrepareOpsResponseBodyForQueue(raw)

	require.NotNil(t, body)
	require.True(t, truncated)
	require.NotNil(t, bytesLen)
	require.Equal(t, len(raw), *bytesLen)
	require.Len(t, *body, opsMaxStoredResponseBodyBytes)
}

func TestPrepareOpsResponseBodyForQueue_EmptyAfterTrim(t *testing.T) {
	body, truncated, bytesLen := PrepareOpsResponseBodyForQueue([]byte("   \n\t  "))

	require.Nil(t, body)
	require.False(t, truncated)
	require.NotNil(t, bytesLen)
	require.Equal(t, 7, *bytesLen)
}
