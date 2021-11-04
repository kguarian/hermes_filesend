#include "consoleio.h"

extern void DeviceConn(char* userid, char* devicename);

int main(){
    //yeah, I know. How do I suppress the warning, though?
    char *uid = "kguarian";
    char *devname = "cplusplus";
    DeviceConn(uid, devname);
}