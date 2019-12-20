package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui/guiwork"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func Main() {

	cleanTmpDir()
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		log.PrintErr(merry.Append(err, "os.RemoveAll(tmpDir)"))
	}
	defer cleanTmpDir()

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(filepath.Dir(os.Args[0]), "atool.sqlite")
	log.Debug("open database: " + dbFilename)
	db, err = data.Open(dbFilename)
	must.PanicIf(err)
	// старт сервера
	stopServer := runServer()

	if envVarDevModeSet() {
		log.Printf("waiting system signal because of %s=%s", envVarDevMode, os.Getenv(envVarDevMode))
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-done
		log.Debug("system signal: " + sig.String())
	} else {
		cmd := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "atoolgui.exe"))
		log.ErrIfFail(cmd.Start)
		log.ErrIfFail(cmd.Wait)
		log.Debug("gui was closed.")
	}

	log.Debug("прервать все фоновые горутины")
	interrupt()
	guiwork.Interrupt()
	guiwork.Wait()

	log.Debug("остановка сервера")
	stopServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func runServer() context.CancelFunc {

	addr := getTCPAddrApi()
	log.Debug("serve api: " + addr)

	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		panic(err)
	}

	server := thrift.NewTSimpleServer4(newApiProcessor(), transport,
		thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryDefault())

	go log.ErrIfFail(server.Serve, "problem", "`failed to serve`")

	return func() {
		log.ErrIfFail(server.Stop, "problem", "`failed to stop server`")
	}
}

var (
	log    = structlog.New()
	tmpDir = filepath.Join(filepath.Dir(os.Args[0]), "tmp")
	db     *sqlx.DB
	appCtx context.Context
)

const (
	envVarApiPort = "ATOOL_API_PORT"
	envVarDevMode = "ATOOL_DEV_MODE"
)

func envVarDevModeSet() bool {
	return os.Getenv(envVarDevMode) == "true"
}

func newApiProcessor() thrift.TProcessor {
	p := thrift.NewTMultiplexedProcessor()
	p.RegisterProcessor("FilesService",
		api.NewFilesServiceProcessor(new(filesSvc)))
	p.RegisterProcessor("HardwareConnectionService",
		api.NewHardwareConnectionServiceProcessor(new(hardwareConnSvc)))
	p.RegisterProcessor("CurrentFileService",
		api.NewCurrentFileServiceProcessor(new(currentFileSvc)))
	p.RegisterProcessor("ProductService",
		api.NewProductServiceProcessor(new(productSvc)))
	p.RegisterProcessor("AppConfigService",
		api.NewAppConfigServiceProcessor(new(appConfigSvc)))
	p.RegisterProcessor("NotifyGuiService",
		api.NewNotifyGuiServiceProcessor(new(notifyGuiSvc)))
	p.RegisterProcessor("HelperService",
		api.NewHelperServiceProcessor(new(helperSvc)))
	p.RegisterProcessor("TemperatureDeviceService",
		api.NewTemperatureDeviceServiceProcessor(new(tempDeviceSvc)))
	p.RegisterProcessor("CoefficientsService",
		api.NewCoefficientsServiceProcessor(new(coefficientsSvc)))

	return p
}

func getTCPAddrApi() string {
	port, errPort := strconv.Atoi(os.Getenv(envVarApiPort))
	if errPort != nil {
		log.Debug("search free port to serve api")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv(envVarApiPort, strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func cleanTmpDir() {
	if err := os.RemoveAll(tmpDir); err != nil {
		log.PrintErr(merry.Append(err, "os.RemoveAll(tmpDir)"))
	}
}
