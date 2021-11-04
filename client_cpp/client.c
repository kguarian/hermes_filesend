#include<stdlib.h>
#include<stdio.h>
#include<string.h>
#include<sys/socket.h>
#include<arpa/inet.h>
#include "const_net.h"
#include<unistd.h>

//https://www.geeksforgeeks.org/socket-programming-cc/
int main(int argc, char **argv){

    if(argc < 3){
        return 0;
    }

    int socketfd;
    int valread;
    struct sockaddr_in serv_addr;
    char *readbuf;

    if((socketfd = socket(AF_INET, SOCK_STREAM, 0))<0){
        printf("Socket creation error\n");
        return -1;
    }

    serv_addr.sin_family=AF_INET;
    serv_addr.sin_port=htons(PORT);

    if(inet_pton(AF_INET, ipserver, &serv_addr.sin_addr)<1){
        printf("Invalid/unsupported IP address.\n");
        return -1;
    }

    if(connect(socketfd, (struct sockaddr *)&serv_addr, sizeof(serv_addr))<0){
        printf("Connection to server failed.\n");
        return -1;
    }
    send(socketfd, NET_MSG_REQUESTINFO, strlen(NET_MSG_REQUESTINFO),0);
    printf("Message sent: %s\n", NET_MSG_REQUESTINFO);

    send(socketfd, NET_MSG_NEWDEVICE, strlen(NET_MSG_NEWDEVICE),0);
    printf("Message sent: %s\n", NET_MSG_NEWDEVICE);

    readbuf = (char *)malloc(sizeof(char) * 1024);
    valread = read(socketfd, readbuf, sizeof(readbuf));
    if (valread < 0){
        printf("failed read.\n");
        return -1;
    }
    printf("received message: %s\n", readbuf);
}