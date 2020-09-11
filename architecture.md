# Multicast to Unicast Project

MulticastGroup: {
IP ipAddress,
int port,
byte[] data,
}

Unicast group: {

}

Multicast Recievers {
byte[] Data
} <!-- -> sends data to unicast sender  -->

unicast Sender{
MulticastGroup self
MulticastGroup[] connectedGroups
}

Multicast Sender {
}

convretedData : {
}

unicast Recievers{
}
