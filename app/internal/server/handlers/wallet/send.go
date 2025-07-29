package wallet

import (
	resp "github.com/bifk/testTask/internal/lib/api/response"
	"github.com/bifk/testTask/internal/logger"
	"github.com/go-chi/render"
	"github.com/shopspring/decimal"
	"net/http"
)

type SendRequest struct {
	From   string          `json:"from"`
	To     string          `json:"to"`
	Amount decimal.Decimal `json:"amount"`
}

type SendResponse struct {
	resp.Response
}

type Sender interface {
	Send(from string, to string, amount decimal.Decimal) error
}

func Send(sender Sender, logg *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.wallet.Send"

		var req SendRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logg.Errorf("%s: %s", op, "Не удалось декодировать тело запроса: "+err.Error())
			render.JSON(w, r, resp.Error("Не удалось декодировать тело запроса: "+err.Error()))

			return
		}

		err = sender.Send(req.From, req.To, req.Amount)
		if err != nil {
			logg.Errorf("%s: %s", op, err.Error())
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		render.JSON(w, r, SendResponse{resp.OK()})
	}
}
