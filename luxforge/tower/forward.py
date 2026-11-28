"""Forward LuxRail's service port to the Tower's loopback API.

The Tower serves plain HTTP only on loopback; LuxRail's Caddy terminates TLS for the service's
domain and connects to the container's $PORT over plain HTTP. This copies bytes both ways, so
HTTP, server-sent events and WebSockets pass through unchanged.

Usage: forward.py LISTEN_PORT TARGET_PORT
"""

import asyncio
import contextlib
import sys

CHUNK = 64 * 1024


async def _pipe(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    try:
        while data := await reader.read(CHUNK):
            writer.write(data)
            await writer.drain()
    except (ConnectionError, asyncio.IncompleteReadError):
        pass
    finally:
        with contextlib.suppress(OSError, RuntimeError):
            writer.write_eof()


async def _close(writer: asyncio.StreamWriter) -> None:
    writer.close()
    with contextlib.suppress(OSError):
        await writer.wait_closed()


def _handler(target: int):
    async def handle(client_r: asyncio.StreamReader, client_w: asyncio.StreamWriter) -> None:
        try:
            tower_r, tower_w = await asyncio.wait_for(asyncio.open_connection("127.0.0.1", target), 5)
        except (OSError, TimeoutError):
            # The Tower is starting or stopped: answer like a proxy would, so health checks fail fast.
            client_w.write(b"HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
            with contextlib.suppress(OSError):
                await client_w.drain()
            await _close(client_w)
            return
        try:
            await asyncio.gather(_pipe(client_r, tower_w), _pipe(tower_r, client_w))
        finally:
            await _close(tower_w)
            await _close(client_w)

    return handle


async def main(listen: int, target: int) -> None:
    server = await asyncio.start_server(_handler(target), host="0.0.0.0", port=listen)  # noqa: S104
    print(f"forwarding 0.0.0.0:{listen} to the Tower at 127.0.0.1:{target}", flush=True)
    async with server:
        await server.serve_forever()


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit("usage: forward.py LISTEN_PORT TARGET_PORT")
    with contextlib.suppress(KeyboardInterrupt):
        asyncio.run(main(int(sys.argv[1]), int(sys.argv[2])))
