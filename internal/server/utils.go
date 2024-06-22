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
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	responseCh := make(chan *http.Response)
	for {
		select {
		case <-ctx.Done():
			break
		case <-timer.C:
			go func() {
				ctx, cancel := context.WithTimeout(ctx, config.ServerConfig.TimeoutOperation)
				defer cancel()
				getLoyaltyOrderInfo(ctx, responseCh, orderNumber)
			}()
		case res := <-responseCh:
			switch res.StatusCode {
			case http.StatusOK:
				data, err := io.ReadAll(res.Body)
				if err != nil {
					continue // TODO log error
				}
				var accrualInfo model.AccrualInfo
				err = json.Unmarshal(data, &accrualInfo)
				if err != nil {
					slog.Error("error occurring unmarshall data from accrual-service:", err.Error())
					continue // TODO log error
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
					break
				}
			case http.StatusTooManyRequests:
				<-time.After(60 * time.Second)
			case http.StatusNoContent:
				break
				// TODO уточнить как обрабатывать
			}
		}
	}

}

func getLoyaltyOrderInfo(ctx context.Context, responseCh chan *http.Response, orderNumber int) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		config.ServerConfig.AccrualSystemAddress+"/api/orders/"+strconv.Itoa(orderNumber),
		nil,
	)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	responseCh <- res
}
