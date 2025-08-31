package common

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
}

// Bet struct that represents a bet
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    uint
}

// Client Entity that encapsulates how
type Client struct {
	config   ClientConfig
	conn     net.Conn
	shutdown chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:   config,
		shutdown: make(chan struct{}),
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
func (c *Client) SendBet(bet Bet) error {
	serializedBet, err := SerializeBet(bet, c.config.ID)
	if err != nil {
		return err
	}

	// Create the connection the server in every loop iteration. Send an
	err = c.createClientSocket()
	if err != nil {
		return err
	}

	// Send the bet to the server using flush to avoid short-write
	writer := bufio.NewWriter(c.conn)
	fmt.Fprintln(writer, serializedBet)
	writer.Flush()

	// Listen for reply
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		return err
	}

	err = ValidateServerResponse(msg)
	if err != nil {
		return err
	}

	return nil
}

// LoadBetsFromCSV loads all bets from the CSV file
func LoadBetsFromCSV(filepath string) ([]Bet, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var bets []Bet
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// CSV format: FirstName,LastName,Document,Birthdate,Number
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid CSV record: %v", record)
		}

		number, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid number in CSV: %s", record[4])
		}

		bet := Bet{
			FirstName: record[0],
			LastName:  record[1],
			Document:  record[2],
			Birthdate: record[3],
			Number:    uint(number),
		}
		bets = append(bets, bet)
	}

	return bets, nil
}

// SendBatchBets sends multiple bets to the server in a single batch
func (c *Client) SendBatchBets(bets []Bet) error {
	if len(bets) == 0 {
		return nil // Nothing to send
	}

	serializedBatch, err := SerializeBatchBets(bets, c.config.ID)
	if err != nil {
		return err
	}

	// Create connection to the server
	err = c.createClientSocket()
	if err != nil {
		return err
	}

	// Send the batch to the server using flush to avoid short-write
	writer := bufio.NewWriter(c.conn)
	_, err = writer.WriteString(serializedBatch)
	if err != nil {
		c.conn.Close()
		return err
	}
	writer.Flush()

	// Listen for reply
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		return err
	}

	err = ValidateServerResponse(msg)
	if err != nil {
		return err
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	bets, err := LoadBetsFromCSV("/data/agency.csv")
	log.Infof("loaded %d bets from CSV", len(bets))
	if err != nil {
		// TODO: log error
		return
	}

	// Process bets in batches
	batchSize := c.config.BatchMaxAmount // TODO: Modify to be lower than 8KB
	for i := 0; i < len(bets); i += batchSize {
		// Check if shutdown signal has been received
		select {
		case <-c.shutdown:
			return
		default:
			// continue execution
		}

		// Calculate end index for current batch
		end := i + batchSize
		if end > len(bets) {
			end = len(bets)
		}

		// Get current batch
		currentBatch := bets[i:end]

		// Send batch
		err = c.SendBatchBets(currentBatch)
		if err != nil {
			// TODO: log error for batch failure
			continue
		}

		// Log success for each bet in batch
		for _, bet := range currentBatch {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				bet.Document,
				bet.Number,
			)
		}
	}

	// TODO: remove log, not complaiant with requirements
	log.Infof("action: stopping client | result: completed | client_id: %v", c.config.ID)
}

func (c *Client) Stop() {
	log.Infof("action: stopping client | result: in_progress | client_id: %v", c.config.ID)

	// Send stop signal to the client loop
	close(c.shutdown)

	if c.conn != nil {
		c.conn.Close()
	}
}
