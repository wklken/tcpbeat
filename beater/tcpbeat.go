package beater

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/wklken/tcpbeat/config"
	"github.com/wklken/tcpbeat/input/tcp"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Tcpbeat struct {
	sync.RWMutex
	config    config.Config
	server    net.Listener
	clients   map[*tcp.Client]struct{}
	wg        sync.WaitGroup
	done      chan struct{}
	splitFunc bufio.SplitFunc
	log       *logp.Logger
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	server, err := net.Listen("tcp", config.Host)
	if err != nil {
		logp.Err("net tcp fail:%s", config.Host)
		return nil, err
	}

	sf := tcp.SplitFunc([]byte(config.LineDelimiter))

	bt := &Tcpbeat{
		done:      make(chan struct{}),
		config:    config,
		clients:   make(map[*tcp.Client]struct{}, 0),
		server:    server,
		splitFunc: sf,
		log:       logp.NewLogger("tcpbeat").With("address", config.Host),
	}
	return bt, nil
}

func publishEvent(c config.Config, client beat.Client, logData []byte) {

	fields := common.MapStr{}
	err := json.Unmarshal(logData, &fields)
	if err != nil {
		logp.Err("Could not load json formated event: %v", err)
		//event["message"] = logData
		//event["tags"] = []string{"_tcpbeat_jsonparsefailure"}
	}

	event := beat.Event{
		Timestamp: time.Now(),
		Fields: common.MapStr{
			"message":  fields["message"],
			c.LabelKey: c.Labels,
		},
	}
	client.Publish(event)
}

func genPublishCallbackFunc(c config.Config, client beat.Client) tcp.CallbackFunc {
	return func(data []byte) {
		publishEvent(c, client, data)
	}
}

func (bt *Tcpbeat) Run(b *beat.Beat) error {
	logp.Info("tcpbeat is running! Hit CTRL-C to stop it.")

	logp.Info("Started listening for incoming TCP connection on: %s", bt.config.Host)
	pubClient, err := b.Publisher.Connect()
	if err != nil {
		// TODO: do something
	}

	for {
		logp.Info("got one tcp connection via server.Accept()")
		conn, err := bt.server.Accept()
		if err != nil {
			select {
			case <-bt.done:
				return nil
			default:
				logp.Err("tcp", "Can not accept the connection: %s", err)
				continue
			}
		}

		callback := genPublishCallbackFunc(bt.config, pubClient)
		client := tcp.NewClient(
			conn,
			bt.log,
			callback,
			bt.splitFunc,
			config.DefaultConfig.MaxMessageSize,
			config.DefaultConfig.Timeout,
		)
		bt.log.Debugw("New client", "remote_address", conn.RemoteAddr(), "total", bt.clientsCount())
		bt.wg.Add(1)
		go func() {
			defer logp.Recover("recovering from a tcp client crash")

			defer bt.wg.Done()
			defer conn.Close()

			bt.registerClient(client)
			defer bt.unregisterClient(client)

			logp.Info("tcp: New client, remote: %s (total clients: %d)", conn.RemoteAddr(), bt.clientsCount())

			logp.Info("start cleint.Handle")
			err := client.Handle()
			if err != nil {
				bt.log.Debugw("Client error", "error", err)
			}

			bt.log.Info(
				"tcp: Client disconnected, remote: %s (total clients: %d)",
				conn.RemoteAddr(),
				bt.clientsCount(),
			)
			logp.Info("stop cleint.Handle")
		}()
	}

}

func (bt *Tcpbeat) Stop() {
	logp.Info("shutting down.")
	//bt.client.Close()

	logp.Info("Stopping TCP harvester")
	//  close(h.done)
	bt.server.Close()
	//
	logp.Debug("tcp", "Closing remote connections")
	for _, client := range bt.allClients() {
		client.Close()
	}
	bt.wg.Wait()
	logp.Debug("tcp", "Remote connections closed")

	close(bt.done)
}

func (bt *Tcpbeat) Reload(cfg *common.Config) {
	logp.Info("reload")
	c := config.DefaultConfig
	err := cfg.Unpack(&c)
	if err != nil {
		logp.Err("error reading configuration file")
	}
	logp.Info("config:%+v", c)
	// TODO parse config here
	bt.config = c
}

func (bt *Tcpbeat) registerClient(client *tcp.Client) {
	bt.Lock()
	defer bt.Unlock()
	bt.clients[client] = struct{}{}
}

func (bt *Tcpbeat) unregisterClient(client *tcp.Client) {
	bt.Lock()
	defer bt.Unlock()
	delete(bt.clients, client)
}

func (bt *Tcpbeat) allClients() []*tcp.Client {
	bt.RLock()
	defer bt.RUnlock()
	currentClients := make([]*tcp.Client, len(bt.clients))
	idx := 0
	for client := range bt.clients {
		currentClients[idx] = client
		idx++
	}
	return currentClients
}

func (bt *Tcpbeat) clientsCount() int {
	bt.RLock()
	defer bt.RUnlock()
	return len(bt.clients)
}
