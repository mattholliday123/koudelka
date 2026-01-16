import socket
import json

def send_message(query):
# Set the path for the Unix socket
    socket_path = '/tmp/koudelka_socket'
# Create the Unix socket client
    client = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
# Connect to the server
    client.connect(socket_path)
# Send a message to the server
    client.sendall(query.encode())
# Receive a response from the server
    response = client.recv(1024).decode()
    results = json.loads(response)
# Close the connection
    client.close()
    return results
