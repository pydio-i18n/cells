package proxy

import (
	"fmt"
	"sync"

	"github.com/micro/micro/v3/service/broker"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/caddy"
	"github.com/pydio/cells/v4/common/config"
)

type watcher struct {
	services     map[string]map[string]bool
	servicesLock *sync.Mutex
	restartFunc  func()
	loggerFunc   func(msg string, fields ...zapcore.Field)
}

// newWatcher initialize internal resources for a watcher
func newWatcher(loggerFunc func(msg string, fields ...zapcore.Field), restartFunc func()) *watcher {
	w := &watcher{
		restartFunc:  restartFunc,
		loggerFunc:   loggerFunc,
		services:     make(map[string]map[string]bool),
		servicesLock: &sync.Mutex{},
	}
	return w
}

func (w *watcher) subscribeToBroker() error {
	// Adding subscriber
	_, err := broker.Subscribe(common.TopicServiceRegistration, func(m *broker.Message) error {
		sType := string(m.Body)
		sName := m.Header[common.EventHeaderServiceRegisterService]
		// sPeer := p.Message().Header[common.EventHeaderServiceRegisterPeer]
		switch sType {
		//case common.EventTypeServiceRegistered:
		//	if w.restartOnStarted(sName, sPeer) {
		//		w.loggerFunc("Register Message triggers Caddy restart", zap.Any("srvName", sName), zap.Any("headers", p.Message().Header))
		//		w.restartFunc()
		//	}
		//case common.EventTypeServiceUnregistered:
		//	if w.restartOnStopped(sName, sPeer) {
		//		w.loggerFunc("Unregister Message triggers Caddy restart", zap.Any("srvName", sName), zap.Any("headers", p.Message().Header))
		//		w.restartFunc()
		//	}
		case common.EventTypeDebugPrintInternals:
			if sName == common.ServiceGatewayProxy && caddy.LastKnownCaddyFile != "" {
				fmt.Println("***********************************************************************************")
				fmt.Println(" => Caddy file currently served by Gateway Proxy 👇 ")
				fmt.Println(caddy.LastKnownCaddyFile)
				fmt.Println("***********************************************************************************")
			}
		}
		return nil
	})
	return err
}

func (w *watcher) subscribeToConfigs(path ...string) error {
	cw, err := config.Watch(path...)
	if err != nil {
		return err
	}
	go func() {
		defer cw.Stop()
		for {
			if _, err := cw.Next(); err != nil {
				break
			}
			w.loggerFunc("Triggers Caddy restart on config change for path", zap.Strings("path", path))
			w.restartFunc()
		}
	}()
	return nil
}
