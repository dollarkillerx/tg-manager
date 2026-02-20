package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tg-manager/internal/conf"
	"github.com/tg-manager/internal/forwarder"
	"github.com/tg-manager/internal/middleware"
	"github.com/tg-manager/internal/storage"
	"github.com/tg-manager/internal/telegram"
)

type ApiServer struct {
	storage  *storage.Storage
	tgSvc    *telegram.Service
	engine   *forwarder.Engine
	conf     conf.Config
	app      *gin.Engine
	rpcHandler *RpcHandler
}

func NewApiServer(st *storage.Storage, tgSvc *telegram.Service, engine *forwarder.Engine, conf conf.Config) *ApiServer {
	server := &ApiServer{
		storage:    st,
		tgSvc:      tgSvc,
		engine:     engine,
		conf:       conf,
		rpcHandler: NewRpcHandler(),
	}
	server.registerRpcMethods()
	return server
}

func (a *ApiServer) Run() error {
	if a.app == nil {
		if !a.conf.ServiceConfiguration.Debug {
			gin.SetMode(gin.ReleaseMode)
		}
		a.app = gin.New()
		a.app.Use(middleware.HttpRecover())
		a.app.Use(gin.Logger())
		a.app.Use(middleware.Cors())
	}
	a.Router()
	return a.app.Run(fmt.Sprintf(":%s", a.conf.ServiceConfiguration.Port))
}

func (a *ApiServer) Router() {
	a.app.GET("/health", a.HealthCheck)
	a.app.POST("/api/rpc", a.Rpc)

	// Serve frontend static files in production
	a.app.Static("/assets", "./frontend/dist/assets")
	a.app.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})
}

func (a *ApiServer) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "healthy"})
}

func (a *ApiServer) Rpc(ctx *gin.Context) {
	a.rpcHandler.HandleRpcRequest(ctx)
}

func (a *ApiServer) registerRpcMethods() {
	// Auth methods
	a.rpcHandler.RegisterMethod(&AuthStatusMethod{tgSvc: a.tgSvc})
	a.rpcHandler.RegisterMethod(&AuthSendCodeMethod{tgSvc: a.tgSvc})
	a.rpcHandler.RegisterMethod(&AuthVerifyCodeMethod{tgSvc: a.tgSvc})
	a.rpcHandler.RegisterMethod(&AuthSendPasswordMethod{tgSvc: a.tgSvc})
	// Dialog methods
	a.rpcHandler.RegisterMethod(&DialogsListMethod{tgSvc: a.tgSvc})
	a.rpcHandler.RegisterMethod(&ChannelsListMethod{tgSvc: a.tgSvc})
	// Rule methods
	a.rpcHandler.RegisterMethod(&RulesCreateMethod{storage: a.storage, engine: a.engine})
	a.rpcHandler.RegisterMethod(&RulesListMethod{storage: a.storage})
	a.rpcHandler.RegisterMethod(&RulesUpdateMethod{storage: a.storage, engine: a.engine})
	a.rpcHandler.RegisterMethod(&RulesDeleteMethod{storage: a.storage, engine: a.engine})
	// Message methods
	a.rpcHandler.RegisterMethod(&MessagesHistoryMethod{tgSvc: a.tgSvc})
}
