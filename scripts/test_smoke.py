"""Unit tests for the polling harness, without Docker or network calls."""
from http.client import RemoteDisconnected
import unittest
from unittest.mock import patch

import smoke


class ReadinessTests(unittest.TestCase):
    def test_waits_through_connection_resets_during_restart(self):
        with patch.object(smoke, "call", side_effect=[
            ConnectionResetError("restart"), RemoteDisconnected("restart"),
            RuntimeError("HTTP 503"), {"status": "ready"},
        ]) as call, patch.object(smoke.time, "sleep"):
            smoke.ready()
            self.assertEqual(call.call_count, 4)

    def test_does_not_wait_forever(self):
        with patch.object(smoke, "call", side_effect=ConnectionRefusedError), \
                patch.object(smoke.time, "sleep"), \
                patch.object(smoke.time, "monotonic", side_effect=[0, 1, 50]):
            with self.assertRaisesRegex(RuntimeError, "45 seconds"):
                smoke.ready()


if __name__ == "__main__":
    unittest.main()
