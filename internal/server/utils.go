package server

import (
	"context"
	"encoding/json"
	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/SpaceSlow/loyalty/internal/model"
	"github.com/SpaceSlow/loyalty/internal/store"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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
	number = reverseNumber(number)

	for isOdd := 1; number > 0; isOdd ^= 1 {
		digit := number % 10
		if isOdd == 1 {
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

func reverseNumber(number int) int {
	reversed := 0
	for ; number > 0; number /= 10 {
		reversed = reversed*10 + number%10
	}
	return reversed
}

func CalculateAccrual(ctx context.Context, db *store.DB, orderNumber int) {
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
			slog.Error("error occurring when making a request to an external service:", err.Error())
		case res := <-responseCh:
			switch res.statusCode {
			case http.StatusOK:
				var accrualInfo model.AccrualInfo
				err := json.Unmarshal(res.data, &accrualInfo)
				if err != nil {
					slog.Error("error occurring unmarshall data from accrual-service:", err.Error())
					continue
				}
				switch accrualInfo.Status {
				case "REGISTERED", "PROCESSING":
					continue
				case "PROCESSED", "INVALID":
					ctx, cancel := context.WithTimeout(ctx, config.ServerConfig.TimeoutOperation)
					defer cancel()
					err := db.UpdateAccrualInfo(ctx, accrualInfo)
					if err != nil {
						slog.Error("error occurring in updating accrual in DB:", err.Error(), accrualInfo)
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

func getLoyaltyOrderInfo(ctx context.Context, responseCh chan response, errorCh chan error, orderNumber int) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		config.ServerConfig.AccrualSystemAddress+"/api/orders/"+strconv.Itoa(orderNumber),
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
