import threading


class AgencyQuorum:

    def __init__(self, minimum: int) -> None:
        self._minimum = minimum
        self._condition = threading.Condition()
        self._agencies_done = set()
        self._cancelled = False

    def wait(self, agency_id: int) -> bool:
        with self._condition:
            if self._cancelled:
                return False
            self._agencies_done.add(agency_id)
            self._condition.notify_all()
            while len(self._agencies_done) < self._minimum and not self._cancelled:
                self._condition.wait()
            return not self._cancelled

    def cancel(self) -> None:
        with self._condition:
            self._cancelled = True
            self._condition.notify_all()
