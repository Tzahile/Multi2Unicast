import socket
import struct

multicastIP = '224.0.0.1'
multicastPort = 11049
IS_ALL_GROUPS = True

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
if IS_ALL_GROUPS:
  # on this port, receives ALL multicast groups
  sock.bind(('', multicastPort))
else:
  # on this port, listen ONLY to multicastIP
  sock.bind((multicastIP, multicastPort))
msgReq = struct.pack("4sl", socket.inet_aton(multicastIP), socket.INADDR_ANY)

sock.setsockopt(socket.IPPROTO_IP, socket.IP_ADD_MEMBERSHIP, msgReq)

while True:
  print(sock.recv(10240))