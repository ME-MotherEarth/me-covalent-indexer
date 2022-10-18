package factory

import (
	"net/http"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	covalent "github.com/ME-MotherEarth/me-covalent-indexer"
	"github.com/ME-MotherEarth/me-covalent-indexer/process"
	"github.com/ME-MotherEarth/me-covalent-indexer/process/factory"
	logger "github.com/ME-MotherEarth/me-logger"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var log = logger.GetOrCreate("covalentIndexer")

// ArgsCovalentIndexerFactory holds all input dependencies required by covalent data indexer factory
// in order to create new instances
type ArgsCovalentIndexerFactory struct {
	Enabled              bool
	URL                  string
	RouteSendData        string
	RouteAcknowledgeData string
	PubKeyConverter      core.PubkeyConverter
	Accounts             covalent.AccountsAdapter
	Hasher               hashing.Hasher
	Marshaller           marshal.Marshalizer
	ShardCoordinator     process.ShardCoordinator
}

// CreateCovalentIndexer creates a new Driver instance of type covalent data indexer
func CreateCovalentIndexer(args *ArgsCovalentIndexerFactory) (covalent.Driver, error) {
	if check.IfNil(args.PubKeyConverter) {
		return nil, covalent.ErrNilPubKeyConverter
	}
	if check.IfNil(args.Accounts) {
		return nil, covalent.ErrNilAccountsAdapter
	}
	if check.IfNil(args.Hasher) {
		return nil, covalent.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, covalent.ErrNilMarshaller
	}

	argsDataProcessor := &factory.ArgsDataProcessor{
		PubKeyConvertor:  args.PubKeyConverter,
		Accounts:         args.Accounts,
		Hasher:           args.Hasher,
		Marshaller:       args.Marshaller,
		ShardCoordinator: args.ShardCoordinator,
	}

	dataProcessor, err := factory.CreateDataProcessor(argsDataProcessor)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	server := &http.Server{
		Addr:    args.URL,
		Handler: router,
	}

	ci, err := covalent.NewCovalentDataIndexer(dataProcessor, server)
	if err != nil {
		return nil, err
	}

	routeSendData := router.HandleFunc(args.RouteSendData, func(w http.ResponseWriter, r *http.Request) {
		log.Debug("new connection", "route", args.RouteSendData)
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, errUpgrade := upgrader.Upgrade(w, r, nil)
		if errUpgrade != nil {
			log.Warn("could not upgrade http connection to websocket", "error", errUpgrade)
			return
		}

		ci.SetWSSender(ws)
	})

	if routeSendData.GetError() != nil {
		log.Error("websocket router failed to handle send data",
			"route", routeSendData.GetName(),
			"error", routeSendData.GetError())
	}

	routeAcknowledgeData := router.HandleFunc(args.RouteAcknowledgeData, func(w http.ResponseWriter, r *http.Request) {
		log.Debug("new connection", "route", args.RouteAcknowledgeData)
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, errUpgrade := upgrader.Upgrade(w, r, nil)
		if errUpgrade != nil {
			log.Warn("could not upgrade http connection to websocket", "error", errUpgrade)
			return
		}

		ci.SetWSReceiver(ws)
	})

	if routeAcknowledgeData.GetError() != nil {
		log.Error("websocket router failed to acknowledge sent data",
			"route", routeSendData.GetName(),
			"error", routeSendData.GetError())
	}

	return ci, nil
}
