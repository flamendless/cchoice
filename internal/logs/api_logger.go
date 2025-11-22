package logs

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"cchoice/internal/database/queries"

	"go.uber.org/zap"
)

type ServiceStringer interface {
	String() string
}

type ExternalAPILogParams struct {
	CheckoutID *int64
	Service    string
	API        ServiceStringer
	Endpoint   string
	HTTPMethod string
	Payload    any
	Response   any
	Error      error
}

func LogExternalAPICall(ctx context.Context, db *queries.Queries, params ExternalAPILogParams) {
	const logtag = "[Log External API Call]"

	var checkoutID sql.NullInt64
	if params.CheckoutID != nil {
		checkoutID = sql.NullInt64{
			Int64: *params.CheckoutID,
			Valid: true,
		}
	}

	var payloadJSON sql.NullString
	if params.Payload != nil {
		if payloadBytes, err := json.Marshal(params.Payload); err == nil {
			payloadJSON = sql.NullString{
				String: string(payloadBytes),
				Valid:  true,
			}
		} else {
			Log().Warn(logtag, zap.Error(err), zap.String("marshal", "payload"))
		}
	}

	var responseJSON sql.NullString
	if params.Response != nil {
		if responseBytes, err := json.Marshal(params.Response); err == nil {
			responseJSON = sql.NullString{
				String: string(responseBytes),
				Valid:  true,
			}
		} else {
			Log().Warn(logtag, zap.Error(err), zap.String("marshal", "response"))
		}
	}

	statusCode := sql.NullInt64{
		Int64: 200,
		Valid: true,
	}
	isSuccessful := int64(1)

	var errorMessage sql.NullString
	if params.Error != nil {
		statusCode.Int64 = 500
		isSuccessful = 0
		errorMessage = sql.NullString{
			String: params.Error.Error(),
			Valid:  true,
		}
	}

	now := time.Now()

	_, err := db.CreateExternalAPILog(ctx, queries.CreateExternalAPILogParams{
		CheckoutID:   checkoutID,
		Service:      params.Service,
		Api:          params.API.String(),
		Endpoint:     params.Endpoint,
		HttpMethod:   params.HTTPMethod,
		Payload:      payloadJSON,
		Response:     responseJSON,
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
		IsSuccessful: isSuccessful,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	if err != nil {
		Log().Error(logtag, zap.Error(err), zap.String("service", params.Service), zap.Stringer("api", params.API))
	}
}
