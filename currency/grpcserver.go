package currency

import (
	"context"
	"golang.org/x/sync/semaphore"
	"math/rand"
	currencypb "something/pb"
	"time"
)

var _ currencypb.CurrencyServer = (*GRPCServer)(nil)

func init() {
	rand.Seed(time.Now().Unix())
}

// GRPCServer ...
type GRPCServer struct{}

var mtx = semaphore.NewWeighted(2)

func (G GRPCServer) DoStrm(request *currencypb.CurRequest, stream currencypb.Currency_DoStrmServer) error {
	if err := mtx.Acquire(stream.Context(), 1); err != nil {
		return err
	}
	defer mtx.Release(1)

	for stream.Context().Err() == nil {
		var currencies [10]currencypb.Currencies
		for _, currency := range currencies {
			country := currencypb.Countries(rand.Intn(5))
			switch country {
			case currencypb.Countries_USA:
				currency = currencypb.Currencies_Dollar
			case currencypb.Countries_Ukraine:
				currency = currencypb.Currencies_Hryvnia
			case currencypb.Countries_Spain:
				currency = currencypb.Currencies_Euro
			case currencypb.Countries_Japan:
				currency = currencypb.Currencies_Yen
			case currencypb.Countries_Belarus:
				currency = currencypb.Currencies_BelRuble
			}
			curRes := &currencypb.CurResponse{
				Currency: currency,
				Value:    rand.Int31(),
			}
			if err := stream.Send(curRes); err != nil {
				return err
			}
		}
		select {
		case <-time.After(1 * time.Second):
		case <-stream.Context().Done():

		}
	}
	println("stop")
	return nil
}

func (G GRPCServer) Do(ctx context.Context, request *currencypb.CurRequest) (*currencypb.CurResponse, error) {
	var country currencypb.Countries
	country = request.GetMessage()
	var currency currencypb.Currencies
	switch country {
	case currencypb.Countries_USA:
		currency = currencypb.Currencies_Dollar
	case currencypb.Countries_Ukraine:
		currency = currencypb.Currencies_Hryvnia
	case currencypb.Countries_Spain:
		currency = currencypb.Currencies_Euro
	case currencypb.Countries_Japan:
		currency = currencypb.Currencies_Yen
	case currencypb.Countries_Belarus:
		currency = currencypb.Currencies_BelRuble
	}

	var value int32
	value = rand.Int31()

	return &currencypb.CurResponse{Currency: currency, Value: value}, nil
}
