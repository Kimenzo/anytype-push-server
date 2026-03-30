package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kimenzo/any-sync/app"
	"github.com/Kimenzo/any-sync/app/logger"
	"github.com/Kimenzo/any-sync/coordinator/coordinatorclient"
	"github.com/Kimenzo/any-sync/coordinator/nodeconfsource"
	"github.com/Kimenzo/any-sync/metric"
	"github.com/Kimenzo/any-sync/nameservice/nameserviceclient"
	"github.com/Kimenzo/any-sync/net/peerservice"
	"github.com/Kimenzo/any-sync/net/pool"
	"github.com/Kimenzo/any-sync/net/rpc/server"
	"github.com/Kimenzo/any-sync/net/secureservice"
	"github.com/Kimenzo/any-sync/net/transport/quic"
	"github.com/Kimenzo/any-sync/net/transport/yamux"
	"github.com/Kimenzo/any-sync/nodeconf"
	"github.com/Kimenzo/any-sync/nodeconf/nodeconfstore"

	// import this to keep govvv in go.mod on mod tidy
	_ "github.com/ahmetb/govvv/integration-test/app-different-package/mypkg"
	"go.uber.org/zap"

	"github.com/Kimenzo/anytype-push-server/account"
	"github.com/Kimenzo/anytype-push-server/config"
	"github.com/Kimenzo/anytype-push-server/db"
	"github.com/Kimenzo/anytype-push-server/push"
	"github.com/Kimenzo/anytype-push-server/queue"
	"github.com/Kimenzo/anytype-push-server/redisprovider"
	"github.com/Kimenzo/anytype-push-server/repo/accountrepo"
	"github.com/Kimenzo/anytype-push-server/repo/spacerepo"
	"github.com/Kimenzo/anytype-push-server/repo/tokenrepo"
	"github.com/Kimenzo/anytype-push-server/sender"
	"github.com/Kimenzo/anytype-push-server/sender/provider/fcm"
)

var log = logger.NewNamed("push.main")

var (
	flagConfigFile = flag.String("c", "etc/anytype-push-server.yml", "path to config file")
	flagVersion    = flag.Bool("v", false, "show version and exit")
	flagHelp       = flag.Bool("h", false, "show help and exit")
)

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(app.AppName)
		fmt.Println(app.Version())
		fmt.Println(app.VersionDescription())
		return
	}
	if *flagHelp {
		flag.PrintDefaults()
		return
	}

	if debug, ok := os.LookupEnv("ANYPROF"); ok && debug != "" {
		go func() {
			http.ListenAndServe(debug, nil)
		}()
	}

	// create app
	ctx := context.Background()
	a := new(app.App)

	// open config file
	conf, err := config.NewFromFile(*flagConfigFile)
	if err != nil {
		log.Fatal("can't open config file", zap.Error(err))
	}

	// bootstrap components
	a.Register(conf)
	Bootstrap(a)

	// start app
	if err := a.Start(ctx); err != nil {
		log.Fatal("can't start app", zap.Error(err))
	}
	log.Info("app started", zap.String("version", a.Version()))

	// wait exit signal
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-exit
	log.Info("received exit signal, stop app...", zap.String("signal", fmt.Sprint(sig)))

	// close app
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if err := a.Close(ctx); err != nil {
		log.Fatal("close error", zap.Error(err))
	} else {
		log.Info("goodbye!")
	}
	time.Sleep(time.Second / 3)
}

func Bootstrap(a *app.App) {
	a.Register(db.New()).
		Register(redisprovider.New()).
		Register(metric.New()).
		Register(server.New()).
		Register(account.New()).
		Register(pool.New()).
		Register(peerservice.New()).
		Register(coordinatorclient.New()).
		Register(nameserviceclient.New()).
		Register(nodeconfsource.New()).
		Register(nodeconfstore.New()).
		Register(nodeconf.New()).
		Register(secureservice.New()).
		Register(tokenrepo.New()).
		Register(accountrepo.New()).
		Register(spacerepo.New()).
		Register(queue.New()).
		Register(sender.New()).
		Register(fcm.New()).
		Register(push.New()).
		Register(quic.New()).
		Register(yamux.New())
}

