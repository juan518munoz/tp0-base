import socket
import logging
import signal
import multiprocessing

from common.utils import has_won, load_bets, parse_batch_bets, recv_until_null, store_bets

AGENCY_COUNT = 5

class Server:
    def __init__(self, port, listen_backlog, agency_count):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Set timeout to allow graceful shutdown
        self._server_socket.settimeout(1.0)
        # Setup bellow is used to control server loop
        self._running = True
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
        self._agency_count = agency_count
        manager = multiprocessing.Manager()
        self._finished_agencies = manager.list()
        self._lottery_done = manager.Value('b', False) # boolean value
        self._storage_lock = manager.Lock()

    def _signal_handler(self, signum, frame):
        logging.info(f"action: received_signal | result: in_progress | signal: {signum}")
        self._running = False

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        try:
            while self._running:
                try:
                    client_sock = self.__accept_new_connection()
                    p = multiprocessing.Process(
                        target=self.__handle_client_connection,
                        args=(client_sock,)
                    )
                    p.daemon = True # Terminate if the main process ends
                    p.start()

                    client_sock.close() # Close the main process copy of the socket
                except socket.timeout:
                    continue
        finally:
            self._server_socket.close()
            logging.info("action: server_shutdown | result: success")

    def __send_client_bets_success_message(self, client_sock):
        """
        Send a success message to the client socket using CSV format
        """
        response = "OK"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __send_client_bets_fail_message(self, client_sock, error_msg):
        """
        Send a fail message to the client socket using CSV format
        """
        response = f"FAIL,{error_msg}"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __handle_client_bets_message(self, client_sock, msg):
        """
        Handle a BETS message from a client
        This function tries to parse the bets received from the client.
        If parsing is successful, the bets are stored and a success message
        is sent back to the client. If parsing fails, an error message is sent
        back to the client.
        """
        bets = []
        try:
            # Parse msg as Bet using CSV format
            bets = parse_batch_bets(msg)
            with self._storage_lock:
                store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

            # Notify client that bet was stored successfully
            self.__send_client_bets_success_message(client_sock)
        except:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            # Notify client that at least one of the bets submitted was invalid
            self.__send_client_bets_fail_message(client_sock, "invalid_bets_batch")

    def __send_client_finished_success_message(self, client_sock):
        """
        Send a success message to the client socket using CSV format
        """
        response = "OK"
        client_sock.sendall((response + "\n").encode('utf-8'))
    
    def __send_client_finished_fail_message(self, client_sock, error_msg):
        """
        Send a fail message to the client socket using CSV format
        """
        response = f"FAIL,{error_msg}"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __handle_client_finished_message(self, client_sock, msg):
        """
        Handle a FINISHED message from a client
        This function adds the agency number to the list of finished
        agencies if it is not already present.
        """
        msg_lines = msg.splitlines()
        msg_header = msg_lines[0]
        msg_header_parts = msg_header.split(',')
        if len(msg_header_parts) != 2:
            logging.error("action: agencia_finalizada | result: fail | error: invalid_format")
            self.__send_client_finished_fail_message(client_sock, "invalid_format")
            return

        agency_number = msg_header_parts[1]
        if agency_number not in self._finished_agencies:
            self._finished_agencies.append(agency_number)
        self.__send_client_finished_success_message(client_sock)

    def __send_client_results_not_ready_message(self, client_sock):
        """
        Send a not ready message to the client socket using CSV format
        """
        response = "FAIL,NOT_READY"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __send_client_results_input_not_valid_message(self, client_sock):
        """
        Send a not valid message to the client socket using CSV format
        """
        response = "FAIL,INVALID_INPUT"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __send_client_results_success_message(self, client_sock, won_bets_count):
        """
        Send a success message to the client socket using CSV format
        """
        response = f"OK,{won_bets_count}"
        client_sock.sendall((response + "\n").encode('utf-8'))

    def __handle_client_results_message(self, client_sock, msg):
        """
        Handle a RESULTS message from a client
        This function checks if all agencies have finished sending bets.
        If not, it logs an error. If all agencies have finished, it counts
        the number of winning bets for the specified agency and sends the
        result back to the client.
        """
        if len(self._finished_agencies) < int(self._agency_count):
            logging.error("action: consulta_ganadores | result: in_progress | error: agencias_no_finalizadas")
            self.__send_client_results_not_ready_message(client_sock)
            return

        if not self._lottery_done:
            self._lottery_done = True
            logging.info("action: sorteo | result: success")

        msg_lines = msg.splitlines()
        msg_header = msg_lines[0]
        msg_header_parts = msg_header.split(',')
        if len(msg_header_parts) != 2:
            logging.error("action: consulta_ganadores | result: fail | error: formato_invalido")
            self.__send_client_results_input_not_valid_message(client_sock)
            return

        agency_number = int(msg_header_parts[1])
        with self._storage_lock:
            won_bets_count = sum(
                1 for bet in load_bets()
                if bet.agency == agency_number and has_won(bet)
            )

        self.__send_client_results_success_message(client_sock, won_bets_count)

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            agency_msg = recv_until_null(client_sock)
            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | agency_msg: {agency_msg}')

            # Read header of message to determine message type
            msg_header = agency_msg.splitlines()[0]
            msg_header_parts = msg_header.split(',')

            msg_type = msg_header_parts[0]
            if msg_type == "BETS":
                self.__handle_client_bets_message(client_sock, agency_msg)
            elif msg_type == "FINISHED":
                self.__handle_client_finished_message(client_sock, agency_msg)
            elif msg_type == "RESULTS":
                self.__handle_client_results_message(client_sock, agency_msg)
            else:
                raise ValueError("Invalid message type")

        except Exception as e:
            logging.error("action: handle_message | result: fail | error: {}".format(e))
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
