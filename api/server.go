package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"

	"transfers/api/models"
	db "transfers/db/sqlc"
	"transfers/service"
	"transfers/util"
)

type Server struct {
	store  db.Store
	engine *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", post[models.CreateAccountRequest, models.CreateAccountResponse](&service.CreateAccountService{Store: store}))
	router.GET("/accounts/:account_id", get[models.GetAccountRequest, models.GetAccountResponse](&service.GetAccountService{Store: store}))
	router.POST("/transactions", post[models.CreateTransactionRequest, models.CreateTransactionResponse](&service.CreateTransactionService{Store: store}))

	server.engine = router
	return server
}

func (s *Server) Run(address string) error {
	return s.engine.Run(address)
}

type Service[Request, Response any] interface {
	Validate(context.Context, *Request) error
	Do(context.Context, *Request) (*Response, error)
}

func get[Req, Resp any](svc Service[Req, Resp]) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		if err := svc.Validate(ctx, &req); err != nil {
			ctx.JSON(status(err, http.StatusBadRequest), errorResponse(err))
			return
		}
		resp, err := svc.Do(ctx, &req)
		if err != nil {
			ctx.JSON(status(err, http.StatusInternalServerError), errorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func post[Req, Resp any](svc Service[Req, Resp]) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		if err := svc.Validate(ctx, &req); err != nil {
			ctx.JSON(status(err, http.StatusBadRequest), errorResponse(err))
			return
		}
		resp, err := svc.Do(ctx, &req)
		if err != nil {
			ctx.JSON(status(err, http.StatusInternalServerError), errorResponse(err))
			return
		}
		ctx.JSON(http.StatusCreated, resp)
	}
}

func status(err error, fallback int) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errorx.HasTrait(err, errorx.Duplicate()):
		return http.StatusBadRequest
	case errorx.HasTrait(err, errorx.NotFound()):
		return http.StatusNotFound
	case errorx.HasTrait(err, util.PaymentRequired):
		return http.StatusPaymentRequired
	case errorx.IsOfType(err, errorx.ExternalError):
		return http.StatusInternalServerError
	case errorx.IsOfType(err, errorx.IllegalArgument):
		return http.StatusBadRequest
	default:
		return fallback
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
