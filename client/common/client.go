package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Bet struct that represents a bet
type Bet struct {
	Name      string
	Surname   string
	Id        string
	Birthdate string
	Number    uint
}

// Client Entity that encapsulates how
type Client struct {
	config   ClientConfig
	conn     net.Conn
	shutdown chan struct{}
	bet      Bet
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, bet Bet) *Client {
	client := &Client{
		config:   config,
		shutdown: make(chan struct{}),
		bet:      bet,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// SendBet sends a bet to the server
func (c *Client) SendBet() (string, error) {
	serliazedBet, err := SerializeBet(c.bet)
	if err != nil {
		return "", err
	}

	// Create the connection the server in every loop iteration. Send an
	err = c.createClientSocket()
	if err != nil {
		return "", err
	}

	// Send the bet to the server using flush to avoid short-write
	writer := bufio.NewWriter(c.conn)
	fmt.Fprintln(writer, serliazedBet)
	writer.Flush()

	// Listen for reply
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		return "", err
	}

	return msg, nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Check if shutdown signal	has been received
		select {
		case <-c.shutdown:
			return
		default:
			// continue execution
		}

		// Send the bet to the server
		msg, err := c.SendBet()
		if err != nil {
			log.Warningf("action: send_bet | result: in_progress | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			// Wait some time between retries
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		log.Infof("action: send_bet | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)
		return
	}
	log.Infof("action: send_bet | result: fail | client_id: %v", c.config.ID)
}

func (c *Client) Stop() {
	log.Infof("action: stopping client | result: in_progress | client_id: %v", c.config.ID)

	// Send stop signal to the client loop
	close(c.shutdown)

	if c.conn != nil {
		c.conn.Close()
	}
}
