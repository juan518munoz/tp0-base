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

// ReadBetBatch reads a batch of bets from the CSV file
// It returns the bets read, whether this is the end of file, and any error
func ReadBetBatch(reader *csv.Reader, batchSize int) ([]Bet, bool, error) {
	var bets []Bet
	isEOF := false

	for i := 0; i < batchSize; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			isEOF = true
			break
		}
		if err != nil {
			return nil, false, err
		}

		// CSV format: FirstName,LastName,Document,Birthdate,Number
		if len(record) < 5 {
			return nil, false, fmt.Errorf("invalid CSV record: %v", record)
		}

		number, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			return nil, false, fmt.Errorf("invalid number in CSV: %s", record[4])
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

	return bets, isEOF, nil
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
		return err
	}
	writer.Flush()

	// Listen for reply
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return err
	}
	c.conn.Close()

	err = ValidateBetsServerResponse(msg)
	if err != nil {
		return err
	}

	return nil
}

func SendBets(c *Client) error {
	// Open the CSV file
	file, err := os.Open("/data/agency.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	batchSize := c.config.BatchMaxAmount

	// Read & send bets in batches
	for {
		// Check if shutdown signal has been received
		select {
		case <-c.shutdown:
			return nil
		default:
			// continue execution
		}

		// Read a batch of bets from the CSV file
		bets, isEOF, err := ReadBetBatch(reader, batchSize)
		if err != nil {
			return err
		}

		if len(bets) > 0 {
			err = c.SendBatchBets(bets)
			if err != nil {
				return err
			}
			log.Info("action: apuesta_enviada | result: success | cantidad: ", len(bets))
		}

		// If we've reached the end of the file, we're done
		if isEOF {
			break
		}
	}

	return nil
}

func SendFinishedNotification(c *Client) error {
	err := c.createClientSocket()
	if err != nil {
		return err
	}

	// Send the finished notification to the server using flush to avoid short-write
	writer := bufio.NewWriter(c.conn)
	fmt.Fprintf(writer, "FINISHED,%s", c.config.ID)
	writer.WriteByte(0) // Null byte to indicate end of message
	writer.Flush()

	// Listen for reply
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()
	if err != nil {
		return err
	}

	err = ValidateFinishedNotificationServerResponse(msg)
	if err != nil {
		return err
	}

	log.Info("action: notificacion_finalizacion | result: success")
	return nil
}

func GetLotteryWinnners(c *Client) error {
	for {
		log.Info("action: consulta_ganadores | result: in_progress")

		err := c.createClientSocket()
		if err != nil {
			return err
		}

		// Check if shutdown signal has been received
		select {
		case <-c.shutdown:
			return nil
		default:
			// continue execution
		}

		writer := bufio.NewWriter(c.conn)
		fmt.Fprintf(writer, "RESULTS,%s", c.config.ID)
		writer.WriteByte(0) // Null byte to indicate end of message
		writer.Flush()

		// Listen for reply
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()
		if err != nil {
			return err
		}

		ready, results, err := GetResultsServerResponse(msg)
		if err != nil {
			return err
		}

		if ready {
			log.Info("action: consulta_ganadores | result: success | cant_ganadores: ", results)
			break
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() error {
	err := SendBets(c)
	if err != nil {
		return err
	}

	err = SendFinishedNotification(c)
	if err != nil {
		return err
	}

	err = GetLotteryWinnners(c)
	if err != nil {
		return err
	}

	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) Stop() {
	log.Infof("action: stopping client | result: in_progress | client_id: %v", c.config.ID)

	// Send stop signal to the client loop
	close(c.shutdown)

	if c.conn != nil {
		c.conn.Close()
	}
}
