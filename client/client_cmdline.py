import sys
import socket

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
        print(f"Connected to Creek Server {host}:{port}")
    except Exception as e:
        print(f"Connection failed: {e}")
        sys.exit(1)

    try:
        while True:
            # Get user input
            message = input("> ")
            if message.lower() in {"exit", "quit"}:
                print("Closing connection...")
                sock.close()
                break

            sock.sendall(message.encode())  # Send message to the server

            # Wait for the response **before taking new input**
            response = sock.recv(1024).decode().strip()
            if not response:
                print("\n[Server closed connection]")
                break

            print(f"{response}")  # Display server response

    except KeyboardInterrupt:
        print("\n[Interrupted, closing connection...]")
        sock.close()
    except Exception as e:
        print(f"\n[Error]: {e}")
        sock.close()

if __name__ == "__main__":
    main()
