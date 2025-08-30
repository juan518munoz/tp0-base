import socket
import logging
import signal
import sys
import json

from common.utils import recv_until_newline, Bet, store_bets

class Server:
    def __init__(self, port, listen_backlog):
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
                    self.__handle_client_connection(client_sock)
                except socket.timeout:
                    continue
        finally:
            self._server_socket.close()
            logging.info("action: server_shutdown | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = recv_until_newline(client_sock)
            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')

            try:
                # Parse msg as Bet - TODO: handle possible errors
                bet = Bet.from_json(json.loads(msg))
                # Store bet
                store_bets([bet])
                logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
                client_sock.sendall("{}\n".format("OK").encode('utf-8'))
            except:
                logging.error(f"action: apuesta_almacenada | result: fail | error: invalid_bet | msg: {msg}")
                client_sock.sendall("{}\n".format("FAIL").encode('utf-8'))

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
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
