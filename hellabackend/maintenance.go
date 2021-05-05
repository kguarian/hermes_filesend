package hellabackend

import (
	"errors"
	"net"
	"sync"
	"time"
)

var (
	availableportmutexes chan *PortMutex
	usedportmutexes      chan *PortMutex
	portrange            [2]int16
	portmutexcount       int
)

type PortMutex struct {
	CreationTime    int64 //(time.Time).Unix()
	EarliestTimeout int64
	Portnumber      int16
	Portconn        *net.Conn
	Mutex           *sync.Mutex
	Locked          bool
}

// func InitPortGenerator(lower int, upper int) error {
// 	var mutexes chan *PortMutex
// 	var newMutex PortMutex
// 	if PortBoundCheck(lower, upper) == false {
// 		err := errors.New(ERRMSG_NETWORK_INVALID_PORT)
// 		return err
// 	}
// 	mutexcount := upper - lower + 1
// 	portmutexcount = mutexcount
// 	availableportmutexes = make(chan *PortMutex, mutexcount)
// 	for index := 0; index < mutexcount; index++ {
// 		newMutex = PortMutex{Portnumber: int16(lower + index), Mutex: &sync.Mutex{}, Locked: false}
// 		//watch out for disappearing PortMutexes. Will have to better understand gc if so.
// 		availableportmutexes <- &newMutex
// 	}
// 	availableportmutexes = mutexes
// 	return nil
// }

func GetPortMutex() (*PortMutex, error) {
	var RetPM *PortMutex
	var err error

	if len(availableportmutexes) == 0 {
		err = errors.New(ERRMSG_NETWORK_PORTS_OCCUPIED)
	}
	RetPM = <-availableportmutexes
	RetPM.Mutex.Lock()
	usedportmutexes <- RetPM
	return RetPM, err
}

func PortBoundCheck(lower int, upper int) bool {
	if lower < PORT_LOWER_BOUND || upper > PORT_UPPER_BOUND {
		return false
	} else {
		return true
	}
}

func PortCheck(portnum int) bool {
	if portnum > PORT_LOWER_BOUND && portnum < PORT_UPPER_BOUND {
		return true
	} else {
		return false
	}
}

func NewPortMutex(portnumber int, validitylength int64) (PortMutex, error) {
	var pm PortMutex
	var err error = nil
	var maxvalidity int64 = int64(0) - 1 - time.Now().Unix()
	var addedPortnumber int16
	if validitylength > maxvalidity {
		err = errors.New(ERRMSG_TIME_INVALID)
	}
	if PortCheck(portnumber) {
		addedPortnumber = int16(portnumber)

		pm = PortMutex{
			CreationTime:    time.Now().Unix(),
			EarliestTimeout: validitylength,
			Portnumber:      addedPortnumber,
			Mutex:           new(sync.Mutex),
			Locked:          false,
		}

	} else {
		err = errors.New(ERRMSG_NETWORK_INVALID_PORT)
	}
	return pm, err
}
