import socket
import sys


def main():
    if len(sys.argv) != 3 or sys.argv[1] != "connect":
        print("Usage: creek connect <host>:<port>")
        sys.exit(1)

    try:
        host, port = sys.argv[2].split(":")
        port = int(port)
    except ValueError:
        print("Invalid address. Use format <host>:<port>")
        sys.exit(1)

    # Establish TCP connection
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((host, port))
    except Exception as e:
        print(f"Connection failed: {e}")
        sys.exit(1)

    try:
        while True:
            # Wait for the response **before taking new input**
            try:
                response = sock.recv(1024).decode().strip()
                print(f"{response}")

            except (socket.error, ConnectionResetError, BrokenPipeError):
                print("\n[Connection lost]")
                break  # Exit program on connection loss

            # Get user input after receiving response
            message = input("> ")
            if message.lower() in {"exit", "quit"}:
                print("Closing connection...")
                sock.close()
                break
            message += "\n"  # add delimiter at the end
            sock.sendall(message.encode())  # Send message to the server

    except KeyboardInterrupt:
        print("\n[Interrupted, closing connection...]")
    except Exception as e:
        print(f"\n[Error]: {e}")
    finally:
        sock.close()
        sys.exit(1)


if __name__ == "__main__":
    main()
