package main

import (
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/atool/internal/app"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/powerman/structlog"
	"os"
	"strings"
)

func main() {

	pkg.InitLog()

	var (
		transport thrift.TTransport
		err       error
	)

	addr := "127.0.0.1:" + os.Getenv(app.EnvVarApiPort)
	log.Info(addr)

	transport, err = thrift.NewTSocket(addr)
	if err != nil {
		panic(err)
	}

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTTransportFactory()

	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		panic(err)
	}

	if err := transport.Open(); err != nil {
		panic(err)
	}
	defer transport.Close()

	proto := func() thrift.TProtocol {
		return thrift.NewTMultiplexedProtocol(protocolFactory.GetProtocol(transport), "ScriptService")
	}

	client := api.NewScriptServiceClient(thrift.NewTStandardClient(proto(), proto()))

	if len(os.Args) > 1 && os.Args[1] == "run" {
		filename := strings.Join(os.Args[2:], "")
		if err := client.RunFile(defaultCtx, filename); err != nil {
			log.PrintErr(err)
		}
	}
}

var (
	log        = structlog.New()
	defaultCtx = context.Background()
)
