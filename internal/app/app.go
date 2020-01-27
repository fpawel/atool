package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/atool/internal"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
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
	log.Debug("open database: "+dbFilename, structlog.KeyTime, time.Now().Format("15:04:05"))
	db, err = data.Open(dbFilename)
	must.PanicIf(err)

	// журнал СОМ порта
	comportLogfile, err = logfile.New(".comport")
	must.PanicIf(err)

	// инициализация отправки оповещений с посылками СОМ порта в gui
	comm.SetNotify(notifyComm)

	// старт сервера
	stopServer := runServer()

	if envVarDevModeSet() {
		log.Printf("waiting system signal because of %s=%s", internal.EnvVarDevMode, os.Getenv(internal.EnvVarDevMode))
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

	log.Debug("закрыть соединение с базой данных", structlog.KeyTime, time.Now().Format("15:04:05"))
	log.ErrIfFail(db.Close)

	log.Debug("закрыть журнал СОМ порта")
	log.ErrIfFail(comportLogfile.Close)

	log.Debug("закрыть журнал")
	log.ErrIfFail(journal.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func notifyComm(x comm.Info) {
	ct := gui.CommTransaction{
		Port:     x.Port,
		Request:  fmt.Sprintf("% X", x.Request),
		Response: fmt.Sprintf("% X", x.Response),
		Ok:       x.Err == nil,
	}
	if x.Err != nil {
		if len(x.Response) > 0 {
			ct.Response += " "
		}
		ct.Response += x.Err.Error()
	}
	ct.Response += " " + x.Duration.String()
	if x.Attempt > 0 {
		ct.Response += fmt.Sprintf(" попытка %d", x.Attempt+1)
	}
	go gui.NotifyNewCommTransaction(ct)

	_, err := fmt.Fprintf(comportLogfile, "%s %s % X -> % X", time.Now().Format("15:04:05.000"), x.Port, x.Request, x.Response)
	must.PanicIf(err)
	if x.Err != nil {
		_, err := fmt.Fprintf(comportLogfile, " %s", x.Err)
		must.PanicIf(err)
	}
	_, err = comportLogfile.WriteString("\n")
	must.PanicIf(err)
}

// newApiProcessor
func newApiProcessor() thrift.TProcessor {
	p := thrift.NewTMultiplexedProcessor()

	p.RegisterProcessor("RunWorkService",
		api.NewRunWorkServiceProcessor(new(runWorkSvc)))
	p.RegisterProcessor("FilesService",
		api.NewFilesServiceProcessor(new(filesSvc)))
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
	p.RegisterProcessor("ScriptService",
		api.NewScriptServiceProcessor(new(scriptSvc)))
	p.RegisterProcessor("ProductParamService",
		api.NewProductParamServiceProcessor(new(prodPrmSvc)))
	return p
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
		log.ErrIfFail(func() error {
			if err := server.Stop(); err != nil {
				return fmt.Errorf("server.Stop(): %w", err)
			}
			if err := transport.Close(); err != nil {
				return fmt.Errorf("transport.Close(): %w", err)
			}
			return nil
		}, "problem", "`failed to stop server`")
	}
}

var (
	log            = structlog.New()
	tmpDir         = filepath.Join(filepath.Dir(os.Args[0]), "tmp")
	db             *sqlx.DB
	appCtx         context.Context
	comportLogfile *os.File
)

func envVarDevModeSet() bool {
	return os.Getenv(internal.EnvVarDevMode) == "true"
}

func getTCPAddrApi() string {
	port, errPort := strconv.Atoi(os.Getenv(internal.EnvVarApiPort))
	if errPort != nil {
		log.Debug("search free port to serve api")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv(internal.EnvVarApiPort, strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func cleanTmpDir() {
	if err := os.RemoveAll(tmpDir); err != nil {
		log.PrintErr(merry.Append(err, "os.RemoveAll(tmpDir)"))
	}
}
