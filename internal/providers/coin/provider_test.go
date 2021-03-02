package coin_provider

import (
	"fmt"
	"testing"

	"github.com/aspiration-labs/pyggpot/internal/models"

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
			name: "PP.P",
			err:  github_com_mwitkow_go_proto_validators.FieldError("PotName", fmt.Errorf(errMsg)),
		},
		{
			name: "PP-",
			err:  github_com_mwitkow_go_proto_validators.FieldError("PotName", fmt.Errorf(errMsg)),
		},
		{
			name: "PP-P",
			err:  nil,
		},
		{
			name: "PPP",
			err:  nil,
		},
		{
			name: "PP P",
			err:  nil,
		},
		{
			name: "PP_P",
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

func TestShakePot(t *testing.T) {
	type testCase struct {
		name  string
		pot   []*models.CoinsInPot
		draws []int
		want  []int
	}

	tests := []testCase{
		{
			name:  "zero pot",
			pot:   []*models.CoinsInPot{},
			draws: []int{1, 2, 3},
			want:  []int{},
		},
		{
			name: "empty pot",
			pot: []*models.CoinsInPot{
				{
					ID:           0,
					Denomination: 1,
					CoinCount:    0,
				},
				{
					ID:           1,
					Denomination: 5,
					CoinCount:    0,
				},
				{
					ID:           2,
					Denomination: 10,
					CoinCount:    0,
				},
			},
			draws: []int{1, 2, 3},
			want:  []int{0, 0, 0},
		},
		{
			name: "draw more",
			pot: []*models.CoinsInPot{
				{
					ID:           0,
					Denomination: 1,
					CoinCount:    10,
				},
				{
					ID:           1,
					Denomination: 5,
					CoinCount:    5,
				},
				{
					ID:           2,
					Denomination: 10,
					CoinCount:    1,
				},
			},
			draws: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18},
			want:  []int{0, 0, 0},
		},
		{
			name: "draw front",
			pot: []*models.CoinsInPot{
				{
					ID:           0,
					Denomination: 1,
					CoinCount:    3,
				},
				{
					ID:           1,
					Denomination: 5,
					CoinCount:    2,
				},
				{
					ID:           2,
					Denomination: 10,
					CoinCount:    2,
				},
			},
			draws: []int{2, 1, 0},
			want:  []int{0, 2, 2},
		},
		{
			name: "draw middle",
			pot: []*models.CoinsInPot{
				{
					ID:           0,
					Denomination: 1,
					CoinCount:    3,
				},
				{
					ID:           1,
					Denomination: 5,
					CoinCount:    2,
				},
				{
					ID:           2,
					Denomination: 10,
					CoinCount:    2,
				},
			},
			draws: []int{3, 3},
			want:  []int{3, 0, 2},
		},
		{
			name: "draw end",
			pot: []*models.CoinsInPot{
				{
					ID:           0,
					Denomination: 1,
					CoinCount:    3,
				},
				{
					ID:           1,
					Denomination: 5,
					CoinCount:    2,
				},
				{
					ID:           2,
					Denomination: 10,
					CoinCount:    2,
				},
			},
			draws: []int{7, 6},
			want:  []int{3, 2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := 0
			draw := func() int {
				r := tt.draws[i]
				i++
				return r
			}

			actual := shakePot(tt.pot, len(tt.draws), draw)
			assert.Equal(t, tt.want, actual)
		})
	}
}
