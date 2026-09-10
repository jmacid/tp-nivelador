import threading


class AgencyQuorum:

    def __init__(self, minimum: int) -> None:
        self._minimum = minimum
        self._condition = threading.Condition()
        self._agencies_done = set()

    def wait(self, agency_id: int) -> None:
        with self._condition:
            self._agencies_done.add(agency_id)
            self._condition.notify_all()
            while len(self._agencies_done) < self._minimum:
                self._condition.wait()
