package messages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBlobsGetArgumentsFromBytes(t *testing.T) {
	args, err := NewBlobsGetArgumentsFromBytes([]byte(`["&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"]`))
	require.NoError(t, err)

	require.Equal(t, "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", args.Id().String())
}
