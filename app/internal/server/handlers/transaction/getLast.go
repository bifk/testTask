package transaction

import (
	"github.com/bifk/testTask/internal/domain/models"
	resp "github.com/bifk/testTask/internal/lib/api/response"
	"github.com/bifk/testTask/internal/logger"
	"github.com/go-chi/render"
	"net/http"
	"strconv"
)

type LastGetter interface {
	GetLast(count int) ([]models.Transaction, error)
}

type GetLastResponse struct {
	resp.Response
	Transactions []models.Transaction `json:"transactions"`
}

func GetLast(lastGetter LastGetter, logg *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.transaction.GetLast"
		countStr := r.URL.Query().Get("count")

		var count int

		// Проверка на корректность параметра count
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count < 0 {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Некоректный параметр count"))

			return
		}

		transactions, err := lastGetter.GetLast(count)
		if err != nil {
			logg.Errorf("%s: %s", op, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))

			return
		}

		render.JSON(w, r, GetLastResponse{resp.OK(), transactions})
	}
}
