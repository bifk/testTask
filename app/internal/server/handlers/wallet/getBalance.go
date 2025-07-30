package wallet

import (
	"github.com/bifk/testTask/internal/domain/models"
	resp "github.com/bifk/testTask/internal/lib/api/response"
	"github.com/bifk/testTask/internal/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/shopspring/decimal"
	"net/http"
)

type BalanceGetter interface {
	GetWallet(address string) (models.Wallet, error)
}

type BalanceResponse struct {
	resp.Response
	Balance decimal.Decimal `json:"balance,omitempty"`
}

func GetBalance(balanceGetter BalanceGetter, logg *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.transaction.GetBalance"
		address := chi.URLParam(r, "address")

		wallet, err := balanceGetter.GetWallet(address)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("Не найден кошелек по данному адрессу: "+address))

			return
		}

		render.JSON(w, r, BalanceResponse{resp.OK(), wallet.Balance})

	}
}
