/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ers

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPER implements EventReader interface for AMQP messaging.
type AMQPER struct {
	cgrCfg *config.CGRConfig
	cfgIdx int // index of current config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	eventChan     chan *erEvent // channel to dispatch the events created
	partialEvChan chan *erEvent // channel to dispatch the partial events created
	errChan       chan error

	client *amqpClient // AMQP client for managing connections and subscriptions.
}

type amqpClient struct {
	exchange     string
	exchangeType string
	routingKey   string
	queueID      string
	consumer     string

	connection      *amqp.Connection
	channel         *amqp.Channel
	available       bool          // indicates if the AMQP channel has been established and is ready for use.
	done            chan struct{} // done signals the shutdown of the AMQP connection.
	mu              sync.RWMutex
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error

	// prefetchCount defines the maximum number of messages that the server will
	// deliver to the consumer without waiting for acknowledgements. It's used to
	// control the QoS settings for the AMQP channel.
	prefetchCount int
}

// NewAMQPER returns a new AMQP EventReader with the provided configurations.
func NewAMQPER(cfg *config.CGRConfig, cfgIdx int, eventChan, partialEvChan chan *erEvent, errChan chan error,
	fltrS *engine.FilterS, exitChan chan struct{}) (EventReader, error) {
	rdr := &AMQPER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		eventChan:     eventChan,
		partialEvChan: partialEvChan,
		errChan:       errChan,
	}
	rdr.createClient(rdr.Config().Opts.AMQP, exitChan)
	return rdr, nil
}

// Config returns the curent configuration
func (rdr *AMQPER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// createClient initializes the AMQP client with the necessary configurations.
func (rdr *AMQPER) createClient(opts *config.AMQPROpts, exitChan chan struct{}) {
	rdrCfg := rdr.Config()
	rdr.client = &amqpClient{
		prefetchCount: rdrCfg.ConcurrentReqs,
		done:          exitChan,
	}
	if opts != nil {
		if opts.QueueID != nil {
			rdr.client.queueID = *opts.QueueID
		}
		if opts.ConsumerTag != nil {
			rdr.client.consumer = *opts.ConsumerTag
		}
		if opts.RoutingKey != nil {
			rdr.client.routingKey = *opts.RoutingKey
		}
		if opts.Exchange != nil {
			rdr.client.exchange = *opts.Exchange
		}
		if opts.ExchangeType != nil {
			rdr.client.exchangeType = *opts.ExchangeType
		}
	}
	go rdr.client.handleReconnect(rdrCfg.SourcePath, rdrCfg.ID,
		rdrCfg.Reconnects, rdrCfg.MaxReconnectInterval)
}

func (rdr *AMQPER) processMessage(msg []byte) error {
	var decodedMessage map[string]any
	err := json.Unmarshal(msg, &decodedMessage)
	if err != nil {
		return err
	}
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaReaderID: utils.NewLeafNode(rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].ID)}}

	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), reqVars,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil) // create an AgentRequest
	pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq)
	if err != nil || !pass {
		return err
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return err
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.eventChan
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvChan
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return nil
}

// Serve starts the goroutine needed to monitor and process delieveries coming from the AMQP queue.
func (rdr *AMQPER) Serve() error {
	rdrCfg := rdr.Config()
	if rdrCfg.RunDelay == time.Duration(0) {
		return nil
	}

	// Wait for the client to be ready
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	fib := utils.FibDuration(time.Millisecond, 0)
	for !rdr.client.isAvailable() {
		select {
		case <-ctx.Done():
			return fmt.Errorf("client not ready to start consuming, error: %w", ctx.Err())
		default:
			time.Sleep(fib())
		}
	}

	// Start consuming messages from the queue.
	deliveries, err := rdr.client.Consume()
	if err != nil {
		return fmt.Errorf("could not start consuming: %w", err)
	}

	// Log the setup complete message here
	utils.Logger.Info(fmt.Sprintf(
		"<%s> Reader '%s' - AMQP channel setup complete",
		utils.ERs, rdr.Config().ID))

	// This channel will receive a notification when a channel closed event
	// happens. This must be different from Client.notifyChanClose because the
	// library sends only one notification and Client.notifyChanClose already has
	// a receiver in handleReconnect().
	chClosedCh := make(chan *amqp.Error, 1) // Buffered to avoid deadlocks.
	rdr.client.channel.NotifyClose(chClosedCh)

	go rdr.monitorAndProcess(deliveries, chClosedCh)

	return nil
}

// monitorAndProcess manages the message processing loop for AMQP events.
// It handles reconnection logic in case the AMQP channel closes unexpectedly.
func (rdr *AMQPER) monitorAndProcess(deliveries <-chan amqp.Delivery, chClosedCh chan *amqp.Error) {
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.client.done:
			rdr.close()
			return
		}
	}
	// Initialize a Fibonacci backoff strategy to progressively wait longer
	// between reconnection attempts, avoiding unnecessary load.
	fib := utils.FibDuration(time.Second, rdr.Config().MaxReconnectInterval)

	for {
		select {
		case <-rdr.client.done:
			rdr.close()
			return
		case amqErr := <-chClosedCh:

			// This case handles the event of closed channel (e.g. abnormal shutdown). The if
			// condition is there to make sure it is logged only the first time.
			if amqErr != nil {
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> Reader '%s', AMQP Channel closed due to error: %v",
					utils.ERs, rdr.Config().ID, amqErr))
			}

			// Attempt to re-establish the delivery channel to continue receiving messages.
			var err error
			deliveries, err = rdr.client.Consume()
			if err != nil {

				// If the AMQP channel is not ready, it will continue the loop. Next
				// iteration will enter this case because chClosedCh is closed by the
				// library.
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> Reader %s, failed to init deliveries channel, will retry. Error: %v",
					utils.ERs, rdr.Config().ID, err))

				// Wait for either the backoff duration or a done signal, making sure
				// to gracefully shutdown when the client is done.
				select {
				case <-rdr.client.done:
					rdr.close()
					return
				case <-time.After(fib()):
				}
				continue
			}

			// Successfully reconnected; reset the Fibonacci backoff counter.
			fib = utils.FibDuration(time.Second, rdr.Config().MaxReconnectInterval)

			// Re-set channel to receive notifications.
			// The library closes this channel after abnormal shutdown.
			chClosedCh = make(chan *amqp.Error, 1)
			rdr.client.channel.NotifyClose(chClosedCh)

		case delivery, ok := <-deliveries:
			if !ok {

				// If the deliveries channel is closed, reset it to nil to stop processing
				// until it is reinitialized.
				deliveries = nil
				continue
			}
			go rdr.handleDelivery(delivery)
		}
	}
}

// handleDelivery processes a single message delivery.
func (rdr *AMQPER) handleDelivery(dlv amqp.Delivery) {
	err := rdr.processMessage(dlv.Body)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> Reader %s, processing message %s error: %v",
			utils.ERs, rdr.Config().ID, dlv.MessageId, err))

		err = dlv.Reject(false)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> Reader %s, error negatively acknowledging message %s: %v",
				utils.ERs, rdr.Config().ID, dlv.MessageId, err))
		}
		return
	}

	err = dlv.Ack(false)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> Reader %s, error acknowledging message %s: %v",
			utils.ERs, rdr.Config().ID, dlv.MessageId, err))
	}
}

func (rdr *AMQPER) close() error {
	utils.Logger.Info(fmt.Sprintf(
		"<%s> Reader %s, stop monitoring amqp queue <%s>", utils.ERs, rdr.Config().ID, rdr.Config().SourcePath))
	return rdr.client.Close()
}

// Close will cleanly shut down the channel and connection.
func (client *amqpClient) Close() error {
	if !client.isAvailable() {
		return errors.New("already closed: not connected to the server")
	}

	var err error
	if client.channel != nil {
		if err = client.channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			return fmt.Errorf("failed to close AMQP client channel: %w", err)
		}
	}
	if client.connection != nil {
		if err = client.connection.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			return fmt.Errorf("failed to close AMQP client connection: %w", err)
		}
	}
	client.mu.Lock()
	client.available = false
	client.mu.Unlock()
	return nil
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *amqpClient) handleReconnect(addr, readerID string, maxRetries int,
	maxReconnectInterval time.Duration) {

	fib := utils.FibDuration(time.Second, maxReconnectInterval)
	retryCount := 0

	for retryCount < maxRetries || maxRetries == -1 { // if maxRetries is -1, retry indefinitely
		client.mu.Lock()
		client.available = false
		client.mu.Unlock()

		// Establish an AMQP connection.
		conn, err := amqp.Dial(addr)
		if err != nil {
			reconnectDelay := fib()
			retryCount++
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> Reader %s, failed to connect to AMQP server '%s', will retry. Error: %v",
				utils.ERs, readerID, addr, err))

			select {
			case <-client.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		// Take a new connection to the queue, and update
		// the close listener to reflect this.
		client.connection = conn
		client.notifyConnClose = make(chan *amqp.Error, 1)
		client.connection.NotifyClose(client.notifyConnClose)

		// Reset the fibonacci sequence and retry count after a successful connection.
		fib = utils.FibDuration(time.Second, maxReconnectInterval)
		retryCount = 0

		if done := client.handleReInit(conn, readerID, maxRetries, maxReconnectInterval); done {
			break
		}
	}
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (client *amqpClient) handleReInit(conn *amqp.Connection, readerID string, maxRetries int, maxReInitInterval time.Duration) bool {
	fib := utils.FibDuration(time.Second, maxReInitInterval)
	retryCount := 0
	for retryCount < maxRetries || maxRetries == -1 { // if maxRetries is -1, retry indefinitely
		client.mu.Lock()
		client.available = false
		client.mu.Unlock()
		err := client.init(conn)
		if err != nil {
			reInitDelay := fib()
			retryCount++
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> Reader %s, channel init failed, will retry. Error: %v",
				utils.ERs, readerID, err))

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> Reader %s, connection closed, will attempt reconnect. Error: %v",
					utils.ERs, readerID, err))
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		// Reset the fibonacci sequence and retry count after a successful init.
		fib = utils.FibDuration(time.Second, maxReInitInterval)
		retryCount = 0

		select {
		case <-client.done:
			return true
		case err := <-client.notifyConnClose:
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> Reader %s, connection closed, will atempt to reconnect. Error: %v",
				utils.ERs, readerID, err))
			return false
		case err := <-client.notifyChanClose:
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> Reader %s, channel closed, will attempt re-init. Error: %v",
				utils.ERs, readerID, err))
		}
	}
	return true
}

// init will initialize channel & declare queue
func (client *amqpClient) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a server channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		client.exchange,     // name
		client.exchangeType, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		client.queueID,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		client.queueID,    // queue name
		client.routingKey, // routing key
		client.exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind the queue to the exchange: %w", err)
	}

	// Take a new channel to the queue, and update
	// the channel listener to reflect this.
	client.channel = ch
	client.notifyChanClose = make(chan *amqp.Error, 1)
	client.channel.NotifyClose(client.notifyChanClose)

	client.mu.Lock()
	client.available = true
	client.mu.Unlock()
	return nil
}

// Consume will continuously put queue items on the channel.
func (client *amqpClient) Consume() (<-chan amqp.Delivery, error) {
	if !client.isAvailable() {
		return nil, utils.ErrDisconnected
	}
	if err := client.channel.Qos(
		client.prefetchCount,
		0,
		false,
	); err != nil {
		return nil, err
	}

	return client.channel.Consume(
		client.queueID,
		client.consumer,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (client *amqpClient) isAvailable() bool {
	client.mu.RLock()
	defer client.mu.RUnlock()
	return client.available
}
