package coin_provider

import (
	"fmt"
	"testing"

	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	"github.com/stretchr/testify/assert"

	pot_service "github.com/aspiration-labs/pyggpot/rpc/go/pot"
)

func TestFixAddCoinValidationMessage(t *testing.T) {
	type testCase struct {
		name string
		err  error
	}

	errMsg := "Must start with 2 alphanumeric characters, optionally follow by one of the following character: [space], [underscore] or [dash]; and end with ONE or MORE alphanumeric characters"
	tests := []testCase{
		{
			name: "PP",
			err:  github_com_mwitkow_go_proto_validators.FieldError("PotName", fmt.Errorf(errMsg)),
		},
		{
			name: "PP.",
			err:  github_com_mwitkow_go_proto_validators.FieldError("PotName", fmt.Errorf(errMsg)),
		},
		{
			name: "PP-",
			err:  github_com_mwitkow_go_proto_validators.FieldError("PotName", fmt.Errorf(errMsg)),
		},
		{
			name: "PPP",
			err:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &pot_service.CreatePotRequest{
				PotName:  test.name,
				MaxCoins: 6,
			}

			err := req.Validate()
			if test.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.err.Error())
				return
			}

			assert.NoError(t, err)
		})
	}

}
