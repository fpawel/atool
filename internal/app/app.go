package app

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func Main() {

	// общий контекст приложения с прерыванием
	ctx, interrupt := context.WithCancel(context.Background())

	// открыть конфиг
	cfg.Open(log)

	// соединение с базой данных
	dbFilename := filepath.Join(filepath.Dir(os.Args[0]), "atool.sqlite")
	log.Debug("open database: " + dbFilename)
	db, err := data.Open(dbFilename)
	must.PanicIf(err)

	// старт сервера
	stopServer := runServer(db)

	// старт ожидания сигнала прерывания ОС
	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-done
		log.Info("приложение закрыто сигналом ОС: прервать все фоновые горутины")
		interrupt()
	}()

	<-ctx.Done()

	log.Debug("прервать все фоновые горутины")
	interrupt()

	log.Debug("остановка сервера")
	stopServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func runServer(db *sqlx.DB) context.CancelFunc {

	port, errPort := strconv.Atoi(os.Getenv(EnvVarProductsPort))
	if errPort != nil {
		log.Debug("finding free port to serve api")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv(EnvVarProductsPort, strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Debug("serve api: " + addr)

	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		panic(err)
	}
	handler := &productsServiceHandler{db: db}
	processor := api.NewProductsServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport,
		thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryDefault())

	go log.ErrIfFail(server.Serve, "problem", "`failed to serve`")

	return func() {
		log.ErrIfFail(server.Stop, "problem", "`failed to stop server`")
	}
}

var (
	log = structlog.New()
)

const (
	EnvVarProductsPort = "ATOOL_API_PORT"
)
