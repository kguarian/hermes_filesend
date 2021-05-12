#define PORT 8081

const char *ipserver = "192.168.1.118";
const char *NET_MSG_NEWDEVICE = "nd_c";
const char *NET_MSG_REQUESTINFO = "GET / HTTP/1.1\r\nUser-Agent: hermes-C-client\r\nHost: localhost\r\nAccept-Language: en-us\r\nAccept-Encoding: gzip, deflate\r\nConnection: Keep-Alive\r\n\r\n";