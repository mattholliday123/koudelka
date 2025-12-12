'''
'Unix sockets for IPC from kpe to backend and vice versa
'''

import socket
import os

socket_path = '/tmp/koudelka_socket'

try:
    os.unlink(socket_path)
except OSError:
    if os.path.exists(socket_path):
        raise

server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

server.bind(socket_path)
server.listen(1)

print('Server is listening for connections')
connection, client_address = server.accept()

try:
    print('Connection from', str(connection).split(", ")[0][-4:])

    while True:
        data = connection.recv(1024)
        if not data:
            break
        print('Received data:', data.decode())
        response = 'Hello from server'
        connection.sendall(response.encode())
finally:
    connection.close()
    os.unlink(socket_path)
