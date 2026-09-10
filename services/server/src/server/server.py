import socket
import signal
import threading
import logger
import safe_socket
import protocol
from .agency_quorum import AgencyQuorum

_STORAGE_PATH = "/tmp/lottery_bets.csv"


class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        from lottery import Lottery

        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path=_STORAGE_PATH)
        self._storage_lock = threading.Lock()
        self._agency_quorum = AgencyQuorum(agency_quorum_min)
        self._shutting_down = threading.Event()
        self._server_socket = None

    def _store_batch(self, agency_id: int, payload: bytes) -> int:
        from lottery import Bet

        bets = [
            Bet(agency_id, first_name, last_name, document, birthdate, number)
            for first_name, last_name, document, birthdate, number in protocol.decode_batch(
                payload
            )
        ]
        with self._storage_lock:
            self.lottery.store_bets(bets)
        return len(bets)

    def _receive_bets(self, client_socket, agency_id: int) -> int:
        bets_amount = 0
        while True:
            payload = safe_socket.recv_frame(client_socket)
            if not payload:
                raise protocol.ProtocolError("received an empty frame")
            if payload[0] == protocol.DONE:
                return bets_amount

            try:
                stored = self._store_batch(agency_id, payload)
            except Exception as e:
                logger.error("handle-batch", logger.LogResult.fail, "err", e)
                safe_socket.send_frame(client_socket, protocol.encode_batch_error())
                continue

            bets_amount += stored
            safe_socket.send_frame(client_socket, protocol.encode_batch_ok())

    def _send_winners(self, client_socket, agency_id: int) -> None:
        with self._storage_lock:
            winners = [
                (bet.first_name, bet.last_name, bet.document, bet.birthdate, bet.number)
                for bet in self.lottery.load_bets()
                if bet.agency_id == agency_id and self.lottery.has_won(bet)
            ]
        safe_socket.send_frame(client_socket, protocol.encode_winners(winners))

    def _handle_client(self, client_socket):
        action = "handle-client"
        logger.info(action, logger.LogResult.in_progress)
        try:
            agency_payload = safe_socket.recv_frame(client_socket)
            agency_id = protocol.decode_agency(agency_payload)

            bets_amount = self._receive_bets(client_socket, agency_id)
            quorum_reached = self._agency_quorum.wait(agency_id)
            if not quorum_reached:
                logger.info(
                    action,
                    logger.LogResult.success,
                    "agency-id",
                    agency_id,
                    "bets-amount",
                    bets_amount,
                    "aborted",
                    "shutting-down",
                )
                return

            self._send_winners(client_socket, agency_id)
            logger.info(
                action,
                logger.LogResult.success,
                "agency-id",
                agency_id,
                "bets-amount",
                bets_amount,
            )
        except Exception as e:
            logger.error(action, logger.LogResult.fail, "err", e)

    def _handle_client_thread(self, client_socket):
        with client_socket:
            self._handle_client(client_socket)

    def _handle_sigterm(self, signum, frame):
        action = "sigterm"
        logger.info(action, logger.LogResult.in_progress)
        self._shutting_down.set()
        self._agency_quorum.cancel()
        if self._server_socket is not None:
            self._server_socket.close()

    def run(self):
        action = "accept-connection"
        signal.signal(signal.SIGTERM, self._handle_sigterm)
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            self._server_socket = server_socket
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    if self._shutting_down.is_set():
                        logger.info(action, logger.LogResult.success, "reason", "sigterm")
                        return
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                threading.Thread(
                    target=self._handle_client_thread,
                    args=(client_socket,),
                    daemon=True,
                ).start()
