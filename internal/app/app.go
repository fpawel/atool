package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/atool/internal"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/ankt"
	"github.com/fpawel/atool/internal/devtypes/ikds4"
	"github.com/fpawel/atool/internal/devtypes/mil82"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
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

type BuildInfo struct {
	Commit    string
	CommitGui string
	UUID      string
	Date      string
	Time      string
}

func Main(buildInfoIn BuildInfo) {

	buildInfo = buildInfoIn

	cleanTmpDir()
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		log.PrintErr(merry.Append(err, "os.RemoveAll(tmpDir)"))
	}
	defer cleanTmpDir()

	log.Debug("открыть журнал")
	log.ErrIfFail(workgui.OpenJournal)

	// инициализация конфигурации
	appcfg.Init(mil82.Device, ikds4.Device, ankt.Device)

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(filepath.Dir(os.Args[0]), "atool.sqlite")
	log.Debug("open database: " + dbFilename)
	must.PanicIf(data.Open(dbFilename))

	// журнал СОМ порта
	comportLogfile, err = logfile.New(".comport")
	must.PanicIf(err)

	// инициализация отправки оповещений с посылками СОМ порта в gui
	comm.SetNotify(notifyComm)

	// старт сервера
	stopApiServer := runApiServer()

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
	workgui.Interrupt()
	workgui.Wait()

	log.Debug("остановка сервера api")
	stopApiServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(data.DB.Close)

	log.Debug("закрыть журнал СОМ порта")
	log.ErrIfFail(comportLogfile.Close)

	log.Debug("закрыть журнал")
	log.ErrIfFail(workgui.CloseJournal)

	log.Debug("сохранить конфигурацию")
	log.ErrIfFail(appcfg.Cfg.Save)
	log.ErrIfFail(appcfg.Sets.Save)

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

	_, err := fmt.Fprintf(comportLogfile, "%s %s % X -> % X %s", time.Now().Format("15:04:05.000"), x.Port, x.Request, x.Response, x.Duration)
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

	p.RegisterProcessor("FileService",
		api.NewFileServiceProcessor(new(fileSvc)))

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

	p.RegisterProcessor("ProductParamService",
		api.NewProductParamServiceProcessor(new(prodPrmSvc)))

	p.RegisterProcessor("WorkDialogService",
		api.NewWorkDialogServiceProcessor(new(workDialogSvc)))

	p.RegisterProcessor("JournalService",
		api.NewJournalServiceProcessor(new(journalSvc)))

	p.RegisterProcessor("AppInfoService",
		api.NewAppInfoServiceProcessor(new(appInfoSvc)))

	return p
}

func runApiServer() context.CancelFunc {

	addr := getTCPAddrEnvVar(internal.EnvVarApiPort)
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
				return merry.Prepend(err, "server.Stop")
			}
			if err := transport.Close(); err != nil {
				return merry.Prepend(err, "transport.Close")
			}
			return nil
		}, "problem", "`failed to stop server`")
	}
}

func envVarDevModeSet() bool {
	return os.Getenv(internal.EnvVarDevMode) == "true"
}

func getTCPAddrEnvVar(envVar string) string {
	port, errPort := strconv.Atoi(os.Getenv(envVar))
	if errPort != nil {
		log.Debug("search free port to serve: " + envVar)
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv(envVar, strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func cleanTmpDir() {
	if err := os.RemoveAll(tmpDir); err != nil {
		log.PrintErr(merry.Append(err, "os.RemoveAll(tmpDir)"))
	}
}

func runSingleTask(w workgui.WorkFunc) {
	go w.RunSingleTask(log, appCtx)
}

func runWork(w workgui.Work) error {
	return w.Run(log, appCtx)
}

func runWorkFunc(name string, w workgui.WorkFunc) error {
	return w.Work(name).Run(log, appCtx)
}

func runWithNotifyPartyChanged(name string, w workgui.WorkFunc) error {
	return workgui.New(name, func(log comm.Logger, ctx context.Context) error {
		defer func() {
			go gui.NotifyCurrentPartyChanged()
		}()
		gui.ShowModalMessage(name)
		return w(log, ctx)
	}).Run(log, appCtx)
}

func runWithNotifyArchiveChanged(name string, w workgui.WorkFunc) error {
	return workgui.New(name, func(log comm.Logger, ctx context.Context) error {
		defer func() {
			go gui.NotifyCurrentPartyChanged()
			go gui.NotifyPartiesArchiveChanged()
		}()
		gui.ShowModalMessage(name)
		return w(log, ctx)
	}).Run(log, appCtx)
}

var (
	buildInfo      BuildInfo
	log            = structlog.New()
	tmpDir         = filepath.Join(filepath.Dir(os.Args[0]), "tmp")
	appCtx         context.Context
	comportLogfile *os.File
)
