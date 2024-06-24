package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/model"
	"github.com/SpaceSlow/loyalty/internal/store"
)

func getOrderNumber(body io.ReadCloser) (int, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return -1, err
	}
	defer body.Close()
	return strconv.Atoi(string(data))
}

func isValidLuhnAlgorithm(number int) bool {
	sum := 0

	for isEven := 1; number > 0; isEven ^= 1 {
		digit := number % 10
		if isEven == 0 {
			digit = digit * 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		number /= 10
	}

	return sum%10 == 0
}

func CalculateAccrual(ctx context.Context, db *store.DB, orderNumber string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	responseCh := make(chan response)
	errorCh := make(chan error)
out:
	for {
		select {
		case <-ctx.Done():
			break out
		case <-ticker.C:
			go func() {
				ctx, cancel := context.WithTimeout(ctx, config.ServerConfig.TimeoutOperation)
				defer cancel()
				getLoyaltyOrderInfo(ctx, responseCh, errorCh, orderNumber)
			}()
		case err := <-errorCh:
			slog.Error("error occurring when making a request to an external service:", slog.Any("err", err.Error()))
		case res := <-responseCh:
			switch res.statusCode {
			case http.StatusOK:
				var accrual model.ExternalAccrual
				err := json.Unmarshal(res.data, &accrual)
				if err != nil {
					slog.Error("error occurring unmarshall data from accrual-service:", slog.Any("err", err.Error()))
					continue
				}
				switch accrual.Status {
				case "REGISTERED", "PROCESSING":
					continue
				case "PROCESSED", "INVALID":
					ctx, cancel := context.WithTimeout(ctx, config.ServerConfig.TimeoutOperation)
					defer cancel()
					err := db.UpdateAccrualInfo(ctx, accrual)
					if err != nil {
						slog.Error("error occurring in updating accrual in DB:", err.Error(), accrual)
						continue
					}
					break out
				}
			case http.StatusTooManyRequests:
				<-time.After(60 * time.Second)
			case http.StatusNoContent:
				break out // TODO уточнить как обрабатывать
			}
		}
	}

}

type response struct {
	statusCode int
	data       []byte
}

func getLoyaltyOrderInfo(ctx context.Context, responseCh chan response, errorCh chan error, orderNumber string) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		config.ServerConfig.AccrualSystemAddress+"/api/orders/"+orderNumber,
		nil,
	)
	if err != nil {
		errorCh <- err
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errorCh <- err
		return
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		errorCh <- err
		return
	}
	responseCh <- response{res.StatusCode, data}
}
