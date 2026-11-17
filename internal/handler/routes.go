package handler

import (
	"net/http"
	"qingdao_backend/internal/handler/timestamp"
	"qingdao_backend/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/rsa",
			Handler: timestamp.Apply(ctx.Timestamp, "SHA256"),
		},
		{
			Method:  http.MethodPost,
			Path:    "/sm",
			Handler: timestamp.Apply(ctx.Timestamp, "SM3"),
		},
		{
			Method:  http.MethodPost,
			Path:    "/verify",
			Handler: timestamp.Verify(ctx.Timestamp),
		},
	},
		rest.WithPrefix("/api/timestamp"))
}
