#!/usr/bin/env python3
"""Keep a pane's command and its children together when the pane closes."""

import os
import signal
import subprocess
import sys


def main():
    if len(sys.argv) < 2:
        sys.exit("usage: zellij-supervise.py command [args...]")

    child = subprocess.Popen(sys.argv[1:], start_new_session=True)

    def stop(signum, _frame):
        # Interrupt the wait; the finally block also handles normal command exit.
        raise SystemExit(128 + signum)

    for sig in (signal.SIGHUP, signal.SIGTERM, signal.SIGINT):
        signal.signal(sig, stop)

    try:
        return child.wait()
    finally:
        # Ignore further signals while the entire process group is shutting down.
        for sig in (signal.SIGHUP, signal.SIGTERM, signal.SIGINT):
            signal.signal(sig, signal.SIG_IGN)
        try:
            os.killpg(child.pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
        try:
            child.wait(timeout=3)
        except subprocess.TimeoutExpired:
            pass
        # The direct child may have exited while a descendant is still alive.
        try:
            os.killpg(child.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        child.wait()


if __name__ == "__main__":
    sys.exit(main())
