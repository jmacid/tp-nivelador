import socket
import logger
import safe_socket
import protocol
from lottery import Bet, Lottery

_STORAGE_PATH = "/tmp/lottery_bets.csv"


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path=_STORAGE_PATH)

    def _receive_agency_bets(self, client_socket) -> tuple[int, list[Bet]]:
        agency_payload = safe_socket.recv_frame(client_socket)
        agency_id = protocol.decode_agency(agency_payload)

        bets = []
        while True:
            payload = safe_socket.recv_frame(client_socket)
            if not payload:
                raise protocol.ProtocolError("received an empty frame")
            if payload[0] == protocol.DONE:
                return agency_id, bets

            first_name, last_name, document, birthdate, number = protocol.decode_bet(
                payload
            )
            bets.append(Bet(agency_id, first_name, last_name, document, birthdate, number))

    def _send_winners(self, client_socket, agency_id: int) -> None:
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
            agency_id, bets = self._receive_agency_bets(client_socket)
            self.lottery.store_bets(bets)
            self._send_winners(client_socket, agency_id)
            logger.info(
                action,
                logger.LogResult.success,
                "agency-id",
                agency_id,
                "bets-amount",
                len(bets),
            )
        except Exception as e:
            logger.error(action, logger.LogResult.fail, "err", e)

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                with client_socket:
                    self._handle_client(client_socket)
