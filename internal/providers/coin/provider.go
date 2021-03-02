package coin_provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/aspiration-labs/pyggpot/internal/models"
	coin_service "github.com/aspiration-labs/pyggpot/rpc/go/coin"
	"github.com/twitchtv/twirp"
)

type coinServer struct {
	DB *sql.DB
}

func New(db *sql.DB) *coinServer {
	return &coinServer{
		DB: db,
	}
}

func (s *coinServer) AddCoins(ctx context.Context, request *coin_service.AddCoinsRequest) (*coin_service.CoinsListResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError(err.Error(), "")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}
	for _, coin := range request.Coins {
		fmt.Println(coin)
		newCoin := models.Coin{
			PotID:        request.PotId,
			Denomination: int32(coin.Kind),
			CoinCount:    coin.Count,
		}
		err = newCoin.Save(tx)
		if err != nil {
			return nil, twirp.InvalidArgumentError(err.Error(), "")
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	return &coin_service.CoinsListResponse{
		Coins: request.Coins,
	}, nil
}

func (s *coinServer) RemoveCoins(ctx context.Context, request *coin_service.RemoveCoinsRequest) (*coin_service.CoinsListResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError(err.Error(), "")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	// CoinsInPotsByPot_id
	pot, err := models.CoinsInPotsByPot_id(tx, int(request.PotId))
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	coins := make([]*models.Coin, len(pot))
	for i, c := range pot {
		coins[i], err = models.CoinByID(tx, c.ID)
		if err != nil {
			return nil, twirp.InternalError(err.Error())
		}
	}

	// shake pot
	final := shakePot(pot, int(request.Count), rand.Int)
	defer func() {
		_ = tx.Rollback()
	}()

	resp := coin_service.CoinsListResponse{
		Coins: make([]*coin_service.Coins, len(pot)),
	}

	for i, c := range coins {
		if c.CoinCount == int32(final[i]) {
			// skip update when no changes
			continue
		}

		c.CoinCount = int32(final[i])
		if err = c.Update(tx); err != nil {
			return nil, twirp.InternalError(err.Error())
		}

		resp.Coins[i] = &coin_service.Coins{
			Kind:  coin_service.Coins_Kind(c.Denomination),
			Count: int32(final[i]),
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	return &resp, nil

}

// shakePot simulate a person pickup the pyggot (lock the resource exclusively) and then shake the pot to remove coin
// the person will let go the pot (unlock the resource) after getting the count or all coins out of the pot
func shakePot(pot []*models.CoinsInPot, count int, randInt func() int) []int {
	total := 0
	tally := make([]int, len(pot))
	for i, c := range pot {
		x := int(c.CoinCount)
		tally[i] = x
		total += x
	}

	if total == 0 || count >= total {
		return make([]int, len(pot))
	}

	for i := count; i > 0; i-- {
		n := randInt()%total + 1
		total--
		m := 0
		for j, c := range tally {
			m += c
			if n <= m {
				tally[j]--
				break
			}
		}
	}

	return tally
}
